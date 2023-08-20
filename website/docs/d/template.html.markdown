---
layout: "zabbix"
page_title: "Zabbix: zabbix_template"
sidebar_current: "docs-zabbix-data-source-template"
description: |-
  Provides a Zabbix Template data source. This can be used to get information about the Zabbix Template.
---

# zabbix_template

Provides a zabbix template data source. This can be used to get information about the Zabbix template.

## Example Usage

Get the ID of a Zabbix Template

```hcl
data "zabbix_template" "demo_template" {
  name = "template-name"
}

output "id" {
  value = data.zabbix_template.demo_template.id
}
```

## Attributes

* `name` - name of the Zabbix Template.
