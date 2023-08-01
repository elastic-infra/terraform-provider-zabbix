package zabbix

import (
	"github.com/atypon/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func dataSourceZabbixTemplate() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceZabbixTemplateRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    false,
				Required:    true,
				Description: "Template to Read",
			},
		},
	}
}

func dataSourceZabbixTemplateRead(d *schema.ResourceData, meta interface{}) (err error) {

	templateName := d.Get("name")
	api := meta.(*zabbix.API)
	params := map[string]interface{}{
		"output": "extend",
		"filter": map[string]interface{}{
			"name": templateName,
		},
	}
	if templateName == "Templates" {
		d.SetId("1")
		return nil
	} else {
		templateID, err := api.TemplatesGet(params)
		d.SetId(templateID[0].TemplateID)
		if err != nil {
			log.Printf("[WARN] Failed to get Template ID from Input Paramater: %v\n", err)
			return err
		}
	}

	return nil
}
