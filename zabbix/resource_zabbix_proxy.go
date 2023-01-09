package zabbix

import (
	"context"
	"fmt"
	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strings"
)

func resourceZabbixProxy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixCreateProxy,
		ReadContext:   resourceZabbixReadProxy,
		DeleteContext: resourceZabbixDeleteProxy,
		UpdateContext: resourceZabbixUpdateProxy,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(ProxyStatusMap.types(), false)),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"last_access": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hosts": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"proxy_addresses": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ConflictsWith: []string{"interface"},
			},
			"interface": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				// For some reason the API isn't working when interface property is set for update method
				// So the proxy object has to be recreated when this property is to be changed
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"dns": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"port": {
							Type:     schema.TypeString,
							Required: true,
						},
						"use_ip": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceZabbixUpdateProxy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	proxy, err := createProxyObjectFromResourceData(data)
	if err != nil {
		return diag.FromErr(err)
	}
	err = api.UpdateAPIObject(proxy)
	errors.addError(err)
	readDiags := resourceZabbixReadProxy(ctx, data, meta)
	return append(errors.getDiagnostics(), readDiags...)
}

func resourceZabbixDeleteProxy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	proxy := &zabbix.Proxy{ProxyID: data.Id()}
	err := api.DeleteAPIObject(proxy)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId("")
	return nil
}

func resourceZabbixCreateProxy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	proxy, err := createProxyObjectFromResourceData(data)
	if err != nil {
		return diag.FromErr(err)
	}
	err = api.CreateAPIObject(proxy)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(proxy.GetID())
	readDiags := resourceZabbixReadProxy(ctx, data, meta)
	return readDiags
}

func resourceZabbixReadProxy(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	proxy := &zabbix.Proxy{ProxyID: data.Id()}
	err := api.ReadAPIObject(proxy)
	errors.addError(err)
	err = data.Set("name", proxy.Name)
	errors.addError(err)
	err = data.Set("status", ProxyStatusMap.getStringType(int(proxy.Status)))
	errors.addError(err)
	err = data.Set("description", proxy.Description)
	errors.addError(err)
	err = data.Set("last_access", proxy.LastAccess)
	errors.addError(err)
	var hostIds []string
	for _, host := range proxy.MonitoredHosts {
		hostIds = append(hostIds, host.HostID)
	}
	err = data.Set("hosts", hostIds)
	errors.addError(err)
	if proxy.ProxyAddress == "" {
		err = data.Set("proxy_addresses", []string{})
	} else {
		err = data.Set("proxy_addresses", strings.Split(proxy.ProxyAddress, ","))
	}
	errors.addError(err)
	useIp := false
	proxyInterface, ok := proxy.Interface.(zabbix.ProxyInterface)
	if ok && proxy.Interface != nil {
		if proxyInterface.UseIP != 0 {
			useIp = true
		}
		err = data.Set("interface", []map[string]any{{
			"id":     proxyInterface.InterfaceID,
			"dns":    proxyInterface.DNS,
			"ip":     proxyInterface.IP,
			"port":   proxyInterface.Port,
			"use_ip": useIp,
		}})
		errors.addError(err)
	}
	return errors.getDiagnostics()
}

var ProxyStatusMap = EnumMap{
	"ACTIVE":  int(zabbix.ActiveProxy),
	"PASSIVE": int(zabbix.PassiveProxy),
}

func createProxyObjectFromResourceData(data *schema.ResourceData) (proxy *zabbix.Proxy, err error) {
	proxyStatus := zabbix.ProxyStatus(ProxyStatusMap[data.Get("status").(string)])
	proxyInterface, err := createProxyInterfaceObjectFromResourceData(data)
	if err != nil {
		return
	}
	var monitoredHosts []zabbix.ProxyMonitoredHost
	for _, hostId := range data.Get("hosts").([]any) {
		monitoredHost := zabbix.ProxyMonitoredHost{HostID: hostId.(string)}
		monitoredHosts = append(monitoredHosts, monitoredHost)
	}
	proxy = &zabbix.Proxy{
		ProxyID:        data.Id(),
		Name:           data.Get("name").(string),
		Status:         proxyStatus,
		Description:    data.Get("description").(string),
		ProxyAddress:   getProxyAddress(data.Get("proxy_addresses").([]any)),
		LastAccess:     data.Get("last_access").(string),
		Interface:      proxyInterface,
		MonitoredHosts: monitoredHosts,
	}
	return
}

func createProxyInterfaceObjectFromResourceData(data *schema.ResourceData) (proxyInterface *zabbix.ProxyInterface, err error) {
	proxyInterfaceList := data.Get("interface").([]any)
	if len(proxyInterfaceList) == 0 {
		return
	}
	proxyInterfaceMap := proxyInterfaceList[0].(map[string]any)
	useIP := 0
	if proxyInterfaceMap["use_ip"].(bool) {
		useIP = 1
	}
	if proxyInterfaceMap["ip"] != "" && proxyInterfaceMap["dns"] != "" {
		err = fmt.Errorf("you can either set ip or dns value for the proxy interface attribute")
		return
	}
	proxyInterface = &zabbix.ProxyInterface{
		InterfaceID: proxyInterfaceMap["id"].(string),
		IP:          proxyInterfaceMap["ip"].(string),
		DNS:         proxyInterfaceMap["dns"].(string),
		Port:        proxyInterfaceMap["port"].(string),
		UseIP:       useIP,
	}
	return
}

func getProxyAddress(addressList []any) (addressString string) {
	var addressStringList []string
	for _, address := range addressList {
		addressStringList = append(addressStringList, address.(string))
	}
	return strings.Join(addressStringList, ",")
}
