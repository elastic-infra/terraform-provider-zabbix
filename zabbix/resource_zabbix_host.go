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

var interfaceSchema *schema.Resource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"dns": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"ip": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
		},
		"main": &schema.Schema{
			Type:     schema.TypeBool,
			Required: true,
		},
		"port": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "10050",
		},
		"type": &schema.Schema{
			Type:     schema.TypeString,
			Optional: true,
			Default:  "agent",
		},
		"interface_id": &schema.Schema{
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

func resourceZabbixHost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixHostCreate,
		ReadContext:   resourceZabbixHostRead,
		UpdateContext: resourceZabbixHostUpdate,
		DeleteContext: resourceZabbixHostDelete,
		Schema: map[string]*schema.Schema{
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Technical name of the host.",
			},
			"host_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				ForceNew:    true,
				Description: "(readonly) ID of the host",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				Computed:    true,
				Description: "Visible name of the host.",
			},
			"monitored": &schema.Schema{
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			//any changes to interface will trigger recreate, zabbix api kinda doesn't
			//work nicely, interface can get linked to various things and replacement
			//simply doesn't work
			"interfaces": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     interfaceSchema,
				Optional: true,
			},
			"groups": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Required: true,
			},
			"templates": &schema.Schema{
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

func getInterfaces(d *schema.ResourceData) (zabbix.HostInterfaces, error) {
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

func getHostGroups(d *schema.ResourceData, api *zabbix.API) (zabbix.HostGroupIDs, error) {
	configGroups := d.Get("groups").(*schema.Set)
	setHostGroups := make([]string, configGroups.Len())

	for i, g := range configGroups.List() {
		setHostGroups[i] = g.(string)
	}

	log.Printf("[DEBUG] Groups %v\n", setHostGroups)

	groupParams := zabbix.Params{
		"output": "extend",
		"filter": map[string]interface{}{
			"name": setHostGroups,
		},
	}

	groups, err := api.HostGroupsGet(groupParams)

	if err != nil {
		return nil, err
	}

	if len(groups) < configGroups.Len() {
		log.Printf("[DEBUG] Not all of the specified groups were found on zabbix server")

		for _, n := range configGroups.List() {
			found := false

			for _, g := range groups {
				if n == g.Name {
					found = true
					break
				}
			}

			if !found && n != "1" {
				return nil, fmt.Errorf("Host group %s doesnt exist in zabbix server", n)
			}
			log.Printf("[DEBUG] %s exists on zabbix server", n)
		}
	}

	hostGroups := make(zabbix.HostGroupIDs, len(groups))

	for i, g := range groups {
		hostGroups[i] = zabbix.HostGroupID{
			GroupID: g.GroupID,
		}
	}

	return hostGroups, nil
}

func getTemplates(d *schema.ResourceData, api *zabbix.API) (zabbix.TemplateIDs, error) {
	configTemplates := d.Get("templates").(*schema.Set)
	templateNames := make([]string, configTemplates.Len())

	if configTemplates.Len() == 0 {
		return nil, nil
	}

	for i, g := range configTemplates.List() {
		templateNames[i] = g.(string)
	}

	log.Printf("[DEBUG] Templates %v\n", templateNames)

	groupParams := zabbix.Params{
		"output": "extend",
		"filter": map[string]interface{}{
			"host": templateNames,
		},
	}

	templates, err := api.TemplatesGet(groupParams)

	if err != nil {
		return nil, err
	}

	if len(templates) < configTemplates.Len() {
		log.Printf("[DEBUG] Not all of the specified templates were found on zabbix server")

		for _, n := range configTemplates.List() {
			found := false

			for _, g := range templates {
				if n == g.Name {
					found = true
					break
				}
			}

			if !found {
				return nil, fmt.Errorf("Template %s doesnt exist in zabbix server", n)
			}
			log.Printf("[DEBUG] Template %s exists on zabbix server", n)
		}
	}

	hostTemplates := make(zabbix.TemplateIDs, len(templates))

	for i, t := range templates {
		hostTemplates[i] = zabbix.TemplateID{
			TemplateID: t.TemplateID,
		}
	}

	return hostTemplates, nil
}

func createHostObj(d *schema.ResourceData, api *zabbix.API) (*zabbix.Host, error) {
	host := zabbix.Host{
		Host:   d.Get("host").(string),
		Name:   d.Get("name").(string),
		Status: 0,
	}

	//0 is monitored, 1 - unmonitored host
	if !d.Get("monitored").(bool) {
		host.Status = 1
	}

	hostGroups, err := getHostGroups(d, api)

	if err != nil {
		return nil, err
	}

	host.GroupIds = hostGroups

	interfaces, err := getInterfaces(d)

	if err != nil {
		return nil, err
	}

	host.Interfaces = interfaces

	templates, err := getTemplates(d, api)

	if err != nil {
		return nil, err
	}

	host.TemplateIDs = templates

	macroMaps := d.Get("macro").([]any)
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
	host.Macros = macros
	return &host, nil
}

func resourceZabbixHostCreate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)

	host, err := createHostObj(d, api)

	if err != nil {
		return diag.FromErr(err)
	}

	hosts := zabbix.Hosts{*host}

	err = api.HostsCreate(hosts)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Created host id is %s", hosts[0].HostID)

	d.Set("host_id", hosts[0].HostID)
	d.SetId(hosts[0].HostID)
	readDiags := resourceZabbixHostRead(context, d, meta)
	return readDiags
}

func resourceZabbixHostRead(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	var errors TerraformErrors

	log.Printf("[DEBUG] Will read host with id %s", d.Get("host_id").(string))

	host := &zabbix.Host{HostID: d.Id()}
	err := api.ReadAPIObject(host)
	log.Printf("[DEBUG] Read host: %+v\n", host)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Host name is %s", host.Name)

	errors.addError(d.Set("host", host.Host))
	errors.addError(d.Set("name", host.Name))

	errors.addError(d.Set("monitored", host.Status == 0))

	var macros []any
	for _, macro := range host.Macros {
		macros = append(macros, map[string]any{
			"name":  macro.MacroName,
			"value": macro.Value,
		})
	}

	errors.addError(d.Set("macro", macros))
	params := zabbix.Params{
		"output": "extend",
		"hostids": []string{
			d.Id(),
		},
	}

	templates, err := api.TemplatesGet(params)

	if err != nil {
		return diag.FromErr(err)
	}

	templateNames := make([]string, len(templates))

	for i, t := range templates {
		templateNames[i] = t.Host
	}

	errors.addError(d.Set("templates", templateNames))

	groups, err := api.HostGroupsGet(params)

	if err != nil {
		return diag.FromErr(err)
	}

	groupNames := make([]string, len(groups))

	for i, g := range groups {
		groupNames[i] = g.Name
	}

	errors.addError(d.Set("groups", groupNames))

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
	errors.addError(d.Set("interfaces", interfaces))

	return errors.getDiagnostics()
}

func resourceZabbixHostUpdate(context context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)

	host, err := createHostObj(d, api)

	if err != nil {
		return diag.FromErr(err)
	}

	host.HostID = d.Id()

	////interfaces can't be updated, changes will trigger recreate
	////sending previous values will also fail the update
	//host.Interfaces = nil

	hosts := zabbix.Hosts{*host}

	err = api.HostsUpdate(hosts)

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] Created host id is %s", hosts[0].HostID)
	readDiags := resourceZabbixHostRead(context, d, meta)
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
