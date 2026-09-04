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

def dedup_cn(input_path, tld_cn_path, output_path, tld_not_cn_path=None, direct_path=None):
    tld_cn = load_tlds(tld_cn_path)
    tld_not_cn = load_tlds(tld_not_cn_path) if tld_not_cn_path else set()
    direct_roots = load_tlds(direct_path) if direct_path else set()

    with open(input_path, "r", encoding="utf-8", errors="ignore") as f:
        lines = f.readlines()

    out_domains = []
    seen = set()
    dedup_count = 0
    excluded_foreign_count = 0
    subdomain_absorbed_count = 0

    # Ensure all tld-cn are injected as root rules
    for tld in sorted(tld_cn):
        seen.add(tld)
        out_domains.append(tld)

    # Ensure all direct root domains are injected
    for root in sorted(direct_roots):
        if root not in seen:
            seen.add(root)
            out_domains.append(root)

    for raw_line in lines:
        prefix, domain, attr = parse_rule(raw_line)
        if not domain:
            continue

        if prefix in ["regexp:", "keyword:"]:
            if raw_line.strip() not in seen:
                seen.add(raw_line.strip())
                out_domains.append(raw_line.strip())
            continue

        parts = domain.split(".")
        tld = parts[-1]

        # 1. If domain ends with .beer (junk hash domains), exclude it
        if tld == "beer":
            excluded_foreign_count += 1
            continue

        # 2. If domain ends with a CN TLD and is a subdomain of that TLD, absorb it
        if tld in tld_cn and len(parts) > 1:
            dedup_count += 1
            continue

        # 3. If domain is a subdomain of an injected direct root domain, absorb it
        absorbed_by_root = False
        for root in direct_roots:
            if domain.endswith("." + root):
                subdomain_absorbed_count += 1
                absorbed_by_root = True
                break
        if absorbed_by_root:
            continue

        # 4. If domain ends with a pure foreign TLD, exclude it from CN
        if tld in tld_not_cn:
            excluded_foreign_count += 1
            continue

        final_entry = f"{prefix}{domain}{attr}".strip()
        if final_entry and final_entry not in seen:
            seen.add(final_entry)
            out_domains.append(final_entry)

    out_dir = os.path.dirname(os.path.abspath(output_path))
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)

    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(out_domains) + "\n")

    print(f"[CN Deduplication] Input: {len(lines)} lines -> Output: {len(out_domains)} lines (Eliminated {dedup_count} subdomains under tld-cn, Absorbed {subdomain_absorbed_count} subdomains under direct roots, Excluded {excluded_foreign_count} foreign/beer domains, Injected {len(tld_cn)} TLDs and {len(direct_roots)} direct roots)")

def dedup_proxy(input_path, tld_not_cn_path, tld_cn_path, output_path):
    tld_not_cn = load_tlds(tld_not_cn_path)
    tld_cn = load_tlds(tld_cn_path)
    # also ensure 'cn' is in tld_cn
    tld_cn.add("cn")

    with open(input_path, "r", encoding="utf-8", errors="ignore") as f:
        lines = f.readlines()

    out_domains = []
    seen = set()
    cn_excluded_count = 0
    tld_absorbed_count = 0

    # Inject all tld-not-cn as root rules
    for tld in sorted(tld_not_cn):
        seen.add(tld)
        out_domains.append(tld)

    for raw_line in lines:
        prefix, domain, attr = parse_rule(raw_line)
        if not domain:
            continue

        if prefix in ["regexp:", "keyword:"]:
            if raw_line.strip() not in seen:
                seen.add(raw_line.strip())
                out_domains.append(raw_line.strip())
            continue

        parts = domain.split(".")
        tld = parts[-1]

        # 1. Exclude any domain ending with a CN TLD
        if tld in tld_cn:
            cn_excluded_count += 1
            continue

        # 2. Absorb any domain ending with a non-CN TLD that is covered by tld-not-cn
        if tld in tld_not_cn and len(parts) > 1:
            tld_absorbed_count += 1
            continue

        final_entry = f"{prefix}{domain}{attr}".strip()
        if final_entry and final_entry not in seen:
            seen.add(final_entry)
            out_domains.append(final_entry)

    out_dir = os.path.dirname(os.path.abspath(output_path))
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)

    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(out_domains) + "\n")

    print(f"[Proxy Deduplication] Input: {len(lines)} lines -> Output: {len(out_domains)} lines (Eliminated {tld_absorbed_count} subdomains under tld-!cn, Excluded {cn_excluded_count} CN domains, Injected {len(tld_not_cn)} TLDs)")

def main():
    parser = argparse.ArgumentParser(description="TLD-based domain list deduplication tool")
    parser.add_argument("--mode", choices=["cn", "proxy"], required=True, help="Processing mode")
    parser.add_argument("--input", required=True, help="Input domain list path")
    parser.add_argument("--output", required=True, help="Output domain list path")
    parser.add_argument("--tld-cn", default="resouces/tld-cn.txt", help="Path to tld-cn file")
    parser.add_argument("--tld-not-cn", default="resouces/tld-not-cn.txt", help="Path to tld-not-cn file")
    parser.add_argument("--direct", default="resouces/direct.txt", help="Path to direct root domains file")

    args = parser.parse_args()

    if args.mode == "cn":
        dedup_cn(args.input, args.tld_cn, args.output, args.tld_not_cn, args.direct)
    elif args.mode == "proxy":
        dedup_proxy(args.input, args.tld_not_cn, args.tld_cn, args.output)

if __name__ == "__main__":
    main()
