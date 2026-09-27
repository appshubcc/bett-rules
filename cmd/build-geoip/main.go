package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/inserter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
	"github.com/metacubex/geo/encoding/v2raygeo"
	"go4.org/netipx"
	"google.golang.org/protobuf/proto"
)

var privateCIDRs = []string{
	"10.0.0.0/8",
	"100.64.0.0/10",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.0.0.0/24",
	"192.0.2.0/24",
	"192.168.0.0/16",
	"198.18.0.0/15",
	"198.51.100.0/24",
	"203.0.113.0/24",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"255.255.255.255/32",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
	"ff00::/8",
	"2001:db8::/32",
	"100::/64",
}

var liteCountries = map[string]bool{
	"CN": true,
}

var fallbackCloudflareCIDRs = []string{
	"173.245.48.0/20", "103.21.244.0/22", "103.22.200.0/22", "103.31.4.0/22",
	"141.101.64.0/18", "108.162.192.0/18", "190.93.240.0/20", "188.114.96.0/20",
	"197.234.240.0/22", "198.41.128.0/17", "162.158.0.0/15", "104.16.0.0/13",
	"104.24.0.0/14", "172.64.0.0/13", "131.0.72.0/22",
	"2400:cb00::/32", "2606:4700::/32", "2803:f800::/32", "2405:b500::/32",
	"2405:8100::/32", "2a06:98c0::/29", "2c0f:f248::/32",
	"1.1.1.0/24", "1.0.0.0/24", "162.159.36.0/24", "162.159.46.0/24",
	"2606:4700:4700::/48", "2606:4700:4701::/48",
}

var supplementAppleCIDRs = []string{
	"63.92.224.0/19",
	"192.35.50.0/24",
	"198.183.17.0/24",
	"205.180.175.0/24",
}

var supplementTelegramCIDRs = []string{
	"194.221.250.50/32",
	"2a0a:f280::/32",
}

var supplementSpotifyCIDRs = []string{
	"35.186.224.0/24",
	"104.154.127.0/24",
	"104.199.241.0/24",
	"35.190.69.0/24",
	"35.190.89.0/24",
	"2a01:280:103::/48",
	"2a01:280:206::/48",
}

var preciseASNCountry = map[uint32]string{
	59930:  "US",
	62014:  "SG",
	62041:  "NL",
	44907:  "NL",
	211157: "NL",
}

var preciseASNs = map[string]map[string]bool{
	"telegram":  {"AS44907": true, "AS62041": true, "AS62014": true, "AS59930": true, "AS211157": true},
	"openai":    {"AS401518": true},
	"twitter":   {"AS13414": true, "AS35995": true, "AS63179": true},
	"steam":     {"AS32590": true},
	"spotify":   {"AS8403": true},
	"netflix":   {"AS2906": true, "AS40027": true, "AS55095": true},
	"facebook":  {"AS32934": true, "AS63293": true, "AS54115": true},
	"apple":     {"AS714": true, "AS6185": true, "AS31128": true, "AS1036": true},
	"tiktok":    {"AS138699": true, "AS396986": true, "AS11983": true},
	"bilibili":  {"AS140633": true},
	"fastly":    {"AS54113": true},
	"akamai":    {"AS20940": true, "AS36183": true, "AS32787": true, "AS16625": true, "AS63949": true, "AS35994": true, "AS24319": true, "AS34164": true, "AS12222": true},
	"microsoft": {"AS8075": true, "AS8069": true, "AS3598": true, "AS8068": true, "AS8070": true, "AS35106": true, "AS12076": true},
	"google":    {"AS15169": true, "AS396982": true, "AS394089": true, "AS45566": true, "AS36384": true, "AS36411": true, "AS36383": true, "AS36040": true, "AS19527": true, "AS43515": true, "AS36492": true},
}

var requireNonCN = map[string]bool{
	"apple":      true,
	"microsoft":  true,
	"akamai":     true,
	"fastly":     true,
	"tiktok":     true,
	"bilibili":   true,
	"cloudflare": true,
	"cloudfront": true,
}

func normalizePrefix(p netip.Prefix) (netip.Prefix, bool) {
	addr := p.Addr()
	if addr.Is4In6() {
		return p, false
	}
	if addr.Is4() {
		if p.Bits() > 24 {
			if norm, err := addr.Prefix(24); err == nil {
				return norm, true
			}
		}
		return p, true
	} else if addr.Is6() {
		if p.Bits() > 44 {
			if norm, err := addr.Prefix(44); err == nil {
				return norm, true
			}
		}
		return p, true
	}
	return p, true
}

func loadDB1(path string, countryFull, countryLite map[string][]netip.Prefix) error {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(bufio.NewReaderSize(f, 4*1024*1024))
	reader.ReuseRecord = true

	count := 0
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(rec) < 2 {
			continue
		}
		cc := strings.ToUpper(strings.TrimSpace(rec[1]))
		if cc == "" || cc == "-" {
			continue
		}
		p, err := netip.ParsePrefix(strings.TrimSpace(rec[0]))
		if err != nil {
			continue
		}
		norm, ok := normalizePrefix(p)
		if !ok {
			continue
		}
		countryFull[cc] = append(countryFull[cc], norm)
		if liteCountries[cc] {
			countryLite[cc] = append(countryLite[cc], norm)
		}
		count++
	}
	fmt.Printf("✅ Loaded %d records from DB1: %s\n", count, path)
	return nil
}

func loadASN(path string, asnMap, asnMapLite map[uint32][]netip.Prefix, asnNameMap map[uint32]string, serviceFull, serviceLite map[string][]netip.Prefix, chinaSet, cfSet, cfFrontSet *netipx.IPSet) error {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	reader := csv.NewReader(bufio.NewReaderSize(f, 4*1024*1024))
	reader.ReuseRecord = true

	count := 0
	for {
		rec, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil || len(rec) < 5 {
			continue
		}
		asnStr := strings.TrimSpace(rec[3])
		if asnStr == "" || asnStr == "-" {
			continue
		}
		num, err := strconv.ParseUint(asnStr, 10, 32)
		if err != nil || num == 0 {
			continue
		}
		asnNum := uint32(num)
		asName := strings.TrimSpace(rec[4])

		p, err := netip.ParsePrefix(strings.TrimSpace(rec[2]))
		if err != nil {
			continue
		}
		norm, ok := normalizePrefix(p)
		if !ok {
			continue
		}
		p = norm

		asnMap[asnNum] = append(asnMap[asnNum], p)
		if _, ok := asnNameMap[asnNum]; !ok && asName != "" && asName != "-" {
			asnNameMap[asnNum] = asName
		}

		formattedASN := "AS" + asnStr
		isLiteASN := false

		for tag, asns := range preciseASNs {
			if asns[formattedASN] {
				isLiteASN = true
				if requireNonCN[tag] && chinaSet != nil && chinaSet.Contains(p.Addr()) {
					continue
				}
				serviceFull[tag] = append(serviceFull[tag], p)
				serviceLite[tag] = append(serviceLite[tag], p)
			}
		}

		if !isLiteASN {
			if cfSet != nil && cfSet.ContainsPrefix(p) {
				isLiteASN = true
			} else if cfFrontSet != nil && cfFrontSet.ContainsPrefix(p) {
				isLiteASN = true
			}
		}

		if isLiteASN {
			asnMapLite[asnNum] = append(asnMapLite[asnNum], p)
		}
		count++
	}
	fmt.Printf("✅ Loaded %d records from ASN: %s\n", count, path)
	return nil
}

func main() {
	db1V4File := flag.String("db1-v4", "", "Path to IP2LOCATION-LITE-DB1.CIDR.CSV")
	db1V6File := flag.String("db1-v6", "", "Path to IP2LOCATION-LITE-DB1.IPV6.CIDR.CSV")
	asnV4File := flag.String("asn-v4", "", "Path to IP2LOCATION-LITE-ASN.CSV")
	asnV6File := flag.String("asn-v6", "", "Path to IP2LOCATION-LITE-ASN.IPV6.CSV")
	chinaIPFile := flag.String("china-ip", "", "Path to china46.txt")
	outputDir := flag.String("out", "./publish", "Output directory")
	flag.Parse()

	startTime := time.Now()
	fmt.Println("🚀 Starting GeoIP build with IP2Location...")

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		fmt.Printf("❌ Failed to create output directory: %v\n", err)
		os.Exit(1)
	}

	var chinaSet *netipx.IPSet
	var chinaPrefixes []netip.Prefix
	if *chinaIPFile != "" {
		if fChina, err := os.Open(*chinaIPFile); err == nil {
			var builder netipx.IPSetBuilder
			scanner := bufio.NewScanner(fChina)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				if p, err := netip.ParsePrefix(line); err == nil {
					builder.AddPrefix(p)
					chinaPrefixes = append(chinaPrefixes, p)
				}
			}
			fChina.Close()
			chinaSet, _ = builder.IPSet()
			fmt.Printf("✅ Loaded %d authoritative CN prefixes from: %s\n", len(chinaPrefixes), *chinaIPFile)
		}
	}

	cfPrefixes := fetchCloudflareIPs()
	cfFrontPrefixes := fetchCloudFrontIPs()

	var cfBuilder netipx.IPSetBuilder
	for _, p := range cfPrefixes {
		cfBuilder.AddPrefix(p)
	}
	cfSet, _ := cfBuilder.IPSet()

	var cfFrontBuilder netipx.IPSetBuilder
	for _, p := range cfFrontPrefixes {
		cfFrontBuilder.AddPrefix(p)
	}
	cfFrontSet, _ := cfFrontBuilder.IPSet()

	countryFull := make(map[string][]netip.Prefix, 256)
	countryLite := make(map[string][]netip.Prefix, 2)
	serviceFull := make(map[string][]netip.Prefix, 20)
	serviceLite := make(map[string][]netip.Prefix, 20)
	asnMap := make(map[uint32][]netip.Prefix, 60000)
	asnMapLite := make(map[uint32][]netip.Prefix, 1000)
	asnNameMap := make(map[uint32]string, 60000)

	var privatePrefixes []netip.Prefix
	for _, s := range privateCIDRs {
		if p, err := netip.ParsePrefix(s); err == nil {
			privatePrefixes = append(privatePrefixes, p)
		}
	}
	serviceFull["private"] = privatePrefixes
	serviceLite["private"] = privatePrefixes

	serviceFull["cloudflare"] = cfPrefixes
	serviceLite["cloudflare"] = cfPrefixes

	if len(cfFrontPrefixes) > 0 {
		serviceFull["cloudfront"] = cfFrontPrefixes
		serviceLite["cloudfront"] = cfFrontPrefixes
	}

	if err := loadDB1(*db1V4File, countryFull, countryLite); err != nil {
		fmt.Printf("❌ Failed to load DB1 v4: %v\n", err)
		os.Exit(1)
	}
	if err := loadDB1(*db1V6File, countryFull, countryLite); err != nil {
		fmt.Printf("❌ Failed to load DB1 v6: %v\n", err)
		os.Exit(1)
	}

	if err := loadASN(*asnV4File, asnMap, asnMapLite, asnNameMap, serviceFull, serviceLite, chinaSet, cfSet, cfFrontSet); err != nil {
		fmt.Printf("❌ Failed to load ASN v4: %v\n", err)
		os.Exit(1)
	}
	if err := loadASN(*asnV6File, asnMap, asnMapLite, asnNameMap, serviceFull, serviceLite, chinaSet, cfSet, cfFrontSet); err != nil {
		fmt.Printf("❌ Failed to load ASN v6: %v\n", err)
		os.Exit(1)
	}

	if len(countryFull) <= 1 {
		fmt.Printf("❌ Fatal: countryFull only contains %d countries, DB1 data is missing\n", len(countryFull))
		os.Exit(1)
	}

	if chinaSet != nil {
		fmt.Println("🇨🇳 Injecting china46.txt prefixes into CN and purging from overseas...")
		for cc := range countryFull {
			if cc == "CN" {
				continue
			}
			var builder netipx.IPSetBuilder
			for _, p := range countryFull[cc] {
				builder.AddPrefix(p)
			}
			s, _ := builder.IPSet()
			var diff netipx.IPSetBuilder
			for _, p := range s.Prefixes() {
				if !chinaSet.Contains(p.Addr()) {
					diff.AddPrefix(p)
				}
			}
			cleanSet, _ := diff.IPSet()
			countryFull[cc] = cleanSet.Prefixes()
		}

		for _, p := range chinaPrefixes {
			countryFull["CN"] = append(countryFull["CN"], p)
			if liteCountries["CN"] {
				countryLite["CN"] = append(countryLite["CN"], p)
			}
		}
	}

	if cfSet != nil && len(countryFull["CN"]) > 0 {
		var diff netipx.IPSetBuilder
		for _, p := range countryFull["CN"] {
			if !cfSet.ContainsPrefix(p) {
				diff.AddPrefix(p)
			}
		}
		cleanSet, _ := diff.IPSet()
		countryFull["CN"] = cleanSet.Prefixes()
		if liteCountries["CN"] {
			countryLite["CN"] = cleanSet.Prefixes()
		}
	}

	for _, s := range supplementAppleCIDRs {
		if p, err := netip.ParsePrefix(s); err == nil {
			serviceFull["apple"] = append(serviceFull["apple"], p)
			serviceLite["apple"] = append(serviceLite["apple"], p)
		}
	}

	for _, s := range supplementTelegramCIDRs {
		if p, err := netip.ParsePrefix(s); err == nil {
			serviceFull["telegram"] = append(serviceFull["telegram"], p)
			serviceLite["telegram"] = append(serviceLite["telegram"], p)
			countryFull["NL"] = append(countryFull["NL"], p)
		}
	}

	for _, s := range supplementSpotifyCIDRs {
		if p, err := netip.ParsePrefix(s); err == nil {
			serviceFull["spotify"] = append(serviceFull["spotify"], p)
			serviceLite["spotify"] = append(serviceLite["spotify"], p)
		}
	}

	for asnNum, targetCC := range preciseASNCountry {
		prefixes := asnMap[asnNum]
		if len(prefixes) == 0 {
			continue
		}
		var b netipx.IPSetBuilder
		for _, p := range prefixes {
			b.AddPrefix(p)
		}
		targetSet, err := b.IPSet()
		if err != nil {
			continue
		}

		for cc, list := range countryFull {
			if cc == targetCC {
				continue
			}
			var builder netipx.IPSetBuilder
			for _, p := range list {
				builder.AddPrefix(p)
			}
			s, _ := builder.IPSet()
			var diff netipx.IPSetBuilder
			diff.AddSet(s)
			diff.RemoveSet(targetSet)
			cleanSet, _ := diff.IPSet()
			countryFull[cc] = cleanSet.Prefixes()
		}

		countryFull[targetCC] = append(countryFull[targetCC], prefixes...)
		if liteCountries[targetCC] {
			countryLite[targetCC] = append(countryLite[targetCC], prefixes...)
		}
	}

	fmt.Println("🧩 Aggregating & merging CIDR ranges with IPSet...")
	for cc, list := range countryFull {
		countryFull[cc] = mergePrefixes(list)
	}
	for cc, list := range countryLite {
		countryLite[cc] = mergePrefixes(list)
	}
	for tag, list := range serviceFull {
		serviceFull[tag] = mergePrefixes(list)
	}
	for tag, list := range serviceLite {
		serviceLite[tag] = mergePrefixes(list)
	}

	fmt.Println("📦 Building Country MMDB databases (Full & Lite)...")
	mmdbFull, _ := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType:            "GeoLite2-Country",
		RecordSize:              24,
		IPVersion:               6,
		Inserter:                inserter.ReplaceWith,
		IncludeReservedNetworks: true,
	})
	mmdbLite, _ := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType:            "GeoLite2-Country",
		RecordSize:              24,
		IPVersion:               6,
		Inserter:                inserter.ReplaceWith,
		IncludeReservedNetworks: true,
	})

	for cc, prefixes := range countryFull {
		record := mmdbtype.Map{
			"country":            mmdbtype.Map{"iso_code": mmdbtype.String(cc)},
			"registered_country": mmdbtype.Map{"iso_code": mmdbtype.String(cc)},
		}
		for _, p := range prefixes {
			_ = mmdbFull.Insert(prefixToIPNet(p), record)
		}
	}
	if privPrefixes, ok := serviceFull["private"]; ok {
		record := mmdbtype.Map{
			"country":            mmdbtype.Map{"iso_code": mmdbtype.String("PRIVATE")},
			"registered_country": mmdbtype.Map{"iso_code": mmdbtype.String("PRIVATE")},
		}
		for _, p := range privPrefixes {
			_ = mmdbFull.Insert(prefixToIPNet(p), record)
		}
	}
	for cc, prefixes := range countryLite {
		record := mmdbtype.Map{
			"country":            mmdbtype.Map{"iso_code": mmdbtype.String(cc)},
			"registered_country": mmdbtype.Map{"iso_code": mmdbtype.String(cc)},
		}
		for _, p := range prefixes {
			_ = mmdbLite.Insert(prefixToIPNet(p), record)
		}
	}
	if privPrefixes, ok := serviceLite["private"]; ok {
		record := mmdbtype.Map{
			"country":            mmdbtype.Map{"iso_code": mmdbtype.String("PRIVATE")},
			"registered_country": mmdbtype.Map{"iso_code": mmdbtype.String("PRIVATE")},
		}
		for _, p := range privPrefixes {
			_ = mmdbLite.Insert(prefixToIPNet(p), record)
		}
	}

	if f, err := os.Create(filepath.Join(*outputDir, "country.mmdb")); err == nil {
		_, _ = mmdbFull.WriteTo(f)
		f.Close()
	}
	if f, err := os.Create(filepath.Join(*outputDir, "country-lite.mmdb")); err == nil {
		_, _ = mmdbLite.WriteTo(f)
		f.Close()
	}

	fmt.Println("📦 Building GeoLite2-ASN MMDB databases (Full & Lite)...")
	buildASNMMDB := func(outputPath string, amap map[uint32][]netip.Prefix) {
		mmdbASN, _ := mmdbwriter.New(mmdbwriter.Options{
			DatabaseType: "GeoLite2-ASN",
			RecordSize:   24,
			IPVersion:    6,
		})
		var asnKeys []uint32
		for k := range amap {
			asnKeys = append(asnKeys, k)
		}
		sort.Slice(asnKeys, func(i, j int) bool { return asnKeys[i] < asnKeys[j] })

		for _, asnNum := range asnKeys {
			prefixes := mergePrefixes(amap[asnNum])
			asName := asnNameMap[asnNum]
			record := mmdbtype.Map{
				"autonomous_system_number":       mmdbtype.Uint32(asnNum),
				"autonomous_system_organization": mmdbtype.String(asName),
			}
			for _, p := range prefixes {
				_ = mmdbASN.Insert(prefixToIPNet(p), record)
			}
		}
		if f, err := os.Create(outputPath); err == nil {
			_, _ = mmdbASN.WriteTo(f)
			f.Close()
		}
	}

	buildASNMMDB(filepath.Join(*outputDir, "GeoLite2-ASN.mmdb"), asnMap)
	buildASNMMDB(filepath.Join(*outputDir, "GeoLite2-ASN-lite.mmdb"), asnMapLite)

	fmt.Println("📦 Building MetaDB (Mihomo)...")
	buildMetaDB(filepath.Join(*outputDir, "geoip.metadb"), countryFull, serviceFull)
	buildMetaDB(filepath.Join(*outputDir, "geoip-lite.metadb"), countryLite, serviceLite)

	fmt.Println("📦 Building V2Ray geoip.dat (Full & Lite)...")
	geoipDatFull := buildV2RayGeoIPList(countryFull, serviceFull)
	if err := saveProto(filepath.Join(*outputDir, "geoip.dat"), geoipDatFull); err != nil {
		fmt.Printf("❌ Failed to write geoip.dat: %v\n", err)
	}

	geoipDatLite := buildV2RayGeoIPList(countryLite, serviceLite)
	if err := saveProto(filepath.Join(*outputDir, "geoip-lite.dat"), geoipDatLite); err != nil {
		fmt.Printf("❌ Failed to write geoip-lite.dat: %v\n", err)
	}

	fmt.Printf("🎉 All GeoIP & ASN assets successfully built in %v!\n", time.Since(startTime))
}

func buildMetaDB(metaPath string, countryMap, serviceMap map[string][]netip.Prefix) {
	writerMeta, _ := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType:            "Meta-geoip0",
		IPVersion:               6,
		RecordSize:              24,
		Inserter:                inserter.ReplaceWith,
		DisableIPv4Aliasing:     true,
		IncludeReservedNetworks: true,
	})

	var included []netip.Prefix
	codeMap := make(map[netip.Prefix][]string)

	addPrefix := func(p netip.Prefix, code string) {
		if len(codeMap[p]) == 0 {
			included = append(included, p)
		}
		codeMap[p] = append(codeMap[p], code)
	}

	for cc, prefixes := range countryMap {
		code := strings.ToLower(cc)
		for _, p := range prefixes {
			addPrefix(p, code)
		}
	}
	for tag, prefixes := range serviceMap {
		code := strings.ToLower(tag)
		for _, p := range prefixes {
			addPrefix(p, code)
		}
	}

	sort.Slice(included, func(i, j int) bool {
		return included[i].Bits() < included[j].Bits()
	})

	for _, p := range included {
		ipNet := prefixToIPNet(p)
		codes := codeMap[p]

		_, existingRecord := writerMeta.Get(ipNet.IP)

		var newSlice []mmdbtype.DataType
		if s, ok := existingRecord.(mmdbtype.String); ok {
			newSlice = append(newSlice, s)
		} else if sl, ok := existingRecord.(mmdbtype.Slice); ok {
			newSlice = append(newSlice, sl...)
		}

		for _, c := range codes {
			newSlice = append(newSlice, mmdbtype.String(c))
		}

		seen := make(map[string]bool)
		var finalSlice mmdbtype.Slice
		for _, item := range newSlice {
			if str, ok := item.(mmdbtype.String); ok {
				if !seen[string(str)] {
					seen[string(str)] = true
					finalSlice = append(finalSlice, item)
				}
			}
		}

		var record mmdbtype.DataType
		if len(finalSlice) == 1 {
			record = finalSlice[0]
		} else {
			record = finalSlice
		}

		_ = writerMeta.Insert(ipNet, record)
	}

	if f, err := os.Create(metaPath); err == nil {
		_, _ = writerMeta.WriteTo(f)
		f.Close()
	}
}

func fetchCloudflareIPs() []netip.Prefix {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://api.cloudflare.com/client/v4/ips")
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		var res struct {
			Result struct {
				IPv4CIDRs []string `json:"ipv4_cidrs"`
				IPv6CIDRs []string `json:"ipv6_cidrs"`
			} `json:"result"`
			Success bool `json:"success"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&res); err == nil && res.Success {
			var prefixes []netip.Prefix
			for _, s := range append(res.Result.IPv4CIDRs, res.Result.IPv6CIDRs...) {
				if p, err := netip.ParsePrefix(s); err == nil {
					prefixes = append(prefixes, p)
				}
			}
			for _, s := range []string{"1.1.1.0/24", "1.0.0.0/24", "162.159.36.0/24", "162.159.46.0/24", "2606:4700:4700::/48", "2606:4700:4701::/48"} {
				if p, err := netip.ParsePrefix(s); err == nil {
					prefixes = append(prefixes, p)
				}
			}
			if len(prefixes) > 0 {
				fmt.Printf("✅ Fetched %d official Cloudflare CIDRs from API\n", len(prefixes))
				return prefixes
			}
		}
	}
	fmt.Println("⚠️ Using fallback official Cloudflare CIDRs")
	var prefixes []netip.Prefix
	for _, s := range fallbackCloudflareCIDRs {
		if p, err := netip.ParsePrefix(s); err == nil {
			prefixes = append(prefixes, p)
		}
	}
	return prefixes
}

func fetchCloudFrontIPs() []netip.Prefix {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("https://ip-ranges.amazonaws.com/ip-ranges.json")
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	var doc struct {
		Prefixes []struct {
			IPPrefix string `json:"ip_prefix"`
			Region   string `json:"region"`
			Service  string `json:"service"`
		} `json:"prefixes"`
		IPv6Prefixes []struct {
			IPv6Prefix string `json:"ipv6_prefix"`
			Region     string `json:"region"`
			Service    string `json:"service"`
		} `json:"ipv6_prefixes"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil
	}

	var prefixes []netip.Prefix
	for _, item := range doc.Prefixes {
		if item.Service == "CLOUDFRONT" && !strings.HasPrefix(strings.ToLower(item.Region), "cn-") {
			if p, err := netip.ParsePrefix(item.IPPrefix); err == nil {
				prefixes = append(prefixes, p)
			}
		}
	}
	for _, item := range doc.IPv6Prefixes {
		if item.Service == "CLOUDFRONT" && !strings.HasPrefix(strings.ToLower(item.Region), "cn-") {
			if p, err := netip.ParsePrefix(item.IPv6Prefix); err == nil {
				prefixes = append(prefixes, p)
			}
		}
	}
	if len(prefixes) > 0 {
		fmt.Printf("✅ Fetched %d official CloudFront CIDRs (Non-CN) from AWS API\n", len(prefixes))
	}
	return prefixes
}

func mergePrefixes(prefixes []netip.Prefix) []netip.Prefix {
	var builder netipx.IPSetBuilder
	for _, p := range prefixes {
		builder.AddPrefix(p)
	}
	s, err := builder.IPSet()
	if err != nil {
		return prefixes
	}
	return s.Prefixes()
}

func buildV2RayGeoIPList(countryMap map[string][]netip.Prefix, serviceMap map[string][]netip.Prefix) *v2raygeo.GeoIPList {
	list := &v2raygeo.GeoIPList{}

	var cKeys []string
	for k := range countryMap {
		cKeys = append(cKeys, k)
	}
	sort.Strings(cKeys)

	for _, k := range cKeys {
		prefixes := countryMap[k]
		entry := &v2raygeo.GeoIP{CountryCode: k}
		for _, p := range prefixes {
			addr := p.Addr()
			var ipBytes []byte
			if addr.Is4() {
				b4 := addr.As4()
				ipBytes = b4[:]
			} else {
				b16 := addr.As16()
				ipBytes = b16[:]
			}
			entry.Cidr = append(entry.Cidr, &v2raygeo.CIDR{
				Ip:     ipBytes,
				Prefix: uint32(p.Bits()),
			})
		}
		list.Entry = append(list.Entry, entry)
	}

	var sKeys []string
	for k := range serviceMap {
		sKeys = append(sKeys, k)
	}
	sort.Strings(sKeys)

	for _, k := range sKeys {
		prefixes := serviceMap[k]
		entry := &v2raygeo.GeoIP{CountryCode: strings.ToUpper(k)}
		for _, p := range prefixes {
			addr := p.Addr()
			var ipBytes []byte
			if addr.Is4() {
				b4 := addr.As4()
				ipBytes = b4[:]
			} else {
				b16 := addr.As16()
				ipBytes = b16[:]
			}
			entry.Cidr = append(entry.Cidr, &v2raygeo.CIDR{
				Ip:     ipBytes,
				Prefix: uint32(p.Bits()),
			})
		}
		list.Entry = append(list.Entry, entry)
	}

	return list
}

func saveProto(filepath string, list *v2raygeo.GeoIPList) error {
	data, err := proto.Marshal(list)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath, data, 0o644)
}

func prefixToIPNet(p netip.Prefix) *net.IPNet {
	addr := p.Addr()
	if addr.Is4() {
		b := addr.As4()
		ip := net.IP(b[:])
		mask := net.CIDRMask(p.Bits(), 32)
		return &net.IPNet{IP: ip, Mask: mask}
	} else {
		b := addr.As16()
		ip := net.IP(b[:])
		mask := net.CIDRMask(p.Bits(), 128)
		return &net.IPNet{IP: ip, Mask: mask}
	}
}
