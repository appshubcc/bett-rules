import os
import sys
import argparse

def load_tlds(filepath):
    tlds = set()
    if not filepath or not os.path.exists(filepath):
        return tlds
    with open(filepath, "r", encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            # Handle comments after '#' and prefixes
            tld = line.split("#")[0].strip().replace("domain:", "").lower()
            if tld:
                tlds.add(tld)
    return tlds

def load_domain_set(filepath):
    exact_remove = set()
    suffix_remove = set()
    if not filepath or not os.path.exists(filepath):
        return exact_remove, suffix_remove
    with open(filepath, "r", encoding="utf-8", errors="ignore") as f:
        for line in f:
            clean = line.split("#")[0].strip()
            if not clean:
                continue
            is_full = clean.startswith("full:")
            for p in ["full:", "domain:", "regexp:", "keyword:", "+."]:
                if clean.startswith(p):
                    clean = clean[len(p):]
                    break
            dom = clean.split(" @")[0].strip().lower()
            if not dom:
                continue
            if is_full or "." not in dom:
                exact_remove.add(dom)
            else:
                suffix_remove.add(dom)
    return exact_remove, suffix_remove

def is_domain_removed(domain, exact_remove, suffix_remove):
    d = domain.lower()
    if d in exact_remove or d in suffix_remove:
        return True
    parts = d.split(".")
    for i in range(1, len(parts) - 1):
        if ".".join(parts[i:]) in suffix_remove:
            return True
    return False

def parse_rule(line):
    line = line.split("#")[0].strip()
    if not line:
        return "", "", ""
    prefix = ""
    domain = line
    for p in ["full:", "domain:", "regexp:", "keyword:"]:
        if line.startswith(p):
            prefix = p
            domain = line[len(p):]
            break
    # Attribute handling, e.g. "domain.com @cn"
    attr = ""
    if " @" in domain:
        parts = domain.split(" @", 1)
        domain = parts[0].strip()
        attr = " @" + parts[1].strip()
    return prefix, domain.lower(), attr

class TrieNode:
    def __init__(self):
        self.children = {}
        self.is_leaf = False  # domain suffix match (covers all subdomains)
        self.is_exact = False # exact match only (full:)

def insert_trie_leaf(root, domain):
    curr = root
    for p in reversed(domain.split(".")):
        if curr.is_leaf:
            return False
        if p not in curr.children:
            curr.children[p] = TrieNode()
        curr = curr.children[p]
    curr.is_leaf = True
    curr.children.clear()
    return True

def dedup_cn(input_path, tld_cn_path, output_path, tld_not_cn_path=None, direct_path=None, remove_path=None):
    tld_cn = load_tlds(tld_cn_path)
    tld_not_cn = load_tlds(tld_not_cn_path) if tld_not_cn_path else set()
    direct_roots = load_tlds(direct_path) if direct_path else set()
    exact_remove, suffix_remove = load_domain_set(remove_path) if remove_path else (set(), set())

    with open(input_path, "r", encoding="utf-8", errors="ignore") as f:
        lines = f.readlines()

    root_trie = TrieNode()
    out_roots = []
    seen = set()
    dedup_count = 0
    excluded_foreign_count = 0
    subdomain_absorbed_count = 0

    # Ensure all tld-cn are injected as root rules
    for tld in sorted(tld_cn):
        if not is_domain_removed(tld, exact_remove, suffix_remove):
            seen.add(tld)
            out_roots.append(tld)
            insert_trie_leaf(root_trie, tld)

    # Ensure all direct root domains are injected (if not removed)
    for root in sorted(direct_roots):
        if root not in seen and not is_domain_removed(root, exact_remove, suffix_remove):
            seen.add(root)
            out_roots.append(root)
            insert_trie_leaf(root_trie, root)

    parsed = []
    special_rules = []

    for raw_line in lines:
        prefix, domain, attr = parse_rule(raw_line)
        if not domain or is_domain_removed(domain, exact_remove, suffix_remove):
            continue

        if prefix in ["regexp:", "keyword:"]:
            s = raw_line.strip()
            if s not in seen:
                seen.add(s)
                special_rules.append(s)
            continue

        parts = domain.split(".")
        if len(parts) <= 1:
            continue
        tld = parts[-1]

        # 1. If domain ends with .beer (junk hash domains), exclude it
        if tld == "beer":
            excluded_foreign_count += 1
            continue

        # 2. If domain ends with a CN TLD and is a subdomain of that TLD, absorb it
        if tld in tld_cn and len(parts) > 1:
            dedup_count += 1
            continue

        # 3. If domain ends with a pure foreign TLD, exclude it from CN
        if tld in tld_not_cn:
            excluded_foreign_count += 1
            continue

        final_entry = f"{prefix}{domain}{attr}".strip()
        if final_entry and final_entry not in seen:
            seen.add(final_entry)
            parsed.append((prefix, domain, attr, final_entry))

    def sort_key(item):
        prefix, domain, attr, final_entry = item
        parts = domain.split(".")
        is_full = 1 if prefix == "full:" else 0
        return (len(parts), is_full)

    parsed.sort(key=sort_key)

    retained_rules = []
    for prefix, domain, attr, final_entry in parsed:
        parts = domain.split(".")
        curr = root_trie
        redundant = False

        for p in reversed(parts):
            if curr.is_leaf:
                redundant = True
                break
            if p not in curr.children:
                curr.children[p] = TrieNode()
            curr = curr.children[p]

        if not redundant:
            if curr.is_leaf:
                redundant = True
            elif prefix == "full:" and curr.is_exact:
                redundant = True

        if redundant:
            subdomain_absorbed_count += 1
        else:
            if prefix in ["", "domain:", "+."]:
                curr.is_leaf = True
                curr.children.clear()
            else:
                curr.is_exact = True
            retained_rules.append(final_entry)

    out_domains = out_roots + retained_rules + special_rules

    out_dir = os.path.dirname(os.path.abspath(output_path))
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)

    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(out_domains) + "\n")

    print(f"[CN Deduplication] Input: {len(lines)} lines -> Output: {len(out_domains)} lines (Eliminated {dedup_count} subdomains under tld-cn, Absorbed {subdomain_absorbed_count} redundant subdomains via Trie, Excluded {excluded_foreign_count} foreign/beer domains, Injected {len(tld_cn)} TLDs and {len(direct_roots)} direct roots)")

def dedup_proxy(input_path, tld_not_cn_path, tld_cn_path, output_path, reserve_path=None, remove_path=None, cn_list_path=None):
    tld_not_cn = load_tlds(tld_not_cn_path)
    tld_cn = load_tlds(tld_cn_path)
    tld_cn.add("cn")
    exact_remove, suffix_remove = load_domain_set(remove_path) if remove_path else (set(), set())

    cn_roots = set()
    if cn_list_path and os.path.exists(cn_list_path):
        with open(cn_list_path, "r", encoding="utf-8", errors="ignore") as f:
            for line in f:
                pfx, d, _ = parse_rule(line)
                if d and pfx in ["", "domain:", "+."]:
                    cn_roots.add(d)

    root = TrieNode()
    # Inject all foreign TLD roots into the Trie as leaves
    for tld in tld_not_cn:
        if tld not in root.children:
            root.children[tld] = TrieNode()
        root.children[tld].is_leaf = True

    lines = []
    if os.path.exists(input_path):
        with open(input_path, "r", encoding="utf-8", errors="ignore") as f:
            lines.extend(f.readlines())

    if reserve_path and os.path.exists(reserve_path):
        with open(reserve_path, "r", encoding="utf-8", errors="ignore") as f:
            lines.extend(f.readlines())

    parsed = []
    special_rules = []
    seen = set()
    cn_excluded_count = 0
    removed_count = 0

    for raw_line in lines:
        prefix, domain, attr = parse_rule(raw_line)
        if not domain:
            continue

        if is_domain_removed(domain, exact_remove, suffix_remove):
            removed_count += 1
            continue

        if prefix in ["regexp:", "keyword:"]:
            line_str = raw_line.strip()
            if line_str not in seen:
                seen.add(line_str)
                special_rules.append(line_str)
            continue

        parts = domain.split(".")
        if len(parts) <= 1:
            continue
        tld = parts[-1]

        # 1. Exclude any domain ending with a CN TLD
        if tld in tld_cn:
            cn_excluded_count += 1
            continue

        # 2. Exclude if exact root domain is already present in CN list (mutual exclusivity)
        if cn_roots and domain in cn_roots and prefix in ["", "domain:", "+."]:
            removed_count += 1
            continue

        final_entry = f"{prefix}{domain}{attr}".strip()
        if final_entry not in seen:
            seen.add(final_entry)
            parsed.append((prefix, domain, attr, final_entry))

    # Sort: shortest domain first (so parent domain is processed before child subdomain)
    # Tie-breaker: suffix match before full exact match
    def sort_key(item):
        prefix, domain, attr, final_entry = item
        parts = domain.split(".")
        is_full = 1 if prefix == "full:" else 0
        return (len(parts), is_full)

    parsed.sort(key=sort_key)

    retained_rules = []
    pruned_subdomains_count = 0

    for prefix, domain, attr, final_entry in parsed:
        parts = domain.split(".")
        curr = root
        redundant = False

        # Walk labels backwards: e.g. com -> adobe -> 3dns-1
        for p in reversed(parts):
            if curr.is_leaf:
                redundant = True
                break
            if p not in curr.children:
                curr.children[p] = TrieNode()
            curr = curr.children[p]

        if not redundant:
            if curr.is_leaf:
                redundant = True
            elif prefix == "full:" and curr.is_exact:
                redundant = True

        if redundant:
            pruned_subdomains_count += 1
        else:
            if prefix in ["", "domain:", "+."]:
                curr.is_leaf = True
            else:
                curr.is_exact = True
            retained_rules.append(final_entry)

    # Injected TLD roots
    out_domains = sorted(list(tld_not_cn)) + retained_rules + special_rules

    out_dir = os.path.dirname(os.path.abspath(output_path))
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)

    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(out_domains) + "\n")

    print(f"[Proxy Deduplication] Input: {len(lines)} lines -> Output: {len(out_domains)} lines (Pruned {pruned_subdomains_count} redundant subdomains via Trie, Excluded {cn_excluded_count} CN domains, Filtered {removed_count} excluded domains, Injected {len(tld_not_cn)} TLD roots)")

def main():
    parser = argparse.ArgumentParser(description="TLD-based domain list deduplication tool")
    parser.add_argument("--mode", choices=["cn", "proxy"], required=True, help="Processing mode")
    parser.add_argument("--input", required=True, help="Input domain list path")
    parser.add_argument("--output", required=True, help="Output domain list path")
    parser.add_argument("--tld-cn", default="resouces/tld-cn.txt", help="Path to tld-cn file")
    parser.add_argument("--tld-not-cn", default="resouces/tld-not-cn.txt", help="Path to tld-not-cn file")
    parser.add_argument("--direct", default="resouces/direct.txt", help="Path to direct root domains file")
    parser.add_argument("--reserve", default=None, help="Path to reserved rules file (full, regexp, keyword)")
    parser.add_argument("--remove", default=None, help="Path to domain removal list")
    parser.add_argument("--cn-list", default=None, help="Path to cn list for mutual exclusivity")

    args = parser.parse_args()

    if args.mode == "cn":
        dedup_cn(args.input, args.tld_cn, args.output, args.tld_not_cn, args.direct, args.remove)
    elif args.mode == "proxy":
        dedup_proxy(args.input, args.tld_not_cn, args.tld_cn, args.output, args.reserve, args.remove, args.cn_list)

if __name__ == "__main__":
    main()
