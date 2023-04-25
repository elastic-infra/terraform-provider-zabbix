package zabbix

import (
	"context"
	"github.com/atypon/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceZabbixHost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixHostCreate,
		ReadContext:   resourceZabbixHostRead,
		UpdateContext: resourceZabbixHostUpdate,
		DeleteContext: resourceZabbixHostDelete,

		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Technical name of the host.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Visible name of the host.",
			},
			"monitored": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"inventory_mode": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"ipmi_username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipmi_password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"ipmi_auth_type": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"ipmi_privilege": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"proxy_host_id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"groups": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
			"templates": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
			"macro": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"tags": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceZabbixHostCreate(context context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	host, err := createHostObjectFromResourceData(data)
	if err != nil {
		return diag.FromErr(err)
	}
	err = api.CreateAPIObject(host)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(host.HostID)
	readDiags := resourceZabbixHostRead(context, data, meta)
	return readDiags
}

func createHostObjectFromResourceData(data *schema.ResourceData) (*zabbix.Host, error) {
	host := zabbix.Host{
		HostID:        data.Id(),
		Host:          data.Get("host").(string),
		Name:          data.Get("name").(string),
		Status:        0,
		Description:   data.Get("description").(string),
		InventoryMode: data.Get("inventory_mode").(int),
		IPMIUsername:  data.Get("ipmi_username").(string),
		IPMIPassword:  data.Get("ipmi_password").(string),
		IPMIAuthType:  data.Get("ipmi_auth_type").(int),
		IPMIPrivilege: data.Get("ipmi_privilege").(int),
		ProxyHostID:   data.Get("proxy_host_id").(string),
	}
	//0 is monitored, 1 - unmonitored host
	if !data.Get("monitored").(bool) {
		host.Status = 1
	}
	host.GroupIds = generateHostGroupIDsFromResourceData(data)
	host.TemplateIDs = generateTemplateIDsFromResourceData(data)
	host.Macros = generateHostMacrosFromResourceData(data)
	host.Tags = generateHostTagsFromResourceData(data)
	return &host, nil
}

func generateHostGroupIDsFromResourceData(data *schema.ResourceData) []zabbix.HostGroupID {
	groupIDsList := data.Get("groups").(*schema.Set).List()
	hostGroupIDs := make([]zabbix.HostGroupID, len(groupIDsList))
	for i, ID := range groupIDsList {
		hostGroupIDs[i] = zabbix.HostGroupID{GroupID: ID.(string)}
	}
	return hostGroupIDs
}

func generateTemplateIDsFromResourceData(d *schema.ResourceData) []zabbix.TemplateID {
	templateIDsList := d.Get("templates").(*schema.Set).List()
	templateIDs := make([]zabbix.TemplateID, len(templateIDsList))
	for i, ID := range templateIDsList {
		templateIDs[i] = zabbix.TemplateID{TemplateID: ID.(string)}
	}
	return templateIDs
}

func generateHostMacrosFromResourceData(data *schema.ResourceData) []zabbix.Macro {
	macroMaps := data.Get("macro").([]any)
	macros := make([]zabbix.Macro, len(macroMaps))
	for i, macro := range macroMaps {
		macroMap := macro.(map[string]any)
		macros[i] = zabbix.Macro{
			MacroName: macroMap["name"].(string),
			Value:     macroMap["value"].(string),
		}
	}
	return macros
}

func generateHostTagsFromResourceData(data *schema.ResourceData) []zabbix.HostTag {
	tagMaps := data.Get("tags").(map[string]any)
	tags := make([]zabbix.HostTag, len(tagMaps))
	i := 0
	for tag, value := range tagMaps {
		tags[i] = zabbix.HostTag{
			Tag:   tag,
			Value: value.(string),
		}
		i++
	}
	return tags
}

func resourceZabbixHostRead(context context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	var errors TerraformErrors
	host := &zabbix.Host{HostID: data.Id()}
	err := api.ReadAPIObject(host)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[DEBUG] Read host: %+v\n", host)
	errors.addError(data.Set("host", host.Host))
	errors.addError(data.Set("name", host.Name))
	errors.addError(data.Set("monitored", host.Status == 0))
	errors.addError(data.Set("description", host.Description))
	errors.addError(data.Set("inventory_mode", host.InventoryMode))
	errors.addError(data.Set("ipmi_username", host.IPMIUsername))
	errors.addError(data.Set("ipmi_password", host.IPMIPassword))
	errors.addError(data.Set("ipmi_auth_type", host.IPMIAuthType))
	errors.addError(data.Set("ipmi_privilege", host.IPMIPrivilege))
	errors.addError(data.Set("proxy_host_id", host.ProxyHostID))
	var macros []any
	for _, macro := range host.Macros {
		macros = append(macros, map[string]any{
			"name":  macro.MacroName,
			"value": macro.Value,
		})
	}
	errors.addError(data.Set("macro", macros))
	params := zabbix.Params{
		"output": "extend",
		"hostids": []string{
			data.Id(),
		},
	}
	templates, err := api.TemplatesGet(params)
	if err != nil {
		return diag.FromErr(err)
	}
	templateIDs := make([]string, len(templates))
	for i, t := range templates {
		templateIDs[i] = t.TemplateID
	}
	errors.addError(data.Set("templates", templateIDs))
	groupIDs := host.GroupIds
	groups := make([]string, len(groupIDs))
	for i, group := range groupIDs {
		groups[i] = group.GroupID
	}
	errors.addError(data.Set("groups", groups))
	tags := make(map[string]string)
	for _, tag := range host.Tags {
		tags[tag.Tag] = tag.Value
	}
	errors.addError(data.Set("tags", tags))
	return errors.getDiagnostics()
}

func resourceZabbixHostUpdate(context context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	host, err := createHostObjectFromResourceData(data)
	if err != nil {
		return diag.FromErr(err)
	}
	err = api.UpdateAPIObject(host)
	if err != nil {
		return diag.FromErr(err)
	}
	readDiags := resourceZabbixHostRead(context, data, meta)
	return readDiags
}

func resourceZabbixHostDelete(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	host := &zabbix.Host{HostID: d.Id()}
	err := api.DeleteAPIObject(host)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}
