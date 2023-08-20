---
layout: "zabbix"
page_title: "Zabbix: zabbix_host_group"
sidebar_current: "docs-zabbix-data-source-host_group"
description: |-
  Provides a Zabbix Host Group data source. This can be used to get information about the Zabbix Host Group.
---

# zabbix_host_group

Provides a zabbix host_group data source. This can be used to get information about the Zabbix Host Group.

## Example Usage

Get the ID of a Zabbix Host Group

```hcl
data "zabbix_host_group" "demo_host_group" {
  name = "host_group-name"
}

output "id" {
  value = data.zabbix_host_group.demo_host_group.id
}
```

## Attributes

* `name` - name of the Zabbix Host Group.
