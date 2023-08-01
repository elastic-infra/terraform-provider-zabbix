package zabbix

import (
	"github.com/atypon/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceZabbixHost() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceZabbixHostRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    false,
				Required:    true,
				Description: "Host to Read",
			},
		},
	}
}

func dataSourceZabbixHostRead(d *schema.ResourceData, meta interface{}) (err error) {

	HostName := d.Get("name")
	api := meta.(*zabbix.API)
	params := map[string]interface{}{
		"output": "extend",
		"filter": map[string]interface{}{
			"name": HostName,
		},
	}

	HostID, err := api.HostsGet(params)
	d.SetId(HostID[0].HostID)
	if err != nil {
		log.Printf("[WARN] Failed to get Host ID from Input Paramater: %v\n", err)
		return err
	}

	return nil
}
