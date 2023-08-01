package zabbix

import (
	"fmt"
	"github.com/atypon/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceZabbixProxy() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceZabbixProxyRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    false,
				Required:    true,
				Description: "Proxy Group Template to Read",
			},
		},
	}
}

func dataSourceZabbixProxyRead(d *schema.ResourceData, meta interface{}) (err error) {

	ProxyName := d.Get("name")

	ProxyID := &zabbix.Proxy{Name: fmt.Sprint(ProxyName)}
	d.SetId(ProxyID.ProxyID)
	if err != nil {
		log.Printf("[WARN] Failed to get Proxy ID from Input Paramater: %v\n", err)
		return err
	}

	return nil
}
