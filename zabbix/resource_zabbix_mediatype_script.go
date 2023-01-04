package zabbix

import (
	"context"
	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strings"
)

func resourceZabbixMediaTypeScript() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixCreateMediaTypeScript,
		ReadContext:   resourceZabbixReadMediaTypeScript,
		DeleteContext: resourceZabbixDeleteMediaType,
		UpdateContext: resourceZabbixUpdateMediaTypeScript,
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
			"exec_path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"exec_parameters": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
			},
		},
	}
}

func resourceZabbixUpdateMediaTypeScript(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	mediaType := createScriptMediaTypeFromSchema(data)
	err := api.UpdateAPIObject(mediaType)
	errors.addError(err)
	return errors.getDiagnostics()
}

func resourceZabbixReadMediaTypeScript(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	mediaType := &zabbix.MediaType{MediaID: data.Id()}
	err := api.ReadAPIObject(mediaType)
	errors.addError(err)
	readErrors := readCommonMediaTypeProperties(data, mediaType)
	errors.addFromTerraformErrors(readErrors)
	errors.addError(err)
	err = data.Set("exec_path", mediaType.ScriptExecPath)
	errors.addError(err)
	var execParams []string
	for _, param := range strings.Split(mediaType.ScriptParams, "\n") {
		if param != "" {
			execParams = append(execParams, param)
		}
	}
	err = data.Set("exec_parameters", execParams)
	errors.addError(err)
	return errors.getDiagnostics()
}

func resourceZabbixCreateMediaTypeScript(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	mediaType := createScriptMediaTypeFromSchema(data)
	err := api.CreateAPIObject(mediaType)
	errors.addError(err)
	data.SetId(mediaType.GetID())
	resourceZabbixReadMediaTypeScript(ctx, data, meta)
	return errors.getDiagnostics()
}

func createScriptMediaTypeFromSchema(data *schema.ResourceData) (mediaType *zabbix.MediaType) {
	mediaType = createMediaTypeFromSchema(data)
	var scriptParams []string
	for _, param := range data.Get("exec_parameters").([]any) {
		scriptParams = append(scriptParams, param.(string))
	}
	mediaType.MediaKind = zabbix.ScriptMedia
	mediaType.ScriptExecPath = data.Get("exec_path").(string)
	mediaType.ScriptParams = strings.Join(scriptParams, "\n") + "\n"
	return
}
