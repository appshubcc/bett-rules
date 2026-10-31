# Bett Rules

适用于 **Mihomo (Clash.Meta)**、**sing-box** 与 **V2Ray / Xray** 的分流规则与 GeoIP / GeoSite / ASN 数据库，每日北京时间 **06:30** 自动更新。

---

## 主要特点

- **IP 网段裁剪与合并**：对细碎的 IPv4（`/24`）与 IPv6（`/44`）网段进行规整与相邻 CIDR 聚合合并，在保证分流准确的前提下大幅减小数据库体积。
- **大陆 IP 边界修正**：基于 `IP2Location` 与 [`china46.txt`](https://github.com/gaoyifan/china-operator-ip) 修正 `CN` 网段，并将国内网段从海外国家及跨国 CDN 集合中剔除。
- **域名去重与校准**：通过顶级域（TLD）和父子域名包含关系自动剔除冗余子域名，并结合国内外 DNS 解析结果修正了部分跨国服务与 CDN 的归属问题。
- **提供 Lite 轻量版**：除全量文件外，同步提供仅保留常用分流集合的 `lite` 版本，适合内存受限的路由器或移动端设备。

---

## 下载地址

### 1. 数据库文件

| 文件名 | 说明 | GitHub | JSDMirror | jsDelivr | Cloudflare | Fastly | Gcore |
| :--- | :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **`geoip.metadb`** | Mihomo 全量 GeoIP | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geoip.metadb) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/geoip.metadb) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.metadb) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.metadb) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.metadb) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.metadb) |
| **`geoip-lite.metadb`** | Mihomo 轻量 GeoIP | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geoip-lite.metadb) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/geoip-lite.metadb) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.metadb) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.metadb) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.metadb) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.metadb) |
| **`country.mmdb`** | MaxMind 全量 MMDB | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/country.mmdb) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/country.mmdb) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/country.mmdb) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/country.mmdb) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/country.mmdb) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/country.mmdb) |
| **`country-lite.mmdb`** | MaxMind 轻量 MMDB | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/country-lite.mmdb) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/country-lite.mmdb) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/country-lite.mmdb) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/country-lite.mmdb) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/country-lite.mmdb) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/country-lite.mmdb) |
| **`geoip.dat`** | V2Ray / Mihomo 全量 GeoIP | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geoip.dat) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/geoip.dat) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.dat) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.dat) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.dat) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.dat) |
| **`geoip-lite.dat`** | V2Ray / Mihomo 轻量 GeoIP | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geoip-lite.dat) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/geoip-lite.dat) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.dat) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.dat) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.dat) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.dat) |
| **`geoip.db`** | sing-box 全量 GeoIP | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geoip.db) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/geoip.db) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.db) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.db) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.db) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.db) |
| **`geoip-lite.db`** | sing-box 轻量 GeoIP | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geoip-lite.db) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/geoip-lite.db) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.db) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.db) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.db) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip-lite.db) |
| **`geosite.dat`** | V2Ray / Mihomo 全量 GeoSite | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geosite.dat) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/geosite.dat) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.dat) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.dat) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.dat) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.dat) |
| **`geosite-lite.dat`** | V2Ray / Mihomo 轻量 GeoSite | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geosite-lite.dat) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/geosite-lite.dat) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite-lite.dat) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite-lite.dat) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite-lite.dat) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite-lite.dat) |
| **`geosite.db`** | sing-box 全量 GeoSite | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geosite.db) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/geosite.db) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.db) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.db) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.db) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.db) |
| **`geosite-lite.db`** | sing-box 轻量 GeoSite | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geosite-lite.db) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/geosite-lite.db) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite-lite.db) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite-lite.db) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite-lite.db) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite-lite.db) |
| **`GeoLite2-ASN.mmdb`** | 全量 ASN 数据库 | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/GeoLite2-ASN.mmdb) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/GeoLite2-ASN.mmdb) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN.mmdb) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN.mmdb) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN.mmdb) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN.mmdb) |
| **`GeoLite2-ASN-lite.mmdb`** | 轻量 ASN 数据库 | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/GeoLite2-ASN-lite.mmdb) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/GeoLite2-ASN-lite.mmdb) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN-lite.mmdb) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN-lite.mmdb) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN-lite.mmdb) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN-lite.mmdb) |
| **`BundleMRS.7z`** | Mihomo `.mrs` 规则集压缩包 | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/BundleMRS.7z) | [下载](https://cdn.jsdmirror.com/gh/appshubcc/bett-rules@release/BundleMRS.7z) | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/BundleMRS.7z) | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/BundleMRS.7z) | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/BundleMRS.7z) | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/BundleMRS.7z) |

### 2. Rule-Set 规则集分支

- **Mihomo (`.mrs` / `.yaml` / `.list`)**：[`meta` 分支](https://github.com/appshubcc/bett-rules/tree/meta)
- **sing-box (`.srs` / `.json`)**：[`sing` 分支](https://github.com/appshubcc/bett-rules/tree/sing)
- **Fake-IP Filter 独立规则**：Release 与 `release` 分支同时提供 `fakeip-filter.mrs`、`fakeip-filter.srs`、`fakeip-filter.yaml`、`fakeip-filter.json`、`fakeip-filter.list`。

---

## 数据库内容说明

### 1. GeoIP (`country*.mmdb` / `geoip*.dat` / `geoip*.metadb` / `geoip*.db`)

- **全量版**：包含全球所有国家/地区代码及局域网保留网段 (`geoip:private`)。
- **轻量版 (`lite`)**：国家码仅保留 `CN` 与 `private`，精简文件体积。
- **内置常用服务分类**（全量版与轻量版均包含，已去除中国大陆节点）：
  - `apple`、`akamai`、`bilibili`、`cloudflare`、`cloudfront`、`facebook`、`fastly`、`google`、`microsoft`、`netflix`、`openai`、`spotify`、`steam`、`telegram`、`tiktok`、`twitter`

### 2. ASN (`GeoLite2-ASN.mmdb` / `GeoLite2-ASN-lite.mmdb`)

- **`GeoLite2-ASN.mmdb`**：包含全球自治系统号（ASN）及组织名称。
- **`GeoLite2-ASN-lite.mmdb`**：轻量版本，仅保留上述常用跨国服务与主流 CDN 的 ASN 段。

### 3. GeoSite 全量版 (`geosite.dat` / `geosite.db`)

用法同 [Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat)：

- **`geosite:cn`**：以 [`dnsmasq-china-list`](https://github.com/felixonmars/dnsmasq-china-list) 为基础，经去重与国内外 DNS 解析校验整理。
- **`geosite:geolocation-!cn`**：整合 GFWList、社区规则与常用海外域名，经海外顶级域（TLD）去重并剔除国内直连子域。
- **`geosite:category-ads-all`**：仅整合少量常见广告域名并内置防误杀白名单，保持精简。
- **`geosite:steam@cn`**：合并 [`ios_rule_script/SteamCN`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/SteamCN) 国内节点数据。
- **`geosite:onedrive` / `geosite:sharepoint`**：合并 [`ios_rule_script/OneDrive`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/OneDrive) 与 [`xishang0128/sharepoint`](https://raw.githubusercontent.com/xishang0128/rules/main/sharepoint.list)。
- **新增分类**：
  - `geosite:biliintl`：来源 [`xishang0128/biliintl`](https://raw.githubusercontent.com/xishang0128/rules/main/biliintl.list)
  - `geosite:tracker`：来源 [`XIU2/TrackersListCollection`](https://trackerslist.com/#/zh) 与 [`ios_rule_script/PrivateTracker`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/PrivateTracker)
  - `geosite:httpdns`：来源 [`ios_rule_script/BlockHttpDNS`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/BlockHttpDNS)
  - `geosite:fakeip-filter`：整理自 [`ShellCrash/fake_ip_filter.list`](https://raw.githubusercontent.com/juewuy/ShellCrash/dev/public/fake_ip_filter.list) 及常用本地/STUN/游戏联机域名

### 4. GeoSite 轻量版 (`geosite-lite.dat` / `geosite-lite.db`)

仅包含以下 26 个常用集合（主要整理自 [`blackmatrix7/ios_rule_script`](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash)）：

```text
abema / apple / applemusic / bahamut / bilibili / biliintl / category-ads /
cloudflare / cn / ehentai / fakeip-filter / github / google / microsoft /
netflix / onedrive / openai / pixiv / private / proxy / proxymedia /
spotify / telegram / tiktok / twitter / youtube
```

---

## 配置示例

### 1. Mihomo `geox-url` 配置

默认使用 MetaDB 模式（`geodata-mode: false`），在此模式下 **`geoip.metadb` 需填写在 `geox-url.mmdb` 字段**：

```yaml
geodata-mode: false

geox-url:
  # 如需轻量版，将文件名替换为 geoip-lite.metadb / geoip-lite.dat / geosite-lite.dat / GeoLite2-ASN-lite.mmdb
  mmdb: "https://github.com/appshubcc/bett-rules/releases/download/latest/geoip.metadb"
  geoip: "https://github.com/appshubcc/bett-rules/releases/download/latest/geoip.dat"
  geosite: "https://github.com/appshubcc/bett-rules/releases/download/latest/geosite.dat"
  asn: "https://github.com/appshubcc/bett-rules/releases/download/latest/GeoLite2-ASN.mmdb"
```

### 2. Mihomo 规则与 DNS 配置参考

```yaml
rule-providers:
  cn:
    type: http
    behavior: domain
    format: mrs
    interval: 86400
    path: ./provider/rule-set/geosite-cn.mrs
    url: "https://raw.githubusercontent.com/appshubcc/bett-rules/meta/geo/geosite/cn.mrs"

dns:
  enable: true
  enhanced-mode: fake-ip
  fake-ip-range: 198.18.0.1/16
  fake-ip-filter:
    - "geosite:fakeip-filter"
    - "geosite:private"
    - "geosite:cn"
  nameserver-policy:
    "geosite:cn,private,apple-cn":
      - https://doh.pub/dns-query
      - https://dns.alidns.com/dns-query
    "geosite:category-ads-all": rcode://success

rules:
  - GEOSITE,category-ads-all,REJECT
  - GEOSITE,private,DIRECT
  - GEOSITE,biliintl,PROXY
  - GEOSITE,apple-cn,DIRECT
  - GEOSITE,steam@cn,DIRECT
  - GEOSITE,microsoft@cn,DIRECT
  - GEOSITE,category-games@cn,DIRECT
  - GEOSITE,geolocation-!cn,PROXY
  - RULE-SET,cn,DIRECT

  # GeoIP 规则
  - GEOIP,private,DIRECT,no-resolve
  - GEOIP,telegram,PROXY,no-resolve
  - GEOIP,CN,DIRECT
  - MATCH,PROXY
```

---

## 辅助工具与致谢

- [MetaCubeX/geo](https://github.com/MetaCubeX/geo)
- [v2fly/domain-list-community](https://github.com/v2fly/domain-list-community)
- [Loyalsoldier/domain-list-custom](https://github.com/Loyalsoldier/domain-list-custom)
- [felixonmars/dnsmasq-china-list](https://github.com/felixonmars/dnsmasq-china-list)
- [gaoyifan/china-operator-ip](https://github.com/gaoyifan/china-operator-ip)
- [IP2Location LITE](https://lite.ip2location.com)
- [gfwlist/gfwlist](https://github.com/gfwlist/gfwlist) & [cokebar/gfwlist2dnsmasq](https://github.com/cokebar/gfwlist2dnsmasq)
- [crazy-max/WindowsSpyBlocker](https://github.com/crazy-max/WindowsSpyBlocker)
- [blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script)
