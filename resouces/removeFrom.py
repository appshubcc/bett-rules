import argparse
import os
import re


def load_removal_sets(file_to_remove):
    exact_remove = set()
    suffix_remove = set()
    if os.path.exists(file_to_remove):
        with open(file_to_remove, "r", encoding="utf-8", errors="ignore") as f_remove:
            for line in f_remove:
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


def is_domain_matched(dom, exact_remove, suffix_remove):
    d = dom.lower()
    if d in exact_remove or d in suffix_remove:
        return True
    parts = d.split(".")
    for i in range(1, len(parts) - 1):
        if ".".join(parts[i:]) in suffix_remove:
            return True
    return False


def extract_domain(raw_line):
    clean = raw_line.split("#")[0].strip()
    if not clean or clean.startswith("include:"):
        return ""
    dom = clean.split(" @")[0].split(":@")[0].strip()
    for p in ["full:", "domain:", "regexp:", "keyword:", "+."]:
        if dom.startswith(p):
            dom = dom[len(p):]
            break
    return dom.strip()


def remove_domains(file_to_remove, file_to_remove_from, output_file):
    exact_remove, suffix_remove = load_removal_sets(file_to_remove)

    with open(file_to_remove_from, "r", encoding="utf-8", errors="ignore") as f_from:
        kept_lines = []
        for line in f_from:
            raw = line.strip()
            if not raw:
                continue
            dom = extract_domain(raw)
            if dom and is_domain_matched(dom, exact_remove, suffix_remove):
                continue
            kept_lines.append(raw)

    out_dir = os.path.dirname(os.path.abspath(output_file))
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)

    with open(output_file, "w", encoding="utf-8") as output:
        output.write("\n".join(kept_lines) + "\n")


def clean_ads_directory(file_to_remove, data_dir):
    """Removes whitelisted domains from category-ads/category-ads-all and strips @ads attributes across data_dir."""
    exact_remove, suffix_remove = load_removal_sets(file_to_remove)
    if not exact_remove and not suffix_remove:
        return

    ads_files = {"category-ads", "category-ads-all"}
    ads_attr_re = re.compile(r"\s+@ads\b")

    for fname in os.listdir(data_dir):
        fpath = os.path.join(data_dir, fname)
        if not os.path.isfile(fpath):
            continue
        with open(fpath, "r", encoding="utf-8", errors="ignore") as f:
            lines = f.readlines()

        modified = False
        new_lines = []
        is_ads_category = fname.lower() in ads_files

        for line in lines:
            raw = line.rstrip("\r\n")
            dom = extract_domain(raw)
            if dom and is_domain_matched(dom, exact_remove, suffix_remove):
                if is_ads_category:
                    modified = True
                    continue
                if "@ads" in raw:
                    raw = ads_attr_re.sub("", raw)
                    modified = True
            new_lines.append(raw)

        if modified:
            with open(fpath, "w", encoding="utf-8") as f:
                f.write("\n".join(new_lines) + "\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Remove domains from a file or clean @ads attributes.")
    parser.add_argument(
        "-remove", required=True, help="File containing domains to be removed"
    )
    parser.add_argument(
        "-from", required=False, dest="from_file", help="File to remove domains from"
    )
    parser.add_argument("-out", required=False, help="Output file")
    parser.add_argument(
        "--clean-ads-dir", required=False, dest="clean_ads_dir",
        help="Directory of community data files to purge whitelisted ad domains and @ads attributes"
    )

    args = parser.parse_args()

    if args.clean_ads_dir:
        clean_ads_directory(args.remove, args.clean_ads_dir)
    if args.from_file and args.out:
        remove_domains(args.remove, args.from_file, args.out)
