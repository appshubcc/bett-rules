import os
import re
import sys


def convert_for_geosite(input_path, output_path):
    """Formats fakeip-filter rules into valid V2Ray GeoSite syntax.

    Filters out '*' and space-containing lines (e.g. 'Mijia Cloud') to avoid
    crashing Loyalsoldier/domain-list-custom attribute parser.
    """
    with open(input_path, "r", encoding="utf-8") as f:
        lines = f.readlines()

    out = []
    seen = set()
    for line in lines:
        line = line.split("#")[0].strip()
        if not line or line == "*" or " " in line:
            continue

        if line.startswith("*."):
            domain = line[2:]
            if "*" in domain:
                pattern = "^" + re.escape(line).replace(r"\*", ".*") + "$"
                entry = "regexp:" + pattern
            else:
                entry = domain
        elif line.startswith("+."):
            domain = line[2:]
            if "*" in domain:
                pattern = "^(.*\\.)?" + re.escape(domain).replace(r"\*", ".*") + "$"
                entry = "regexp:" + pattern
            else:
                entry = domain
        elif "*" in line:
            pattern = "^" + re.escape(line).replace(r"\*", ".*") + "$"
            entry = "regexp:" + pattern
        else:
            entry = "full:" + line

        if entry not in seen:
            seen.add(entry)
            out.append(entry)

    out_dir = os.path.dirname(os.path.abspath(output_path))
    if out_dir:
        os.makedirs(out_dir, exist_ok=True)

    with open(output_path, "w", encoding="utf-8") as f:
        f.write("\n".join(out) + "\n")

    print(f"Converted {len(out)} entries for GeoSite: {input_path} -> {output_path}")


def generate_ruleset(input_path, out_dir):
    """Extracts all 121 non-comment rules verbatim from fakeip-filter.txt.

    Writes:
      - fakeip-filter.list (exactly 121 lines)
      - fakeip-filter.yaml (payload with 121 items)
    """
    with open(input_path, "r", encoding="utf-8") as f:
        rules = [line.strip() for line in f if line.strip() and not line.strip().startswith("#")]

    os.makedirs(out_dir, exist_ok=True)

    list_path = os.path.join(out_dir, "fakeip-filter.list")
    with open(list_path, "w", encoding="utf-8") as f:
        f.write("\n".join(rules) + "\n")

    yaml_path = os.path.join(out_dir, "fakeip-filter.yaml")
    with open(yaml_path, "w", encoding="utf-8") as f:
        f.write("payload:\n")
        for r in rules:
            f.write(f"  - '{r}'\n")

    print(f"Generated standalone ruleset ({len(rules)} rules) in {out_dir}: {list_path}, {yaml_path}")


def main():
    if len(sys.argv) < 3:
        print("Usage:")
        print("  python formatFakeIP.py <input_path> <output_path>")
        print("  python formatFakeIP.py --gen-ruleset <input_path> <out_dir>")
        sys.exit(1)

    if sys.argv[1] == "--gen-ruleset":
        generate_ruleset(sys.argv[2], sys.argv[3])
    else:
        convert_for_geosite(sys.argv[1], sys.argv[2])


if __name__ == "__main__":
    main()
