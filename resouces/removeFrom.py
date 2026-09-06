import argparse
import os


def remove_domains(file_to_remove, file_to_remove_from, output_file):
    domains_to_remove = set()
    if os.path.exists(file_to_remove):
        with open(file_to_remove, "r", encoding="utf-8", errors="ignore") as f_remove:
            for line in f_remove:
                clean = line.split("#")[0].strip()
                if not clean:
                    continue
                for p in ["full:", "domain:", "regexp:", "keyword:"]:
                    if clean.startswith(p):
                        clean = clean[len(p):]
                        break
                dom = clean.split(" @")[0].strip().lower()
                if dom:
                    domains_to_remove.add(dom)

    with open(file_to_remove_from, "r", encoding="utf-8", errors="ignore") as f_from:
        kept_lines = []
        for line in f_from:
            raw = line.strip()
            if not raw:
                continue
            clean = raw.split("#")[0].strip()
            if not clean:
                kept_lines.append(raw)
                continue
            dom = clean.split(" @")[0].strip()
            for p in ["full:", "domain:", "regexp:", "keyword:"]:
                if dom.startswith(p):
                    dom = dom[len(p):]
                    break
            if dom.lower() in domains_to_remove:
                continue
            kept_lines.append(raw)

    out_dir = os.path.dirname(os.path.abspath(output_file))
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)

    with open(output_file, "w", encoding="utf-8") as output:
        output.write("\n".join(kept_lines) + "\n")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Remove domains from a file.")
    parser.add_argument(
        "-remove", required=True, help="File containing domains to be removed"
    )
    parser.add_argument(
        "-from", required=True, dest="from_file", help="File to remove domains from"
    )
    parser.add_argument("-out", required=True, help="Output file")

    args = parser.parse_args()

    remove_domains(args.remove, args.from_file, args.out)
