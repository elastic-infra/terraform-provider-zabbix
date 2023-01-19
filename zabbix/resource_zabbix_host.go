package zabbix

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// HostInterfaceTypes zabbix different interface type
var HostInterfaceTypes = EnumMap{
	"agent": int(zabbix.AgentInterface),
	"snmp":  int(zabbix.SNMPInterface),
	"ipmi":  int(zabbix.IPMIInterface),
	"jmx":   int(zabbix.JMXInterface),
}

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
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "Visible name of the host.",
			},
			"monitored": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"interfaces": {
				Type:     schema.TypeList,
				Elem:     interfaceSchema,
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
		},
	}
}

var interfaceSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"dns": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"ip": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"main": {
			Type:     schema.TypeBool,
			Required: true,
		},
		"port": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "10050",
		},
		"type": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "agent",
		},
		"interface_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"snmp_configs": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"version": {
						Type:     schema.TypeInt,
						Required: true,
					},
					"community": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"snmpv3_config": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
	},
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
		HostID: data.Id(),
		Host:   data.Get("host").(string),
		Name:   data.Get("name").(string),
		Status: 0,
	}
	//0 is monitored, 1 - unmonitored host
	if !data.Get("monitored").(bool) {
		host.Status = 1
	}
	host.GroupIds = generateHostGroupIDsFromResourceData(data)
	interfaces, err := generateInterfacesFromResourceData(data)
	if err != nil {
		return nil, err
	}
	host.Interfaces = interfaces
	host.TemplateIDs = generateTemplateIDsFromResourceData(data)
	host.Macros = generateHostMacrosFromResourceData(data)
	return &host, nil
}

func generateInterfacesFromResourceData(d *schema.ResourceData) ([]zabbix.HostInterface, error) {
	interfaceCount := d.Get("interfaces.#").(int)
	interfaces := make(zabbix.HostInterfaces, interfaceCount)
	for i := 0; i < interfaceCount; i++ {
		prefix := fmt.Sprintf("interfaces.%d.", i)
		interfaceType := d.Get(prefix + "type").(string)
		typeID, ok := HostInterfaceTypes[interfaceType]
		if !ok {
			return nil, fmt.Errorf("%s isnt valid interface type", interfaceType)
		}
		ip := d.Get(prefix + "ip").(string)
		dns := d.Get(prefix + "dns").(string)
		interfaceId := d.Get(prefix + "interface_id").(string)
		if ip == "" && dns == "" {
			return nil, fmt.Errorf("atleast one of two dns or ip must be set")
		}
		useip := 1
		if ip == "" {
			useip = 0
		}
		main := 1
		if !d.Get(prefix + "main").(bool) {
			main = 0
		}
		interfaces[i] = zabbix.HostInterface{
			InterfaceID: interfaceId,
			IP:          ip,
			DNS:         dns,
			Main:        main,
			Port:        d.Get(prefix + "port").(string),
			Type:        zabbix.InterfaceType(typeID),
			UseIP:       useip,
		}
	}
	return interfaces, nil
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
	var macros []zabbix.Macro
	for _, macro := range macroMaps {
		macroMap := macro.(map[string]any)
		macros = append(macros,
			zabbix.Macro{
				MacroName: macroMap["name"].(string),
				Value:     macroMap["value"].(string),
			},
		)
	}
	return macros
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
	var interfaces []map[string]any
	for _, hostInterface := range host.Interfaces {
		interfaceType := HostInterfaceTypes.getStringType(int(hostInterface.Type))
		isMain := true
		if hostInterface.Main == 0 {
			isMain = false
		}
		interfaces = append(interfaces, map[string]any{
			"dns":          hostInterface.DNS,
			"ip":           hostInterface.IP,
			"main":         isMain,
			"port":         hostInterface.Port,
			"type":         interfaceType,
			"interface_id": hostInterface.InterfaceID,
		})
	}
	errors.addError(data.Set("interfaces", interfaces))
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
