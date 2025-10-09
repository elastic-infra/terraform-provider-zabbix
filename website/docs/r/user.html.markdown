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
  passwd     = "securepassword"
  url        = "https://example.com"
  autologin  = 0
  autologout = "15m"
  lang       = "en_US"
  refresh    = "30s"
  theme      = "default"
  type       = 1
  
  role_ids = ["3"]
  
  user_groups {
    usrgrpid = "7"
  }
  
  user_medias {
    mediatypeid = "1"
    sendto      = ["user@example.com"]
    active      = 0
    severity    = 63
    period      = "1-7,00:00-24:00"
  }
}
```

## Argument Reference

The following arguments are supported:

* `username` - (Required) User login name.
* `name` - (Optional) User first name.
* `surname` - (Optional) User last name.
* `passwd` - (Optional) User password. Required when creating a new user.
* `url` - (Optional) URL after successful login.
* `autologin` - (Optional) Whether to enable auto-login. Can be `0` (default, disabled) or `1` (enabled).
* `autologout` - (Optional) User session life time in seconds. If set to 0, the session will never expire. Default is `15m`.
* `lang` - (Optional) Language code of the user's locale. Default is `en_US`.
* `refresh` - (Optional) Automatic refresh period in seconds. Default is `30s`.
* `theme` - (Optional) User's theme. Can be `default`, `blue-theme`, `dark-theme`.
* `type` - (Optional) Type of the user. Can be `1` (Zabbix user, default), `2` (Zabbix admin), `3` (Zabbix super admin).
* `role_ids` - (Optional) List of role IDs assigned to the user.
* `user_groups` - (Optional, Multiple) List of user groups the user belongs to.
  * `usrgrpid` - (Required) User group ID.
* `user_medias` - (Optional, Multiple) List of user media (notification methods).
  * `mediatypeid` - (Required) Media type ID.
  * `sendto` - (Required) List of addresses, phone numbers or other identifiers where notifications will be sent.
  * `active` - (Optional) Whether the media is enabled. Can be `0` (enabled, default) or `1` (disabled).
  * `severity` - (Optional) Trigger severities to send notifications about. Default is `63` (all severities).
  * `period` - (Optional) Time when the notifications can be sent as a time period or user macro. Default is `1-7,00:00-24:00`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `user_id` - The zabbix user ID.
* `user_groups`
  * `usrgrpid` - User group ID.
* `user_medias`
  * `mediaid` - Media ID.
  * `mediatypeid` - Media type ID.
  * `sendto` - List of addresses where notifications are sent.
  * `active` - Media status.
  * `severity` - Trigger severities.
  * `period` - Time period for notifications.

## Import

Users can be imported using their user ID:

```
$ terraform import zabbix_user.demo_user 123
```
