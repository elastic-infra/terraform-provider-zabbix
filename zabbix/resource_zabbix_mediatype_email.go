package zabbix

import (
	"context"
	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceZabbixMediaTypeEmail() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixCreateMediaTypeEmail,
		ReadContext:   resourceZabbixReadMediaTypeEmail,
		DeleteContext: resourceZabbixDeleteMediaType,
		UpdateContext: resourceZabbixUpdateMediaTypeEmail,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"smtp_auth_user": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"smtp_auth_password": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"smtp_from_email": {
				Type:     schema.TypeString,
				Required: true,
			},
			"smtp_helo": {
				Type:     schema.TypeString,
				Required: true,
			},
			"smtp_server": {
				Type:     schema.TypeString,
				Required: true,
			},
			"smtp_port": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceZabbixUpdateMediaTypeEmail(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	mediaType := createEmailMediaTypeFromSchema(data)
	err := api.UpdateAPIObject(mediaType)
	errors.addError(err)
	return errors.getDiagnostics()
}

func resourceZabbixCreateMediaTypeEmail(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	mediaType := createEmailMediaTypeFromSchema(data)
	err := api.CreateAPIObject(mediaType)
	errors.addError(err)
	data.SetId(mediaType.GetID())
	resourceZabbixReadMediaTypeEmail(ctx, data, meta)
	return errors.getDiagnostics()
}

func resourceZabbixReadMediaTypeEmail(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	mediaType := &zabbix.MediaType{MediaID: data.Id()}
	err := api.ReadAPIObject(mediaType)
	errors.addError(err)
	readErrors := readCommonMediaTypeProperties(data, mediaType)
	errors.addFromTerraformErrors(readErrors)
	errors.addError(err)
	err = data.Set("smtp_auth_user", mediaType.SMTPAuthUser)
	errors.addError(err)
	err = data.Set("smtp_auth_password", mediaType.SMTPAuthPassword)
	errors.addError(err)
	err = data.Set("smtp_from_email", mediaType.SMTPFromEmail)
	errors.addError(err)
	err = data.Set("smtp_helo", mediaType.SMTPHelo)
	errors.addError(err)
	err = data.Set("smtp_server", mediaType.SMTPServer)
	errors.addError(err)
	err = data.Set("smtp_port", mediaType.SMTPPort)
	errors.addError(err)
	return errors.getDiagnostics()
}

func createEmailMediaTypeFromSchema(data *schema.ResourceData) (mediaType *zabbix.MediaType) {
	mediaType = createMediaTypeFromSchema(data)
	mediaType.MediaKind = zabbix.EmailMedia
	mediaType.SMTPAuthUser = data.Get("smtp_auth_user").(string)
	mediaType.SMTPAuthPassword = data.Get("smtp_auth_password").(string)
	mediaType.SMTPFromEmail = data.Get("smtp_from_email").(string)
	mediaType.SMTPHelo = data.Get("smtp_helo").(string)
	mediaType.SMTPServer = data.Get("smtp_server").(string)
	mediaType.SMTPPort = data.Get("smtp_port").(string)
	return
}
