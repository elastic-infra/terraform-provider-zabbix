package zabbix

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
	"strings"

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

var SNMP3SecurityLevels = EnumMap{
	"noauth_nopriv": int(zabbix.SNMP3NoAuthNoPriv),
	"auth_nopriv":   int(zabbix.SNMP3AuthNoPriv),
	"auth_priv":     int(zabbix.SNMP3AuthPriv),
}

var SNMP3AuthProtocols = EnumMap{
	"MD5":    int(zabbix.SNMP3MD5Auth),
	"SHA1":   int(zabbix.SNMP3SHA1Auth),
	"SHA224": int(zabbix.SNMP3SHA224Auth),
	"SHA256": int(zabbix.SNMP3SHA256Auth),
	"SHA384": int(zabbix.SNMP3SHA384Auth),
	"SHA512": int(zabbix.SNMP3SHA512Auth),
}

var SNMP3PrivProtocol = EnumMap{
	"DES":     int(zabbix.SNMP3DESPriv),
	"AES128":  int(zabbix.SNMP3AES128Priv),
	"AES192":  int(zabbix.SNMP3AES192Priv),
	"AES256":  int(zabbix.SNMP3AES256Priv),
	"AES192C": int(zabbix.SNMP3AES192CPriv),
	"AES256C": int(zabbix.SNMP3AES256CPriv),
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
			"interfaces": {
				Type:     schema.TypeSet,
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
				Type:     schema.TypeSet,
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
			Optional: true,
			Default:  false,
		},
		"port": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "10050",
		},
		"type": {
			Type:             schema.TypeString,
			Optional:         true,
			Default:          "agent",
			ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(HostInterfaceTypes.types(), true)),
		},
		"interface_id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"snmp_config": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"version": {
						Type:     schema.TypeString,
						Required: true,
					},
					"bulk": {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
					},
					"community": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"snmpv3_config": {
						Type:     schema.TypeSet,
						Optional: true,
						MaxItems: 1,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"security_name": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"security_level": {
									Type:             schema.TypeString,
									Optional:         true,
									Computed:         true,
									ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(SNMP3SecurityLevels.types(), true)),
								},
								"auth_passphrase": {
									Type:     schema.TypeString,
									Optional: true,
								},
								"auth_protocol": {
									Type:             schema.TypeString,
									Optional:         true,
									Computed:         true,
									ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(SNMP3AuthProtocols.types(), true)),
								},
								"priv_protocol": {
									Type:             schema.TypeString,
									Optional:         true,
									Computed:         true,
									ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(SNMP3PrivProtocol.types(), true)),
								},
								"context_name": {
									Type:     schema.TypeString,
									Optional: true,
								},
							},
						},
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
	interfaces, err := generateInterfacesFromResourceData(data)
	if err != nil {
		return nil, err
	}
	host.Interfaces = interfaces
	host.TemplateIDs = generateTemplateIDsFromResourceData(data)
	host.Macros = generateHostMacrosFromResourceData(data)
	host.Tags = generateHostTagsFromResourceData(data)
	return &host, nil
}

func generateInterfacesFromResourceData(d *schema.ResourceData) ([]zabbix.HostInterface, error) {
	interfaceBlocks := d.Get("interfaces").(*schema.Set).List()
	interfaces := make(zabbix.HostInterfaces, len(interfaceBlocks))

	for i, interfaceBlock := range interfaceBlocks {
		interfaceMap := interfaceBlock.(map[string]any)
		interfaceType := interfaceMap["type"].(string)
		typeID, ok := HostInterfaceTypes[strings.ToLower(interfaceType)]
		if !ok {
			return nil, fmt.Errorf("%s isnt valid interface type", interfaceType)
		}
		ip := interfaceMap["ip"].(string)
		dns := interfaceMap["dns"].(string)
		interfaceId := interfaceMap["interface_id"].(string)
		if ip == "" && dns == "" {
			return nil, fmt.Errorf("atleast one of two dns or ip must be set")
		}
		useip := 1
		if ip == "" {
			useip = 0
		}
		main := 1
		if !interfaceMap["main"].(bool) {
			main = 0
		}
		interfaces[i] = zabbix.HostInterface{
			InterfaceID: interfaceId,
			IP:          ip,
			DNS:         dns,
			Main:        main,
			Port:        interfaceMap["port"].(string),
			Type:        zabbix.InterfaceType(typeID),
			UseIP:       useip,
		}

		SNMPConfigList := interfaceMap["snmp_config"].([]any)
		if typeID == int(zabbix.SNMPInterface) && len(SNMPConfigList) == 0 {
			return nil, fmt.Errorf("snmp_config block must be filled for %s interface type", HostInterfaceTypes.getStringType(typeID))
		} else if len(SNMPConfigList) > 0 {
			SNMPConfig := SNMPConfigList[0].(map[string]any)
			SNMPDetails, err := generateSNMPDetails(SNMPConfig)
			if err != nil {
				return nil, err
			}
			interfaces[i].Details = SNMPDetails
		}
	}
	return interfaces, nil
}

func generateSNMPDetails(SNMPConfig map[string]any) (*zabbix.SNMPDetails, error) {
	bulk := 0
	if SNMPConfig["bulk"].(bool) {
		bulk = 1
	}
	version := SNMPConfig["version"].(string)
	community := SNMPConfig["community"].(string)
	if (version == "1" || version == "2") && community == "" {
		return nil, fmt.Errorf("snmp_config.snmpv3_config.community must be defined for snmp version %s", version)
	}
	SNMPDetails := &zabbix.SNMPDetails{
		Version:   version,
		Community: community,
		Bulk:      bulk,
	}
	SNMP3ConfigList := SNMPConfig["snmpv3_config"].(*schema.Set).List()
	if len(SNMP3ConfigList) != 0 {
		SNMP3Config := SNMP3ConfigList[0].(map[string]any)
		SNMPDetails.SecurityName = SNMP3Config["security_name"].(string)
		SNMPDetails.SecurityLevel = SNMP3SecurityLevels[strings.ToLower(SNMP3Config["security_level"].(string))]
		SNMPDetails.AuthPassphrase = SNMP3Config["auth_passphrase"].(string)
		SNMPDetails.AuthProtocol = SNMP3AuthProtocols[strings.ToLower(SNMP3Config["auth_protocol"].(string))]
		SNMPDetails.PrivProtocol = SNMP3PrivProtocol[strings.ToLower(SNMP3Config["priv_protocol"].(string))]
		SNMPDetails.ContextName = SNMP3Config["context_name"].(string)
	}
	return SNMPDetails, nil
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
	macroMaps := data.Get("macro").(*schema.Set).List()
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
	errors.addError(readInterfacesIntoState(host.Interfaces, data))
	tags := make(map[string]string)
	for _, tag := range host.Tags {
		tags[tag.Tag] = tag.Value
	}
	errors.addError(data.Set("tags", tags))
	return errors.getDiagnostics()
}

func readInterfacesIntoState(hostInterfaces []zabbix.HostInterface, data *schema.ResourceData) error {
	interfaces := make([]map[string]any, len(hostInterfaces))
	for i, hostInterface := range hostInterfaces {
		interfaceType := HostInterfaceTypes.getStringType(int(hostInterface.Type))
		isMain := true
		if hostInterface.Main == 0 {
			isMain = false
		}
		interfaces[i] = map[string]any{
			"dns":          hostInterface.DNS,
			"ip":           hostInterface.IP,
			"main":         isMain,
			"port":         hostInterface.Port,
			"type":         interfaceType,
			"interface_id": hostInterface.InterfaceID,
		}
		SNMPDetailsMap, ok := hostInterface.Details.(map[string]any)
		SNMPDetailsBytes, err := json.Marshal(SNMPDetailsMap)
		if err != nil {
			return err
		}
		var SNMPDetails zabbix.SNMPDetails
		err = json.Unmarshal(SNMPDetailsBytes, &SNMPDetails)
		if err != nil {
			return err
		}
		if ok {
			bulk := false
			if SNMPDetails.Bulk != 0 {
				bulk = true
			}
			interfaces[i]["snmp_config"] = []map[string]any{{
				"version":   SNMPDetails.Version,
				"bulk":      bulk,
				"community": SNMPDetails.Community,
				"snmpv3_config": []map[string]any{{
					"security_name":   SNMPDetails.SecurityName,
					"security_level":  SNMP3SecurityLevels.getStringType(SNMPDetails.SecurityLevel),
					"auth_passphrase": SNMPDetails.AuthPassphrase,
					"auth_protocol":   SNMP3AuthProtocols.getStringType(SNMPDetails.AuthProtocol),
					"priv_protocol":   SNMP3PrivProtocol.getStringType(SNMPDetails.PrivProtocol),
					"context_name":    SNMPDetails.ContextName,
				}},
			}}
		}
	}
	return data.Set("interfaces", interfaces)
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
