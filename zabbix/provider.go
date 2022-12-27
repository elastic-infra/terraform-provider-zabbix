package zabbix

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider define the provider and his resources
func Provider() *schema.Provider {
	p := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"user": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ZABBIX_USER", nil),
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ZABBIX_PASSWORD", nil),
			},
			"server_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ZABBIX_SERVER_URL", nil),
			},
			"tls_insecure": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ZABBIX_TLS_INSECURE", nil),
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"zabbix_server":     dataSourceZabbixServer(),
			"zabbix_host_group": dataSourceZabbixHostGroup(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"zabbix_host":              resourceZabbixHost(),
			"zabbix_host_group":        resourceZabbixHostGroup(),
			"zabbix_item":              resourceZabbixItem(),
			"zabbix_trigger":           resourceZabbixTrigger(),
			"zabbix_template":          resourceZabbixTemplate(),
			"zabbix_template_link":     resourceZabbixTemplateLink(),
			"zabbix_lld_rule":          resourceZabbixLLDRule(),
			"zabbix_item_prototype":    resourceZabbixItemPrototype(),
			"zabbix_trigger_prototype": resourceZabbixTriggerPrototype(),
			"zabbix_action":            resourceZabbixAction(),
			"zabbix_role":              resourceZabbixRole(),
			"zabbix_valuemap":          resourceZabbixValueMap(),
			"zabbix_user_group":        resourceZabbixUserGroup(),
		},
	}

	p.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		terraformVersion := p.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(d, terraformVersion)
	}

	return p
}

func providerConfigure(d *schema.ResourceData, terraformVersion string) (interface{}, error) {
	api := zabbix.NewAPI(d.Get("server_url").(string))

	api.UserAgent = fmt.Sprintf("HashiCorp/1.0 Terraform/%s", terraformVersion)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: d.Get("tls_insecure").(bool),
			},
		},
	}

	if logging.IsDebugOrHigher() {
		client.Transport = logging.NewTransport("Zabbix", client.Transport)
	}

	api.SetClient(client)

	if _, err := api.Login(d.Get("user").(string), d.Get("password").(string)); err != nil {
		return nil, err
	}

	return api, nil
}

func getZabbixServerVersion(meta interface{}) string {
	api := meta.(*zabbix.API)
	v, err := api.Version()
	if err != nil {
		log.Printf("[WARN] Failed to get Zabbix Server version: %v\n", err)
		return ""
	}
	log.Printf("[DEBUG] Zabbix Server version is %s\n", v)

	return v
}

func isZabbixServerVersion34OrHigher(zabbixVersion string) bool {
	v1, _ := version.NewVersion(zabbixVersion)
	v2, _ := version.NewVersion("3.4.0")

	return v1.GreaterThanOrEqual(v2)
}

func getZabbixServerUnitDays(zabbixVersion string) string {
	if isZabbixServerVersion34OrHigher(zabbixVersion) {
		return "d"
	}
	return ""
}

func getZabbixServerUnitHours(zabbixVersion string) string {
	if isZabbixServerVersion34OrHigher(zabbixVersion) {
		return "h"
	}
	return ""
}

func getZabbixServerUnitMinutes(zabbixVersion string) string {
	if isZabbixServerVersion34OrHigher(zabbixVersion) {
		return "m"
	}
	return ""
}

func getZabbixServerUnitSeconds(zabbixVersion string) string {
	if isZabbixServerVersion34OrHigher(zabbixVersion) {
		return "s"
	}
	return ""
}

func getZabbixServerUnitWeeks(zabbixVersion string) string {
	if isZabbixServerVersion34OrHigher(zabbixVersion) {
		return "w"
	}
	return ""
}
