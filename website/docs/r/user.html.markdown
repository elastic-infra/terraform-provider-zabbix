---
layout: "zabbix"
page_title: "Zabbix: zabbix_user"
sidebar_current: "docs-zabbix-resource-user"
description: |-
  Provides a zabbix user resource. This can be used to create and manage Zabbix Users.
---

# zabbix_user

A [user](https://www.zabbix.com/documentation/current/manual/api/reference/user) represents a Zabbix user account with authentication and authorization capabilities.

## Example Usage

Create a new user

```hcl
resource "zabbix_user" "demo_user" {
  username   = "demouser"
  name       = "Demo"
  surname    = "User"
  password   = "securepassword"
  role_id    = "3"
  autologin  = false
  autologout = "15m"
  lang       = "en_US"
  refresh    = "30s"
  theme      = "default"

  user_groups = ["Zabbix administrators"]

  medias {
    mediatypeid = "1"
    sendto      = ["user@example.com"]
    active      = true
    severity    = 63
    period      = "1-7,00:00-24:00"
  }
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Required) User login name.
* `password` - (Required) User password.
* `role_id` - (Required) ID of the user role.
* `name` - (Optional) User first name.
* `surname` - (Optional) User last name.
* `url` - (Optional) URL of the page to redirect the user to after logging in.
* `autologin` - (Optional) Whether to enable auto-login. Default is `false`.
* `autologout` - (Optional) User session life time. If set to `"0"`, the session will never expire. Default is `"0"`.
* `lang` - (Optional) Language code of the user's language. Default is `"default"`.
* `refresh` - (Optional) Automatic refresh period. Default is `"30s"`.
* `theme` - (Optional) User's theme. Can be `"default"`, `"blue-theme"`, `"dark-theme"`. Default is `"default"`.
* `rows_per_page` - (Optional) Amount of object rows to show per page. Must be between 1 and 999999. Default is `50`.
* `timezone` - (Optional) User's timezone. Default is `"default"`.
* `user_groups` - (Optional) Set of user group names to add the user to.
* `medias` - (Optional) List of user media (notification methods). Each `medias` block supports:
  * `mediatypeid` - (Required) ID of the media type used by the media.
  * `sendto` - (Required) List of addresses, phone numbers or other identifiers where notifications will be sent.
  * `active` - (Optional) Whether the media is enabled. Default is `true`.
  * `severity` - (Optional) Trigger severities to send notifications about as a bitmask. Must be between 1 and 63. Default is `63` (all severities).
  * `period` - (Optional) Time when the notifications can be sent as a time period or user macro. Default is `"1-7,00:00-24:00"`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The zabbix user ID.

## Import

Users can be imported using their user ID:

```
$ terraform import zabbix_user.demo_user 123
```

~> **Note:** The `password` attribute cannot be read from the Zabbix API. After importing, the first `terraform plan` will show a diff for `password`. Running `terraform apply` will resolve this.
