package zabbix

import (
	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"log"
)

func resourceZabbixAuth() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixAuthCreate,
		Read:   resourceZabbixAuthRead,
		//Exists: resourceZabbixAuthExists,
		Update: resourceZabbixAuthUpdate,
		Delete: resourceZabbixAuthDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"authentication_type": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Type Of Authentication to Use LDAP Must be configure before using this field",
			},
			"http_auth_enabled": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Enable Http Auth True/False",
			},
			"http_login_form": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"http_strip_domains": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"http_case_sensitive": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ldap_configured": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ldap_host": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ldap_port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ldap_base_dn": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ldap_search_attribute": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ldap_bind_dn": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ldap_case_sensitive": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ldap_bind_password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"saml_auth_enabled": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Enable SAML Auth True/False",
			},
			"saml_idp_entityid": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"saml_sso_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"saml_slo_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"saml_username_attribute": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"saml_sp_entityid": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"saml_nameid_format": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"saml_sign_messages": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"saml_sign_assertions": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"saml_sign_authn_requests": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"saml_sign_logout_requests": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"saml_sign_logout_responses": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"saml_encrypt_nameid": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"saml_encrypt_assertions": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"ldap_userdirectoryid": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"saml_case_sensitive": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"passwd_min_length": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
			},
			"passwd_check_rules": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceZabbixAuthCreate(d *schema.ResourceData, meta interface{}) error {
	auth := createAuthObject(d)
	createAuth(auth, meta)
	log.Printf("[DEBUG] Updated Auth Settings")
	d.Set("auth_id", "1")
	d.SetId("1")
	return nil
}

func createAuthObject(d *schema.ResourceData) zabbix.AuthPrototype {

	auth := zabbix.AuthPrototype{
		AuthenticationType:      d.Get("authentication_type").(int),
		HttpAuthEnabled:         d.Get("http_auth_enabled").(int),
		HttpLoginForm:           d.Get("http_login_form").(int),
		HttpStripDomains:        d.Get("http_strip_domains").(string),
		HttpCaseSensitive:       d.Get("http_case_sensitive").(int),
		LdapConfigured:          d.Get("ldap_configured").(int),
		LdapCaseSensitive:       d.Get("ldap_case_sensitive").(int),
		LdapUserdirectoryid:     d.Get("ldap_userdirectoryid").(int),
		SamlAuthEnabled:         d.Get("saml_auth_enabled").(int),
		SamlIdpEntityid:         d.Get("saml_idp_entityid").(string),
		SamlSsoUrl:              d.Get("saml_sso_url").(string),
		SamlSloUrl:              d.Get("saml_slo_url").(string),
		SamlUsernameAttribute:   d.Get("saml_username_attribute").(string),
		SamlSpEntityid:          d.Get("saml_sp_entityid").(string),
		SamlNameidFormat:        d.Get("saml_nameid_format").(string),
		SamlSignMessages:        d.Get("saml_sign_messages").(int),
		SamlSignAssertions:      d.Get("saml_sign_assertions").(int),
		SamlSignAuthnRequests:   d.Get("saml_sign_authn_requests").(int),
		SamlSignLogoutRequests:  d.Get("saml_sign_logout_requests").(int),
		SamlSignLogoutResponses: d.Get("saml_sign_logout_responses").(int),
		SamlEncryptNameid:       d.Get("saml_encrypt_nameid").(int),
		SamlEncryptAssertions:   d.Get("saml_encrypt_assertions").(int),
		SamlCaseSensitive:       d.Get("saml_case_sensitive").(int),
		PasswdMinLength:         d.Get("passwd_min_length").(int),
		PasswdCheckRules:        d.Get("passwd_check_rules").(string),
	}
	return auth
}

func resourceZabbixAuthRead(d *schema.ResourceData, meta interface{}) error {
	//api := meta.(*zabbix.API)
	//params := zabbix.Params{
	//	"output": "extend",
	//}
	//
	//res, err := api.AuthGet(params)
	//
	//if err != nil {
	//	return err
	//}
	//
	//log.Printf("inside read")
	//
	//if err != nil {
	//	log.Printf("inside read 1 ")
	//	return err
	//}
	//
	//err2 := d.Set("saml_auth_enabled", res.ID)
	//if err2 != nil {
	//	log.Printf("inside read 2")
	//	return err2
	//}
	////d.Set("host_id", auth.SAMLIDP)
	////d.Set("interface_id", auth.InterfaceID)
	////d.Set("key", auth.Key)
	////d.Set("name", auth.Name)
	////d.Set("type", auth.Type)
	////d.Set("value_type", auth.ValueType)
	////d.Set("data_type", auth.DataType)
	////d.Set("delta", auth.Delta)
	////d.Set("description", auth.Description)
	//
	////log.Printf("[DEBUG] Item name is %s\n", item.Name)
	return nil
}

func resourceZabbixAuthExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	//api := meta.(*zabbix.API)
	//
	//_, err := api.ItemGetByID(d.Id())
	//if err != nil {
	//	if strings.Contains(err.Error(), "Expected exactly one result") {
	//		log.Printf("[DEBUG] Item with id %s doesn't exist", d.Id())
	//		return false, nil
	//	}
	//	return false, err
	//}
	return true, nil
}

func resourceZabbixAuthUpdate(d *schema.ResourceData, meta interface{}) error {
	/*
		read the object , refill with changes and apply to api
	*/
	return nil
}

func resourceZabbixAuthDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func createAuth(auth zabbix.AuthPrototype, meta interface{}) (id string, err error) {

	api := meta.(*zabbix.API)
	_, err = api.AuthSet(auth)

	if err != nil {
		log.Printf("Error in create Auth")
		return
	}

	return "1", nil
}
