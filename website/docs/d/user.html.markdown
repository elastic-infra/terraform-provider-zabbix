---
layout: "zabbix"
page_title: "Zabbix: zabbix_user"
sidebar_current: "docs-zabbix-data-source-user"
description: |-
  Provides a Zabbix User data source. This can be used to get information about existing Zabbix users.
---

# zabbix_user

Provides a zabbix user data source. This can be used to get information about existing Zabbix users.

## Example Usage

Get information about a user by username

```hcl
data "zabbix_user" "demo_user" {
  username = "demouser"
}

output "user_id" {
  value = data.zabbix_user.demo_user.user_id
}

output "user_type" {
  value = data.zabbix_user.demo_user.type
}
```

Get information about a user by user ID

```hcl
data "zabbix_user" "demo_user_by_id" {
  user_id = "123"
}

output "username" {
  value = data.zabbix_user.demo_user_by_id.username
}
```

## Argument Reference

The following arguments are supported (at least one is required):

* `username` - (Optional) User login name to search for.
* `user_id` - (Optional) User ID to search for.

## Attributes

* `user_id` - The zabbix user ID.
* `username` - User login name.
* `name` - User first name.
* `surname` - User last name.
* `url` - URL after successful login.
* `autologin` - Whether auto-login is enabled.
* `autologout` - User session life time.
* `lang` - Language code of the user's locale.
* `refresh` - Automatic refresh period.
* `theme` - User's theme.
* `type` - Type of the user (1=Zabbix user, 2=Zabbix admin, 3=Zabbix super admin).
* `role_ids` - List of role IDs assigned to the user.
* `user_groups` - List of user groups the user belongs to.
  * `usrgrpid` - User group ID.
* `user_medias` - List of user media (notification methods).
  * `mediaid` - Media ID.
  * `mediatypeid` - Media type ID.
  * `sendto` - List of addresses where notifications are sent.
  * `active` - Media status.
  * `severity` - Trigger severities.
  * `period` - Time period for notifications.

