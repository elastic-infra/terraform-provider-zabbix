---
layout: "zabbix"
page_title: "Zabbix: zabbix_proxy"
sidebar_current: "docs-zabbix-data-source-proxy"
description: |-
  Provides a Zabbix Proxy data source. This can be used to get information about the Zabbix Proxy.
---

# zabbix_proxy

Provides a zabbix proxy data source. This can be used to get information about the Zabbix proxy.

## Example Usage

Get the ID of a Zabbix Proxy

```hcl
data "zabbix_proxy" "demo_proxy" {
  name = "proxy-name"
}

resource "zabbix_host" "host" {
  host = "host"
  name = "name"

  proxy_host_id = data.zabbix_proxy.demo_proxy.id
  groups        = []
  templates     = []
}
```

## Attributes

* `name` - name of the Zabbix Proxy.
