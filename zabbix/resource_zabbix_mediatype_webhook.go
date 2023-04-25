package zabbix

import (
	"context"
	"github.com/atypon/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceZabbixMediaTypeWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixCreateMediaTypeWebhook,
		ReadContext:   resourceZabbixReadMediaTypeWebhook,
		DeleteContext: resourceZabbixDeleteMediaType,
		UpdateContext: resourceZabbixUpdateMediaTypeWebhook,
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
			"script": {
				Type:     schema.TypeString,
				Required: true,
			},
			"timeout": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"parameter": {
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

func resourceZabbixUpdateMediaTypeWebhook(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	mediaType := createWebhookMediaTypeFromSchema(data)
	err := api.UpdateAPIObject(mediaType)
	errors.addError(err)
	readDiags := resourceZabbixReadMediaTypeWebhook(ctx, data, meta)
	return append(errors.getDiagnostics(), readDiags...)
}

func resourceZabbixReadMediaTypeWebhook(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	mediaType := &zabbix.MediaType{MediaID: data.Id()}
	err := api.ReadAPIObject(mediaType)
	errors.addError(err)
	readErrors := readCommonMediaTypeProperties(data, mediaType)
	errors.addFromTerraformErrors(readErrors)
	errors.addError(err)
	err = data.Set("script", mediaType.WebhookScript)
	errors.addError(err)
	err = data.Set("timeout", mediaType.WebhookTimeout)
	errors.addError(err)
	var parameters []any
	for _, param := range mediaType.WebhookParameters {
		parameters = append(parameters,
			map[string]any{
				"name":  param.Name,
				"value": param.Value,
			},
		)
	}
	return errors.getDiagnostics()
}

func resourceZabbixCreateMediaTypeWebhook(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	mediaType := createWebhookMediaTypeFromSchema(data)
	err := api.CreateAPIObject(mediaType)
	errors.addError(err)
	data.SetId(mediaType.GetID())
	readDiags := resourceZabbixReadMediaTypeWebhook(ctx, data, meta)
	return append(errors.getDiagnostics(), readDiags...)
}

func createWebhookMediaTypeFromSchema(data *schema.ResourceData) (mediaType *zabbix.MediaType) {
	mediaType = createMediaTypeFromSchema(data)
	mediaType.MediaKind = zabbix.WebhookMedia
	mediaType.WebhookScript = data.Get("script").(string)
	mediaType.WebhookTimeout = data.Get("timeout").(string)
	var webhookParameters []zabbix.WebhookParam
	for _, parameter := range data.Get("parameter").([]any) {
		parameterData := parameter.(map[string]any)
		webhookParam := zabbix.WebhookParam{
			Name:  parameterData["name"].(string),
			Value: parameterData["value"].(string),
		}
		webhookParameters = append(webhookParameters, webhookParam)
	}
	mediaType.WebhookParameters = webhookParameters
	return
}
