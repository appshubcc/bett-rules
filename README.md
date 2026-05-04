## **下载地址**：

| 文件名              | Github                                                                                                            | JSdelivr                                                                                                                           | Cloudflare                                                                                                                              | Fastly                                                                                                                              | Gcore                                                                                                                              |
|---------------------|---------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------|
| country.mmdb        | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/country.mmdb)                                 | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/country.mmdb)                                                  | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/country.mmdb)                                                 | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/country.mmdb)                                                | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/country.mmdb)                                                |
| geoip.dat           | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geoip.dat)                                    | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.dat)                                                     | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.dat)                                                    | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.dat)                                                   | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.dat)                                                   |
| geoip.db            | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geoip.db)                                     | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.db)                                                      | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.db)                                                     | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.db)                                                    | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.db)                                                    |
| geoip.metadb        | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geoip.metadb)                                 | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.metadb)                                                  | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.metadb)                                                 | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.metadb)                                                | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geoip.metadb)                                                |
| geosite.dat         | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geosite.dat)                                  | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.dat)                                                   | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.dat)                                                  | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.dat)                                                 | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.dat)                                                 |
| geosite.db          | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/geosite.db)                                   | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.db)                                                    | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.db)                                                   | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.db)                                                  | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/geosite.db)                                                  |
| GeoLite2-ASN.mmdb   | [下载](https://github.com/appshubcc/bett-rules/releases/download/latest/GeoLite2-ASN.mmdb)                            | [下载](https://cdn.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN.mmdb)                                             | [下载](https://testingcf.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN.mmdb)                                            | [下载](https://fastly.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN.mmdb)                                           | [下载](https://gcore.jsdelivr.net/gh/appshubcc/bett-rules@release/GeoLite2-ASN.mmdb)                                           |

### **mihomo rule-set (.mrs格式)**

mihomo：[meta branch](https://github.com/MetaCubeX/meta-rules-dat/tree/meta)

## **country.mmdb,geoip.dat,geoip.db 内容**

同 [Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat)

- 新增类别（方便有特殊需求的用户使用）：
  - `geoip:cloudflare`
  - `geoip:cloudfront`
  - `geoip:facebook`
  - `geoip:fastly`
  - `geoip:google`
  - `geoip:netflix`
  - `geoip:telegram`
  - `geoip:twitter`

## **geosite.dat,geosite.db 内容**

用法同 [Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat)

- `geosite:category-ads-all` 仅使用域名作为广告拦截用途作用有限，因此不作额外域名添加
- `geosite:cn` 源替换回 [Loyalsoldier/v2ray-rules-dat](https://github.com/Loyalsoldier/v2ray-rules-dat/tree/meta/geo/geosite/cn.yaml)
- `geosite:onedrive` 合并 [ios_rule_script/OneDrive](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/OneDrive)
- `geosite:onedrive` 合并 [ios_rule_script/OneDrive](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/OneDrive)
- `geosite:steam@cn` 合并 [ios_rule_script/SteamCN](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/SteamCN) 的内数据
- 新增类别 - `geosite:biliintl` 来源 [biliintl](https://raw.githubusercontent.com/xishang0128/rules/main/biliintl.list) - `geosite:tracker` 来源 [TrackersList](https://trackerslist.com/#/zh)以及[blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script/tree/master/rule/Clash/PrivateTracker)

## **示例**

```yaml
rule-providers:
  cn:
    behavior: domain
    interval: 86400
    path: ./provider/rule-set/cn_domain.yaml
    type: http
    url: "https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/meta/geo/geosite/cn.yaml"

dns:
  nameserver-policy:
    "geosite:cn,private,apple":
      - https://doh.pub/dns-query
      - https://dns.alidns.com/dns-query
    "geosite:category-ads-all": rcode://success

rules:
  - GEOSITE,category-ads-all,REJECT
  - GEOSITE,private,DIRECT
  - GEOSITE,youtube,PROXY
  - GEOSITE,google,PROXY
  - GEOSITE,twitter,PROXY
  - GEOSITE,pixiv,PROXY
  - GEOSITE,category-scholar-!cn,PROXY
  - GEOSITE,biliintl,PROXY
  - GEOSITE,onedrive,DIRECT
  - GEOSITE,microsoft@cn,DIRECT
  - GEOSITE,apple-cn,DIRECT
  - GEOSITE,steam@cn,DIRECT
  - GEOSITE,category-games@cn,DIRECT
  - GEOSITE,geolocation-!cn,PROXY

  #GEOIP 规则
  - GEOIP,private,DIRECT,no-resolve
  - GEOIP,telegram,PROXY,no-resolve
  - GEOIP,CN,DIRECT
  - DST-PORT,22,DIRECT
  - MATCH,PROXY
```

## 辅助工具

https://github.com/MetaCubeX/geo

🗺 An easy way to manage all your Geo resources.

## 致谢

- [@Loyalsoldier/geoip](https://github.com/Loyalsoldier/geoip)
- [@v2fly/domain-list-community](https://github.com/v2fly/domain-list-community)
- [@Loyalsoldier/domain-list-custom](https://github.com/Loyalsoldier/domain-list-custom)
- [@felixonmars/dnsmasq-china-list](https://github.com/felixonmars/dnsmasq-china-list)
- [@gfwlist/gfwlist](https://github.com/gfwlist/gfwlist)
- [@cokebar/gfwlist2dnsmasq](https://github.com/cokebar/gfwlist2dnsmasq)
- [@Loyalsoldier/cn-blocked-domain](https://github.com/Loyalsoldier/cn-blocked-domain)
- [@AdblockPlus/EasylistChina+Easylist.txt](https://easylist-downloads.adblockplus.org/easylistchina+easylist.txt)
- [@AdGuard/DNS-filter](https://kb.adguard.com/en/general/adguard-ad-filters#dns-filter)
- [@PeterLowe/adservers](https://pgl.yoyo.org/adservers)
- [@DanPollock/hosts](https://someonewhocares.org/hosts)
- [@crazy-max/WindowsSpyBlocker](https://github.com/crazy-max/WindowsSpyBlocker)
- [@blackmatrix7/ios_rule_script](https://github.com/blackmatrix7/ios_rule_script)

