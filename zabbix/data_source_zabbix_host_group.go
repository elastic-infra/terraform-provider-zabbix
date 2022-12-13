package zabbix

import (
	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceZabbixHostGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceZabbixHostGroupRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    false,
				Required:    true,
				Description: "Host Group Template to Read",
			},
		},
	}
}

func dataSourceZabbixHostGroupRead(d *schema.ResourceData, meta interface{}) (err error) {

	hostGroupName := d.Get("name")
	api := meta.(*zabbix.API)
	params := map[string]interface{}{
		"output": "extend",
		"filter": map[string]interface{}{
			"name": hostGroupName,
		},
	}
	if hostGroupName == "Templates" {
		d.SetId("1")
		return nil
	} else {
		hostGroupID, err := api.HostGroupsGet(params)
		d.SetId(hostGroupID[0].GroupID)
		if err != nil {
			log.Printf("[WARN] Failed to get Host Group ID from Input Paramater: %v\n", err)
			return err
		}
	}

	return nil
}
