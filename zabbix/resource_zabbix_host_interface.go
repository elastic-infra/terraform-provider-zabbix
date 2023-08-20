package zabbix

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/atypon/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
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

func resourceZabbixHostInterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixHostInterfaceCreate,
		ReadContext:   resourceZabbixHostInterfaceRead,
		UpdateContext: resourceZabbixHostInterfaceUpdate,
		DeleteContext: resourceZabbixHostInterfaceDelete,
		Schema: map[string]*schema.Schema{
			"host_id": {
				Type:     schema.TypeString,
				Required: true,
			},
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
			//"interface_id": {
			//	Type:     schema.TypeString,
			//	Computed: true,
			//},
			"snmp_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				//DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				//	return true
				//},
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
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Computed: true,
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
										Type:      schema.TypeString,
										Optional:  true,
										Sensitive: true,
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceZabbixHostInterfaceUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	hostInterface, err := generateInterfaceObjectFromResourceData(data)
	if err != nil {
		return diag.FromErr(err)
	}
	err = api.UpdateAPIObject(hostInterface)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceZabbixHostInterfaceCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	hostInterface, err := generateInterfaceObjectFromResourceData(data)
	if err != nil {
		return diag.FromErr(err)
	}
	err = api.CreateAPIObject(hostInterface)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(hostInterface.GetID())
	return nil
}

func generateInterfaceObjectFromResourceData(d *schema.ResourceData) (*zabbix.HostInterface, error) {
	interfaceType := d.Get("type").(string)
	typeID, ok := HostInterfaceTypes[strings.ToLower(interfaceType)]
	if !ok {
		return nil, fmt.Errorf("%s isnt valid interface type", interfaceType)
	}
	ip := d.Get("ip").(string)
	dns := d.Get("dns").(string)
	if ip == "" && dns == "" {
		return nil, fmt.Errorf("atleast one of two dns or ip must be set")
	}
	useIP := 1
	if ip == "" {
		useIP = 0
	}
	main := 1
	if !d.Get("main").(bool) {
		main = 0
	}
	hostInterface := &zabbix.HostInterface{
		InterfaceID: d.Id(),
		HostID:      d.Get("host_id").(string),
		IP:          ip,
		DNS:         dns,
		Main:        main,
		Port:        d.Get("port").(string),
		Type:        zabbix.InterfaceType(typeID),
		UseIP:       useIP,
	}

	SNMPConfigList := d.Get("snmp_config").([]any)
	if typeID == int(zabbix.SNMPInterface) && len(SNMPConfigList) == 0 {
		return nil, fmt.Errorf("snmp_config block must be filled for %s interface type", HostInterfaceTypes.getStringType(typeID))
	} else if len(SNMPConfigList) > 0 {
		SNMPConfig := SNMPConfigList[0].(map[string]any)
		SNMPDetails, err := generateSNMPDetails(SNMPConfig)
		if err != nil {
			return nil, err
		}
		hostInterface.Details = SNMPDetails
	}
	//} else {
	//	hostInterface.Details = []string{}
	//}
	return hostInterface, nil
}

func generateSNMPDetails(SNMPConfig map[string]any) (*zabbix.SNMPDetails, error) {
	bulk := 0
	if SNMPConfig["bulk"].(bool) {
		bulk = 1
	}
	version := SNMPConfig["version"].(string)
	community := SNMPConfig["community"].(string)
	if (version == "1" || version == "2") && community == "" {
		return nil, fmt.Errorf("snmp_config.community must be defined for snmp version %s", version)
	}
	SNMPDetails := &zabbix.SNMPDetails{
		Version:   version,
		Community: community,
		Bulk:      bulk,
	}
	SNMP3ConfigList := SNMPConfig["snmpv3_config"].([]any) //(*schema.Set).List()
	if len(SNMP3ConfigList) != 0 && SNMP3ConfigList[0] != nil {
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

func resourceZabbixHostInterfaceDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	hostInterface := &zabbix.HostInterface{InterfaceID: data.Id()}
	err := api.DeleteAPIObject(hostInterface)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId("")
	return nil
}

func resourceZabbixHostInterfaceRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	hostInterface := &zabbix.HostInterface{InterfaceID: data.Id()}
	err := api.ReadAPIObject(hostInterface)
	if err != nil {
		return diag.FromErr(err)
	}
	var errors TerraformErrors
	isMain := true
	if hostInterface.Main == 0 {
		isMain = false
	}
	interfaceType := HostInterfaceTypes.getStringType(int(hostInterface.Type))
	errors.addError(data.Set("dns", hostInterface.DNS))
	errors.addError(data.Set("ip", hostInterface.IP))
	errors.addError(data.Set("main", isMain))
	errors.addError(data.Set("port", hostInterface.Port))
	errors.addError(data.Set("type", interfaceType))
	SNMPDetailsMap, ok := hostInterface.Details.(map[string]any)
	SNMPDetailsBytes, err := json.Marshal(SNMPDetailsMap)
	errors.addError(err)
	var SNMPDetails zabbix.SNMPDetails
	err = json.Unmarshal(SNMPDetailsBytes, &SNMPDetails)
	errors.addError(err)
	if ok {
		bulk := false
		if SNMPDetails.Bulk != 0 {
			bulk = true
		}
		SNMPConfig := []any{map[string]any{
			"version":   SNMPDetails.Version,
			"bulk":      bulk,
			"community": SNMPDetails.Community,
			"snmpv3_config": []any{map[string]any{
				"security_name":   SNMPDetails.SecurityName,
				"security_level":  SNMP3SecurityLevels.getStringType(SNMPDetails.SecurityLevel),
				"auth_passphrase": SNMPDetails.AuthPassphrase,
				"auth_protocol":   SNMP3AuthProtocols.getStringType(SNMPDetails.AuthProtocol),
				"priv_protocol":   SNMP3PrivProtocol.getStringType(SNMPDetails.PrivProtocol),
				"context_name":    SNMPDetails.ContextName,
			}},
		}}
		errors.addError(data.Set("snmp_config", SNMPConfig))
	}
	return errors.getDiagnostics()
}
