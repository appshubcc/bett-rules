package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/metacubex/geo/convert"
	"github.com/metacubex/geo/encoding/v2raygeo"
	"google.golang.org/protobuf/proto"
)

type domainNode struct {
	children map[string]*domainNode
	isLeaf   bool
}

func newDomainNode() *domainNode {
	return &domainNode{children: make(map[string]*domainNode)}
}

func (n *domainNode) insert(domain string) bool {
	parts := strings.Split(strings.ToLower(domain), ".")
	curr := n
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if curr.isLeaf {
			return false
		}
		if curr.children[part] == nil {
			curr.children[part] = newDomainNode()
		}
		curr = curr.children[part]
	}
	if len(curr.children) > 0 {
		curr.children = make(map[string]*domainNode)
	}
	curr.isLeaf = true
	return true
}

type removalSet struct {
	exact  map[string]bool
	suffix map[string]bool
}

func loadRemovalSet(path string) *removalSet {
	rs := &removalSet{
		exact:  make(map[string]bool),
		suffix: make(map[string]bool),
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return rs
	}
	for _, line := range strings.Split(string(b), "\n") {
		clean := strings.SplitN(line, "#", 2)[0]
		clean = strings.TrimSpace(clean)
		if clean == "" {
			continue
		}
		isFull := strings.HasPrefix(clean, "full:")
		for _, p := range []string{"full:", "domain:", "regexp:", "keyword:", "+."} {
			if strings.HasPrefix(clean, p) {
				clean = clean[len(p):]
				break
			}
		}
		clean = strings.TrimSpace(strings.SplitN(clean, " @", 2)[0])
		clean = strings.ToLower(clean)
		if clean == "" {
			continue
		}
		if isFull || !strings.Contains(clean, ".") {
			rs.exact[clean] = true
		} else {
			rs.suffix[clean] = true
		}
	}
	return rs
}

func (rs *removalSet) matches(domain string) bool {
	if rs == nil {
		return false
	}
	d := strings.ToLower(strings.TrimSpace(domain))
	if d == "" {
		return false
	}
	if rs.exact[d] || rs.suffix[d] {
		return true
	}
	parts := strings.Split(d, ".")
	for i := 1; i < len(parts)-1; i++ {
		if rs.suffix[strings.Join(parts[i:], ".")] {
			return true
		}
	}
	return false
}

func fetchURL(url string) ([]byte, error) {
	client := http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d for %s", resp.StatusCode, url)
	}
	return io.ReadAll(resp.Body)
}

func fetchLines(url string) ([]string, error) {
	b, err := fetchURL(url)
	if err != nil {
		return nil, err
	}
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(string(b)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

// parseClashDomainLines parses Clash .yaml / .list lines into v2fly domain syntax,
// strictly filtering out PROCESS-NAME, IP-CIDR, IP-ASN, USER-AGENT, etc.
func parseClashDomainLines(lines []string) []string {
	var out []string
	for _, l := range lines {
		l = strings.SplitN(l, "#", 2)[0]
		l = strings.TrimSpace(l)
		if l == "" || l == "payload:" {
			continue
		}
		l = strings.TrimPrefix(l, "-")
		l = strings.TrimSpace(l)
		l = strings.Trim(l, "'\"")
		if strings.HasPrefix(l, "DOMAIN,") {
			dom := strings.TrimSpace(strings.SplitN(strings.TrimPrefix(l, "DOMAIN,"), ",", 2)[0])
			if dom != "" {
				out = append(out, "full:"+dom)
			}
		} else if strings.HasPrefix(l, "DOMAIN-SUFFIX,") {
			dom := strings.TrimSpace(strings.SplitN(strings.TrimPrefix(l, "DOMAIN-SUFFIX,"), ",", 2)[0])
			if dom != "" {
				out = append(out, dom)
			}
		} else if strings.HasPrefix(l, "DOMAIN-KEYWORD,") {
			kw := strings.TrimSpace(strings.SplitN(strings.TrimPrefix(l, "DOMAIN-KEYWORD,"), ",", 2)[0])
			if kw != "" {
				out = append(out, "keyword:"+kw)
			}
		}
	}
	return out
}

func stripRulePrefixAndAttrs(line string) string {
	s := strings.SplitN(line, "#", 2)[0]
	s = strings.TrimSpace(s)
	for _, p := range []string{"full:", "domain:", "regexp:", "keyword:", "+."} {
		if strings.HasPrefix(s, p) {
			s = s[len(p):]
			break
		}
	}
	if idx := strings.Index(s, ":@"); idx != -1 {
		s = s[:idx]
	}
	if idx := strings.Index(s, "@"); idx != -1 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}

func main() {
	outputDir := flag.String("out", "./publish", "Output directory for compiled assets")
	flag.Parse()

	startTime := time.Now()
	fmt.Println("🚀 Starting Native GeoSite Build Engine (Full & Lite)...")

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Printf("❌ Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	tempDir, err := os.MkdirTemp("", "geosite-build-*")
	if err != nil {
		fmt.Println("Error creating temp dir:", err)
		return
	}
	defer os.RemoveAll(tempDir)

	dataDir := filepath.Join(tempDir, "data")
	_ = os.MkdirAll(dataDir, 0o755)

	fmt.Println("⏳ Fetching v2fly/domain-list-community master archive...")
	tarGzData, err := fetchURL("https://github.com/v2fly/domain-list-community/archive/refs/heads/master.tar.gz")
	if err != nil {
		fmt.Println("❌ Failed to download community archive:", err)
		return
	}

	gzReader, _ := gzip.NewReader(strings.NewReader(string(tarGzData)))
	tarReader := tar.NewReader(gzReader)

	extractedCount := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF || err != nil {
			break
		}

		if strings.Contains(header.Name, "/data/") && !header.FileInfo().IsDir() {
			fileName := filepath.Base(header.Name)
			destPath := filepath.Join(dataDir, fileName)
			outFile, err := os.Create(destPath)
			if err == nil {
				_, _ = io.Copy(outFile, tarReader)
				outFile.Close()
				extractedCount++
			}
		}
	}
	fmt.Printf("✅ Extracted %d community rule files\n", extractedCount)

	directRemoveSet := loadRemovalSet("resouces/direct-need-to-remove.txt")
	proxyRemoveSet := loadRemovalSet("resouces/proxy-need-to-remove.txt")
	rejectRemoveSet := loadRemovalSet("resouces/reject-need-to-remove.txt")
	cnRemoveSet := loadRemovalSet("resouces/cn-need-to-remove.txt")

	fmt.Println("⏳ Fetching dnsmasq-china-list (CN direct)...")
	cnRaw, err := fetchURL("https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/accelerated-domains.china.conf")
	var cnClean []string
	if err == nil {
		tree := newDomainNode()
		addCleanCN := func(dom string) {
			dom = strings.ToLower(strings.TrimSpace(dom))
			if dom == "" || directRemoveSet.matches(dom) {
				return
			}
			if tree.insert(dom) {
				cnClean = append(cnClean, dom)
			}
		}

		if tldCnRaw, err := os.ReadFile("resouces/tld-cn.txt"); err == nil {
			for _, line := range strings.Split(string(tldCnRaw), "\n") {
				line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
				if line != "" {
					addCleanCN(line)
				}
			}
		}
		if directRaw, err := os.ReadFile("resouces/direct.txt"); err == nil {
			for _, line := range strings.Split(string(directRaw), "\n") {
				line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
				if line != "" {
					addCleanCN(line)
				}
			}
		}
		foreignTLDMap := make(map[string]bool)
		if tldNotCnRaw, err := os.ReadFile("resouces/tld-not-cn.txt"); err == nil {
			for _, line := range strings.Split(string(tldNotCnRaw), "\n") {
				line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
				if line != "" {
					foreignTLDMap[strings.ToLower(line)] = true
				}
			}
		}

		var candidates []string
		scanner := bufio.NewScanner(strings.NewReader(string(cnRaw)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "server=/") {
				parts := strings.Split(line, "/")
				if len(parts) >= 2 {
					dom := strings.ToLower(strings.TrimSpace(parts[1]))
					if dom == "" || strings.Contains(dom, "google") || strings.Contains(dom, "manhua") || strings.Contains(dom, "ooklaserver") || strings.Contains(dom, "acg.rip") {
						continue
					}
					if directRemoveSet.matches(dom) {
						continue
					}
					dParts := strings.Split(dom, ".")
					tld := dParts[len(dParts)-1]
					if foreignTLDMap[tld] || tld == "beer" {
						continue
					}
					candidates = append(candidates, dom)
				}
			}
		}
		if geoCnRaw, err := os.ReadFile("resouces/geolocation-cn.txt"); err == nil {
			for _, line := range strings.Split(string(geoCnRaw), "\n") {
				line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
				if line != "" && !directRemoveSet.matches(line) {
					candidates = append(candidates, strings.ToLower(line))
				}
			}
		}
		sort.Slice(candidates, func(i, j int) bool {
			li := strings.Count(candidates[i], ".")
			lj := strings.Count(candidates[j], ".")
			if li != lj {
				return li < lj
			}
			return candidates[i] < candidates[j]
		})
		for _, dom := range candidates {
			if tree.insert(dom) {
				cnClean = append(cnClean, dom)
			}
		}
		_ = os.WriteFile(filepath.Join(dataDir, "cn"), []byte(strings.Join(cnClean, "\n")+"\n"), 0o644)
		fmt.Printf("✅ Injected pure CN list: %d domains\n", len(cnClean))
	}

	gcnRaw, _ := fetchURL("https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/google.china.conf")
	var gcnDomains []string
	if len(gcnRaw) > 0 {
		var gcn []string
		scanner := bufio.NewScanner(strings.NewReader(string(gcnRaw)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "server=/") {
				parts := strings.Split(line, "/")
				if len(parts) >= 2 {
					gcn = append(gcn, "full:"+parts[1])
					gcnDomains = append(gcnDomains, parts[1])
				}
			}
		}
		_ = os.WriteFile(filepath.Join(dataDir, "google-cn"), []byte(strings.Join(gcn, "\n")+"\n"), 0o644)
	}

	acnRaw, _ := fetchURL("https://raw.githubusercontent.com/felixonmars/dnsmasq-china-list/master/apple.china.conf")
	var acnDomains []string
	if len(acnRaw) > 0 {
		var acn []string
		scanner := bufio.NewScanner(strings.NewReader(string(acnRaw)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "server=/") {
				parts := strings.Split(line, "/")
				if len(parts) >= 2 {
					acn = append(acn, "full:"+parts[1])
					acnDomains = append(acnDomains, parts[1])
				}
			}
		}
		_ = os.WriteFile(filepath.Join(dataDir, "apple-cn"), []byte(strings.Join(acn, "\n")+"\n"), 0o644)
	}

	var customRejects []string
	if rejRaw, err := os.ReadFile("resouces/reject.txt"); err == nil {
		tree := newDomainNode()
		for _, line := range strings.Split(string(rejRaw), "\n") {
			line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
			if line != "" && !rejectRemoveSet.matches(stripRulePrefixAndAttrs(line)) {
				if tree.insert(line) {
					customRejects = append(customRejects, line)
				}
			}
		}
		if len(customRejects) > 0 {
			adsPath := filepath.Join(dataDir, "category-ads")
			f, _ := os.OpenFile(adsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			_, _ = f.WriteString("\n" + strings.Join(customRejects, "\n") + "\n")
			f.Close()
			fmt.Printf("✅ Injected %d custom reject rules into category-ads\n", len(customRejects))
		}
	}

	var customProxies []string
	cnTLDMap := make(map[string]bool)
	if tldCnRaw, err := os.ReadFile("resouces/tld-cn.txt"); err == nil {
		for _, line := range strings.Split(string(tldCnRaw), "\n") {
			line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
			if line != "" {
				cnTLDMap[strings.ToLower(line)] = true
			}
		}
	}
	cnTLDMap["cn"] = true

	prxTree := newDomainNode()
	if tldNotCnRaw, err := os.ReadFile("resouces/tld-not-cn.txt"); err == nil {
		for _, line := range strings.Split(string(tldNotCnRaw), "\n") {
			line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
			if line != "" {
				dom := strings.ToLower(line)
				if prxTree.insert(dom) {
					customProxies = append(customProxies, dom)
				}
			}
		}
	}

	prxPath := filepath.Join(dataDir, "geolocation-!cn")
	var proxyCandidates []string
	if existRaw, err := os.ReadFile(prxPath); err == nil {
		proxyCandidates = append(proxyCandidates, strings.Split(string(existRaw), "\n")...)
	}
	if customPrxLines, err := fetchLines("https://raw.githubusercontent.com/Loyalsoldier/domain-list-custom/release/geolocation-!cn.txt"); err == nil {
		for _, l := range customPrxLines {
			if !strings.Contains(l, ":@cn") {
				proxyCandidates = append(proxyCandidates, strings.SplitN(l, ":@", 2)[0])
			}
		}
	}
	proxyCandidates = append(proxyCandidates, gcnDomains...)
	if prxRaw, err := os.ReadFile("resouces/proxy.txt"); err == nil {
		proxyCandidates = append(proxyCandidates, strings.Split(string(prxRaw), "\n")...)
	}
	proxyCandidates = append(proxyCandidates, "ipleak.net", "browserleaks.org")

	var existingProxy []string
	var specialProxy []string
	seenProxy := make(map[string]bool)
	for _, line := range proxyCandidates {
		line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "include:") || strings.HasPrefix(line, "regexp:") || strings.HasPrefix(line, "keyword:") {
			dom := stripRulePrefixAndAttrs(line)
			if proxyRemoveSet.matches(dom) {
				continue
			}
			if !seenProxy[line] {
				seenProxy[line] = true
				specialProxy = append(specialProxy, line)
			}
			continue
		}
		dom := stripRulePrefixAndAttrs(line)
		if dom == "" || proxyRemoveSet.matches(dom) {
			continue
		}
		parts := strings.Split(dom, ".")
		tld := strings.ToLower(parts[len(parts)-1])
		if cnTLDMap[tld] {
			continue
		}
		existingProxy = append(existingProxy, line)
	}

	sort.Slice(existingProxy, func(i, j int) bool {
		di := stripRulePrefixAndAttrs(existingProxy[i])
		dj := stripRulePrefixAndAttrs(existingProxy[j])
		li := strings.Count(di, ".")
		lj := strings.Count(dj, ".")
		if li != lj {
			return li < lj
		}
		fi := strings.HasPrefix(existingProxy[i], "full:")
		fj := strings.HasPrefix(existingProxy[j], "full:")
		if fi != fj {
			return !fi
		}
		return di < dj
	})

	for _, line := range existingProxy {
		dom := stripRulePrefixAndAttrs(line)
		if strings.HasPrefix(line, "full:") {
			if !seenProxy[line] {
				seenProxy[line] = true
				customProxies = append(customProxies, line)
			}
		} else if prxTree.insert(dom) {
			customProxies = append(customProxies, line)
		}
	}
	allProxies := append(customProxies, specialProxy...)
	_ = os.WriteFile(prxPath, []byte(strings.Join(allProxies, "\n")+"\n"), 0o644)
	fmt.Printf("✅ Injected & deduplicated geolocation-!cn: %d rules\n", len(allProxies))

	geoCnPath := filepath.Join(dataDir, "geolocation-cn")
	if geoCnRaw, err := os.ReadFile("resouces/geolocation-cn.txt"); err == nil {
		var cleanGeoCn []string
		for _, line := range strings.Split(string(geoCnRaw), "\n") {
			line = strings.TrimSpace(strings.SplitN(line, "#", 2)[0])
			if line != "" {
				cleanGeoCn = append(cleanGeoCn, line)
			}
		}
		if len(cleanGeoCn) > 0 {
			f, _ := os.OpenFile(geoCnPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			_, _ = f.WriteString("\n" + strings.Join(cleanGeoCn, "\n") + "\n")
			f.Close()
			fmt.Printf("✅ Injected verified domestic endpoints into geolocation-cn & cn\n")
		}
	}

	cleanShadowFile := func(p string) {
		if b, err := os.ReadFile(p); err == nil {
			var kept []string
			for _, line := range strings.Split(string(b), "\n") {
				dom := stripRulePrefixAndAttrs(line)
				if dom != "" && cnRemoveSet.matches(dom) {
					continue
				}
				kept = append(kept, line)
			}
			_ = os.WriteFile(p, []byte(strings.Join(kept, "\n")), 0o644)
		}
	}
	for _, cat := range []string{"geolocation-cn", "huaweicloud", "wangsu", "alibabacloud", "aliyun"} {
		cleanShadowFile(filepath.Join(dataDir, cat))
	}
	fmt.Println("✅ Cleaned shadowing parent domains from community CN categories")

	// Merge SteamCN (@cn) into steam
	if steamCnLines, err := fetchLines("https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/SteamCN/SteamCN.list"); err == nil {
		steamCnDoms := make(map[string]bool)
		for _, l := range parseClashDomainLines(steamCnLines) {
			steamCnDoms[strings.ToLower(stripRulePrefixAndAttrs(l))] = true
		}
		steamPath := filepath.Join(dataDir, "steam")
		if steamRaw, err := os.ReadFile(steamPath); err == nil {
			var updatedSteam []string
			for _, line := range strings.Split(string(steamRaw), "\n") {
				dom := strings.ToLower(stripRulePrefixAndAttrs(line))
				if dom != "" && steamCnDoms[dom] && !strings.Contains(line, "@cn") {
					line = strings.TrimSpace(line) + " @cn"
				}
				updatedSteam = append(updatedSteam, line)
			}
			_ = os.WriteFile(steamPath, []byte(strings.Join(updatedSteam, "\n")), 0o644)
		}
	}

	// Fetch additional community / meta categories matching run.yml
	if biliLines, err := fetchLines("https://raw.githubusercontent.com/xishang0128/rules/main/biliintl.list"); err == nil {
		_ = os.WriteFile(filepath.Join(dataDir, "biliintl"), []byte(strings.Join(biliLines, "\n")+"\n"), 0o644)
	}
	if odLines, err := fetchLines("https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/OneDrive/OneDrive.list"); err == nil {
		parsed := append(parseClashDomainLines(odLines), "sharepoint.cn")
		_ = os.WriteFile(filepath.Join(dataDir, "onedrive"), []byte(strings.Join(parsed, "\n")+"\n"), 0o644)
	}
	if spLines, err := fetchLines("https://raw.githubusercontent.com/xishang0128/rules/main/sharepoint.list"); err == nil {
		_ = os.WriteFile(filepath.Join(dataDir, "sharepoint"), []byte(strings.Join(spLines, "\n")+"\n"), 0o644)
	}
	var trackerRules []string
	if ptLines, err := fetchLines("https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/PrivateTracker/PrivateTracker.list"); err == nil {
		trackerRules = append(trackerRules, parseClashDomainLines(ptLines)...)
	}
	if trkLines, err := fetchLines("https://gitea.com/XIU2/TrackersListCollection/raw/branch/master/all.txt"); err == nil {
		for _, l := range trkLines {
			if idx := strings.Index(l, "://"); idx != -1 {
				host := l[idx+3:]
				host = strings.SplitN(host, "/", 2)[0]
				host = strings.SplitN(host, ":", 2)[0]
				if strings.Contains(host, ".") && !strings.Contains(strings.ToLower(host), "tracker") {
					trackerRules = append(trackerRules, host)
				}
			}
		}
	}
	if len(trackerRules) > 0 {
		_ = os.WriteFile(filepath.Join(dataDir, "tracker"), []byte(strings.Join(trackerRules, "\n")+"\n"), 0o644)
	}
	if dnsLines, err := fetchLines("https://github.com/blackmatrix7/ios_rule_script/raw/master/rule/Clash/BlockHttpDNS/BlockHttpDNS.list"); err == nil {
		_ = os.WriteFile(filepath.Join(dataDir, "httpdns"), []byte(strings.Join(parseClashDomainLines(dnsLines), "\n")+"\n"), 0o644)
	}
	for winTag, winURL := range map[string]string{
		"win-spy":    "https://raw.githubusercontent.com/crazy-max/WindowsSpyBlocker/master/data/hosts/spy.txt",
		"win-update": "https://raw.githubusercontent.com/crazy-max/WindowsSpyBlocker/master/data/hosts/update.txt",
		"win-extra":  "https://raw.githubusercontent.com/crazy-max/WindowsSpyBlocker/master/data/hosts/extra.txt",
	} {
		if wLines, err := fetchLines(winURL); err == nil {
			var hosts []string
			for _, l := range wLines {
				if strings.HasPrefix(l, "0.0.0.0 ") {
					fields := strings.Fields(l)
					if len(fields) >= 2 {
						hosts = append(hosts, fields[1])
					}
				}
			}
			if len(hosts) > 0 {
				_ = os.WriteFile(filepath.Join(dataDir, winTag), []byte(strings.Join(hosts, "\n")+"\n"), 0o644)
			}
		}
	}
	if devFile, err := os.OpenFile(filepath.Join(dataDir, "category-dev"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
		_, _ = devFile.WriteString("\nzed.dev\n")
		devFile.Close()
	}

	var customFakeIP []string
	if fakeRaw, err := os.ReadFile("resouces/fakeip-filter.txt"); err == nil {
		seen := make(map[string]bool)
		for _, line := range strings.Split(string(fakeRaw), "\n") {
			line = strings.SplitN(line, "#", 2)[0]
			line = strings.TrimSpace(line)
			if line == "" || line == "*" || strings.Contains(line, " ") {
				continue
			}
			var entry string
			if strings.HasPrefix(line, "*.") {
				domain := line[2:]
				if strings.Contains(domain, "*") {
					entry = "regexp:" + "^" + regexp.QuoteMeta(line) + "$"
					entry = strings.ReplaceAll(entry, `\*`, ".*")
				} else {
					entry = domain
				}
			} else if strings.HasPrefix(line, "+.") {
				domain := line[2:]
				if strings.Contains(domain, "*") {
					pat := "^(.*\\.)?" + regexp.QuoteMeta(domain) + "$"
					pat = strings.ReplaceAll(pat, `\*`, ".*")
					entry = "regexp:" + pat
				} else {
					entry = domain
				}
			} else if strings.Contains(line, "*") {
				entry = "regexp:" + "^" + regexp.QuoteMeta(line) + "$"
				entry = strings.ReplaceAll(entry, `\*`, ".*")
			} else {
				entry = "full:" + line
			}
			if !seen[entry] {
				seen[entry] = true
				customFakeIP = append(customFakeIP, entry)
			}
		}
		if len(customFakeIP) > 0 {
			fakePath := filepath.Join(dataDir, "fakeip-filter")
			_ = os.WriteFile(fakePath, []byte(strings.Join(customFakeIP, "\n")+"\n"), 0o644)
			fmt.Printf("✅ Injected %d rules into fakeip-filter\n", len(customFakeIP))
		}
	}

	fmt.Println("🧩 Parsing and recursively resolving Full GeoSite entries...")
	files, _ := os.ReadDir(dataDir)

	type rawEntry struct {
		line     string
		isInc    bool
		incName  string
		incAttrs []string
	}

	rawMap := make(map[string][]rawEntry)

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		tagName := strings.ToLower(file.Name())
		filePath := filepath.Join(dataDir, file.Name())

		f, err := os.Open(filePath)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace( strings.SplitN(scanner.Text(), "#", 2)[0] )
			if line == "" {
				continue
			}
			if strings.HasPrefix(line, "include:") {
				incPart := strings.TrimPrefix(line, "include:")
				incParts := strings.Fields(incPart)
				incName := strings.ToLower(incParts[0])
				var incAttrs []string
				for _, tok := range incParts[1:] {
					if strings.HasPrefix(tok, "@") && len(tok) > 1 {
						incAttrs = append(incAttrs, strings.ToLower(tok[1:]))
					}
				}
				rawMap[tagName] = append(rawMap[tagName], rawEntry{
					isInc:    true,
					incName:  incName,
					incAttrs: incAttrs,
				})
			} else {
				rawMap[tagName] = append(rawMap[tagName], rawEntry{
					line: line,
				})
			}
		}
		f.Close()
	}

	parseDomainLine := func(line string) (v2raygeo.Domain_Type, string, []string) {
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			return v2raygeo.Domain_Domain, "", nil
		}
		mainTok := tokens[0]
		var attrs []string
		if idx := strings.Index(mainTok, ":@"); idx != -1 {
			attrPart := mainTok[idx+2:]
			mainTok = mainTok[:idx]
			for _, a := range strings.Split(attrPart, "@") {
				a = strings.TrimSpace(strings.ToLower(a))
				if a != "" {
					attrs = append(attrs, a)
				}
			}
		}
		for _, tok := range tokens[1:] {
			if strings.HasPrefix(tok, "@") && len(tok) > 1 {
				attrs = append(attrs, strings.ToLower(tok[1:]))
			}
		}

		dType := v2raygeo.Domain_Domain
		val := mainTok
		if strings.HasPrefix(mainTok, "full:") {
			dType = v2raygeo.Domain_Full
			val = strings.TrimPrefix(mainTok, "full:")
		} else if strings.HasPrefix(mainTok, "regexp:") {
			dType = v2raygeo.Domain_Regex
			val = strings.TrimPrefix(mainTok, "regexp:")
		} else if strings.HasPrefix(mainTok, "keyword:") {
			dType = v2raygeo.Domain_Plain
			val = strings.TrimPrefix(mainTok, "keyword:")
		} else if strings.HasPrefix(mainTok, "domain:") {
			dType = v2raygeo.Domain_Domain
			val = strings.TrimPrefix(mainTok, "domain:")
		}
		return dType, strings.TrimSpace(val), attrs
	}

	hasAllAttrs := func(lineAttrs, required []string) bool {
		if len(required) == 0 {
			return true
		}
		attrSet := make(map[string]bool, len(lineAttrs))
		for _, a := range lineAttrs {
			attrSet[a] = true
		}
		for _, req := range required {
			if strings.HasPrefix(req, "!") {
				if attrSet[req[1:]] {
					return false
				}
			} else if !attrSet[req] {
				return false
			}
		}
		return true
	}

	type resolvedItem struct {
		dType v2raygeo.Domain_Type
		val   string
		attrs []string
	}

	resolveTag := func(tag string) []resolvedItem {
		var result []resolvedItem
		visited := make(map[string]bool)

		var dfs func(currTag string, filterAttrs []string)
		dfs = func(currTag string, filterAttrs []string) {
			visitKey := currTag + "|" + strings.Join(filterAttrs, ",")
			if visited[visitKey] {
				return
			}
			visited[visitKey] = true

			for _, item := range rawMap[currTag] {
				if item.isInc {
					mergedFilters := append([]string{}, filterAttrs...)
					mergedFilters = append(mergedFilters, item.incAttrs...)
					dfs(item.incName, mergedFilters)
				} else {
					dType, val, attrs := parseDomainLine(item.line)
					if val == "" {
						continue
					}
					// If domain is in rejectRemoveSet, strip @ads attribute so include:xxx @ads won't pull it in
					if rejectRemoveSet.matches(val) && len(attrs) > 0 {
						var filteredAttrs []string
						for _, a := range attrs {
							if a != "ads" {
								filteredAttrs = append(filteredAttrs, a)
							}
						}
						attrs = filteredAttrs
					}
					if !hasAllAttrs(attrs, filterAttrs) {
						continue
					}
					result = append(result, resolvedItem{
						dType: dType,
						val:   val,
						attrs: attrs,
					})
				}
			}
		}
		dfs(tag, nil)
		return result
	}

	var siteListFull []*v2raygeo.GeoSite
	var allTags []string
	for tag := range rawMap {
		allTags = append(allTags, tag)
	}
	sort.Strings(allTags)

	for _, tag := range allTags {
		allItems := resolveTag(tag)
		var rawItems []resolvedItem
		seen := make(map[string]bool)
		lowerTag := strings.ToLower(tag)

		for _, it := range allItems {
			// Punch-through: If tag is geolocation-!cn or proxy, exclude anything in proxyRemoveSet
			if (lowerTag == "geolocation-!cn" || lowerTag == "proxy") && proxyRemoveSet.matches(it.val) {
				continue
			}
			// Shadowing parent exclusion for geolocation-cn
			if lowerTag == "geolocation-cn" && cnRemoveSet.matches(it.val) {
				continue
			}
			// Ad-blocking whitelist exclusion for category-ads & category-ads-all
			if (lowerTag == "category-ads" || lowerTag == "category-ads-all") && rejectRemoveSet.matches(it.val) {
				continue
			}

			key := fmt.Sprintf("%d:%s:%s", it.dType, strings.ToLower(it.val), strings.Join(it.attrs, ","))
			if !seen[key] {
				seen[key] = true
				rawItems = append(rawItems, it)
			}
		}

		// Sort: shorter domain label count first so root domain inserts before child subdomain
		sort.Slice(rawItems, func(i, j int) bool {
			lenI := len(strings.Split(rawItems[i].val, "."))
			lenJ := len(strings.Split(rawItems[j].val, "."))
			if lenI != lenJ {
				return lenI < lenJ
			}
			return rawItems[i].dType < rawItems[j].dType
		})

		tree := newDomainNode()
		var domains []*v2raygeo.Domain
		for _, it := range rawItems {
			var protoAttrs []*v2raygeo.Domain_Attribute
			for _, a := range it.attrs {
				protoAttrs = append(protoAttrs, &v2raygeo.Domain_Attribute{
					Key:        a,
					TypedValue: &v2raygeo.Domain_Attribute_BoolValue{BoolValue: true},
				})
			}
			if it.dType == v2raygeo.Domain_Domain && len(protoAttrs) == 0 {
				if tree.insert(it.val) {
					domains = append(domains, &v2raygeo.Domain{
						Type:  it.dType,
						Value: it.val,
					})
				}
			} else {
				domains = append(domains, &v2raygeo.Domain{
					Type:      it.dType,
					Value:     it.val,
					Attribute: protoAttrs,
				})
			}
		}

		if len(domains) > 0 {
			siteListFull = append(siteListFull, &v2raygeo.GeoSite{
				CountryCode: strings.ToUpper(tag),
				Domain:      domains,
			})
		}
	}

	protoDataFull, _ := proto.Marshal(&v2raygeo.GeoSiteList{Entry: siteListFull})
	datPath := filepath.Join(*outputDir, "geosite.dat")
	_ = os.WriteFile(datPath, protoDataFull, 0o644)

	dbPath := filepath.Join(*outputDir, "geosite.db")
	if f, err := os.Create(dbPath); err == nil {
		if err := convert.V2RaySiteToSing(siteListFull, f); err != nil {
			fmt.Printf("❌ Failed to write geosite.db: %v\n", err)
		}
		f.Close()
	}

	fmt.Println("🧩 Compiling Lite GeoSite entries...")
	var siteListLite []*v2raygeo.GeoSite

	var cnDomainsLite []*v2raygeo.Domain
	for _, dom := range cnClean {
		cnDomainsLite = append(cnDomainsLite, &v2raygeo.Domain{
			Type:  v2raygeo.Domain_Domain,
			Value: dom,
		})
	}
	siteListLite = append(siteListLite, &v2raygeo.GeoSite{
		CountryCode: "CN",
		Domain:      cnDomainsLite,
	})

	liteClashSources := map[string]string{
		"abema":      "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/AbemaTV/AbemaTV.yaml",
		"apple":      "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Apple/Apple_Classical.yaml",
		"applemusic": "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/AppleMusic/AppleMusic.yaml",
		"bilibili":   "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/BiliBili/BiliBili.yaml",
		"bahamut":    "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Bahamut/Bahamut.yaml",
		"cloudflare": "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Cloudflare/Cloudflare.yaml",
		"google":     "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Google/Google.yaml",
		"github":     "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/GitHub/GitHub.yaml",
		"microsoft":  "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Microsoft/Microsoft.yaml",
		"netflix":    "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Netflix/Netflix.yaml",
		"openai":     "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/OpenAI/OpenAI.yaml",
		"onedrive":   "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/OneDrive/OneDrive.yaml",
		"pixiv":      "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Pixiv/Pixiv.yaml",
		"proxy":      "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/ProxyLite/ProxyLite.yaml",
		"spotify":    "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Spotify/Spotify.yaml",
		"telegram":   "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Telegram/Telegram.yaml",
		"twitter":    "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Twitter/Twitter.yaml",
		"tiktok":     "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/TikTok/TikTok.yaml",
		"youtube":    "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/YouTube/YouTube.yaml",
		"proxymedia": "https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/GlobalMedia/GlobalMedia.yaml",
	}

	liteParsedMap := make(map[string][]string)
	for tag, url := range liteClashSources {
		if lines, err := fetchLines(url); err == nil {
			liteParsedMap[tag] = parseClashDomainLines(lines)
		}
	}
	liteParsedMap["openai"] = append(liteParsedMap["openai"], "full:o33249.ingest.sentry.io", "openai.com")
	liteParsedMap["proxy"] = append(liteParsedMap["proxy"], liteParsedMap["google"]...)
	liteParsedMap["proxy"] = append(liteParsedMap["proxy"], liteParsedMap["github"]...)
	liteParsedMap["proxy"] = append(liteParsedMap["proxy"], liteParsedMap["netflix"]...)
	liteParsedMap["proxy"] = append(liteParsedMap["proxy"], "ipleak.net", "browserleaks.org")

	if biliLines, err := fetchLines("https://raw.githubusercontent.com/xishang0128/rules/main/biliintl.list"); err == nil {
		liteParsedMap["biliintl"] = biliLines
	}
	for _, v2Tag := range []string{"ehentai", "private"} {
		if items := resolveTag(v2Tag); len(items) > 0 {
			for _, it := range items {
				switch it.dType {
				case v2raygeo.Domain_Full:
					liteParsedMap[v2Tag] = append(liteParsedMap[v2Tag], "full:"+it.val)
				case v2raygeo.Domain_Plain:
					liteParsedMap[v2Tag] = append(liteParsedMap[v2Tag], "keyword:"+it.val)
				case v2raygeo.Domain_Regex:
					liteParsedMap[v2Tag] = append(liteParsedMap[v2Tag], "regexp:"+it.val)
				default:
					liteParsedMap[v2Tag] = append(liteParsedMap[v2Tag], it.val)
				}
			}
		}
	}

	var liteKeys []string
	for k := range liteParsedMap {
		liteKeys = append(liteKeys, k)
	}
	sort.Strings(liteKeys)

	for _, tag := range liteKeys {
		var domains []*v2raygeo.Domain
		t := newDomainNode()
		seen := make(map[string]bool)
		for _, r := range liteParsedMap[tag] {
			dType, val, _ := parseDomainLine(r)
			if val == "" {
				continue
			}
			key := fmt.Sprintf("%d:%s", dType, strings.ToLower(val))
			if seen[key] {
				continue
			}
			seen[key] = true
			if dType == v2raygeo.Domain_Domain {
				if t.insert(val) {
					domains = append(domains, &v2raygeo.Domain{
						Type:  dType,
						Value: val,
					})
				}
			} else {
				domains = append(domains, &v2raygeo.Domain{
					Type:  dType,
					Value: val,
				})
			}
		}
		if len(domains) > 0 {
			siteListLite = append(siteListLite, &v2raygeo.GeoSite{
				CountryCode: strings.ToUpper(tag),
				Domain:      domains,
			})
		}
	}

	if len(customRejects) > 0 {
		var adsDomainsLite []*v2raygeo.Domain
		for _, d := range customRejects {
			adsDomainsLite = append(adsDomainsLite, &v2raygeo.Domain{
				Type:  v2raygeo.Domain_Domain,
				Value: d,
			})
		}
		siteListLite = append(siteListLite, &v2raygeo.GeoSite{
			CountryCode: "CATEGORY-ADS",
			Domain:      adsDomainsLite,
		})
	}

	if len(customFakeIP) > 0 {
		var fakeDomainsLite []*v2raygeo.Domain
		for _, d := range customFakeIP {
			dType := v2raygeo.Domain_Domain
			val := d
			if strings.HasPrefix(d, "full:") {
				dType = v2raygeo.Domain_Full
				val = strings.TrimPrefix(d, "full:")
			} else if strings.HasPrefix(d, "regexp:") {
				dType = v2raygeo.Domain_Regex
				val = strings.TrimPrefix(d, "regexp:")
			}
			fakeDomainsLite = append(fakeDomainsLite, &v2raygeo.Domain{
				Type:  dType,
				Value: val,
			})
		}
		siteListLite = append(siteListLite, &v2raygeo.GeoSite{
			CountryCode: "FAKEIP-FILTER",
			Domain:      fakeDomainsLite,
		})
	}

	protoDataLite, _ := proto.Marshal(&v2raygeo.GeoSiteList{Entry: siteListLite})
	datLitePath := filepath.Join(*outputDir, "geosite-lite.dat")
	_ = os.WriteFile(datLitePath, protoDataLite, 0o644)

	dbLitePath := filepath.Join(*outputDir, "geosite-lite.db")
	if f, err := os.Create(dbLitePath); err == nil {
		if err := convert.V2RaySiteToSing(siteListLite, f); err != nil {
			fmt.Printf("❌ Failed to write geosite-lite.db: %v\n", err)
		}
		f.Close()
	}

	fmt.Printf("🎉 All GeoSite assets successfully built in %v!\n", time.Since(startTime))
}
