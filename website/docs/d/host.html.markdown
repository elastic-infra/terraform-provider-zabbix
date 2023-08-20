---
layout: "zabbix"
page_title: "Zabbix: zabbix_host"
sidebar_current: "docs-zabbix-data-source-host"
description: |-
  Provides a Zabbix Host data source. This can be used to get information about the Zabbix Host.
---

# zabbix_host

Provides a zabbix host data source. This can be used to get information about the Zabbix Host.

## Example Usage

Get the ID of a Zabbix Host

```hcl
data "zabbix_host" "demo_host" {
  name = "host-name"
}

output "id" {
  value = data.zabbix_host.demo_host.id
}
```

## Attributes

* `name` - name of the Zabbix Host.
