package zabbix

import (
	"context"
	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceZabbixAuthenticationSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixAuthenticationCreate,
		ReadContext:   resourceZabbixAuthenticationRead,
		UpdateContext: resourceZabbixAuthenticationUpdate,
		DeleteContext: resourceZabbixAuthenticationDelete,
		Schema: map[string]*schema.Schema{
			"authentication_type": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Type Of Authentication to Use LDAP Must be configure before using this field",
			},
			"http_auth_enabled": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Enable Http Auth True/False",
			},
			"http_login_form": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"http_strip_domains": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"http_case_sensitive": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"ldap_configured": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"ldap_host": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ldap_port": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ldap_base_dn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ldap_search_attribute": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ldap_bind_dn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ldap_case_sensitive": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"ldap_bind_password": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"saml_auth_enabled": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "Enable SAML Auth True/False",
			},
			"saml_idp_entityid": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"saml_sso_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"saml_slo_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"saml_username_attribute": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"saml_sp_entityid": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"saml_nameid_format": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"saml_sign_messages": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"saml_sign_assertions": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"saml_sign_authn_requests": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"saml_sign_logout_requests": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"saml_sign_logout_responses": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"saml_encrypt_nameid": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"saml_encrypt_assertions": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"ldap_userdirectoryid": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"saml_case_sensitive": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"passwd_min_length": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"passwd_check_rules": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceZabbixAuthenticationDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	warning := diag.Diagnostic{
		Summary: "The authentication_settings resource is a singleton that can only be read and updated, " +
			"this will only delete the resource from terraform state",
		Severity: diag.Warning,
	}
	data.SetId("")
	return diag.Diagnostics{warning}
}

func resourceZabbixAuthenticationUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	authSettings := createAuthObjectFromResourceData(data)
	err := api.AuthSet(authSettings)
	if err != nil {
		return diag.FromErr(err)
	}
	readDiags := resourceZabbixAuthenticationRead(ctx, data, meta)
	return readDiags
}

func resourceZabbixAuthenticationCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	warning := diag.Diagnostic{
		Summary: "The authentication_settings resource is a singleton that can only be read and updated, " +
			"this will only read the resource into terraform state",
		Severity: diag.Warning,
	}
	readDiags := resourceZabbixAuthenticationRead(ctx, data, meta)
	data.SetId("authentication_settings")
	return append(readDiags, warning)
}

func resourceZabbixAuthenticationRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	authentication, err := api.AuthGet()
	if err != nil {
		return diag.FromErr(err)
	}
	errors.addError(data.Set("authentication_type", authentication.AuthenticationType))
	errors.addError(data.Set("http_auth_enabled", authentication.HttpAuthEnabled))
	errors.addError(data.Set("http_login_form", authentication.HttpLoginForm))
	errors.addError(data.Set("http_strip_domains", authentication.HttpStripDomains))
	errors.addError(data.Set("http_case_sensitive", authentication.HttpCaseSensitive))
	errors.addError(data.Set("ldap_configured", authentication.LdapConfigured))
	errors.addError(data.Set("ldap_case_sensitive", authentication.LdapCaseSensitive))
	//errors.addError(data.Set("ldap_userdirectoryid", authentication.LdapUserdirectoryid))
	errors.addError(data.Set("saml_auth_enabled", authentication.SamlAuthEnabled))
	errors.addError(data.Set("saml_idp_entityid", authentication.SamlIdpEntityid))
	errors.addError(data.Set("saml_sso_url", authentication.SamlSsoUrl))
	errors.addError(data.Set("saml_slo_url", authentication.SamlSloUrl))
	errors.addError(data.Set("saml_username_attribute", authentication.SamlUsernameAttribute))
	errors.addError(data.Set("saml_sp_entityid", authentication.SamlSpEntityid))
	errors.addError(data.Set("saml_nameid_format", authentication.SamlNameidFormat))
	errors.addError(data.Set("saml_sign_messages", authentication.SamlSignMessages))
	errors.addError(data.Set("saml_sign_assertions", authentication.SamlSignAssertions))
	errors.addError(data.Set("saml_sign_authn_requests", authentication.SamlSignAuthnRequests))
	errors.addError(data.Set("saml_sign_logout_requests", authentication.SamlSignLogoutRequests))
	errors.addError(data.Set("saml_sign_logout_responses", authentication.SamlSignLogoutResponses))
	errors.addError(data.Set("saml_encrypt_nameid", authentication.SamlEncryptNameid))
	errors.addError(data.Set("saml_encrypt_assertions", authentication.SamlEncryptAssertions))
	errors.addError(data.Set("saml_case_sensitive", authentication.SamlCaseSensitive))
	//errors.addError(data.Set("passwd_min_length", authentication.PasswdMinLength))
	//errors.addError(data.Set("passwd_check_rules", authentication.PasswdCheckRules))
	return errors.getDiagnostics()
}

func createAuthObjectFromResourceData(d *schema.ResourceData) *zabbix.AuthenticationSettings {
	auth := &zabbix.AuthenticationSettings{
		AuthenticationType: d.Get("authentication_type").(int),
		HttpAuthEnabled:    d.Get("http_auth_enabled").(int),
		HttpLoginForm:      d.Get("http_login_form").(int),
		HttpStripDomains:   d.Get("http_strip_domains").(string),
		HttpCaseSensitive:  d.Get("http_case_sensitive").(int),
		LdapConfigured:     d.Get("ldap_configured").(int),
		LdapCaseSensitive:  d.Get("ldap_case_sensitive").(int),
		//LdapUserdirectoryid:     d.Get("ldap_userdirectoryid").(int),
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
		//PasswdMinLength:         d.Get("passwd_min_length").(int),
		//PasswdCheckRules:        d.Get("passwd_check_rules").(int),
	}
	return auth
}
