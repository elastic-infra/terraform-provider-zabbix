package zabbix

import (
	"context"
	"github.com/atypon/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func createMediaTypeFromSchema(data *schema.ResourceData) (mediaType *zabbix.MediaType) {
	var disabled int
	if data.Get("enabled").(bool) {
		disabled = 0
	} else {
		disabled = 1
	}
	mediaType = &zabbix.MediaType{
		MediaID:     data.Id(),
		MediaName:   data.Get("name").(string),
		MediaKind:   zabbix.ScriptMedia,
		Description: data.Get("description").(string),
		Disabled:    disabled,
	}
	return
}

func readCommonMediaTypeProperties(data *schema.ResourceData, mediaType *zabbix.MediaType) TerraformErrors {
	var errors TerraformErrors
	err := data.Set("name", mediaType.MediaName)
	errors.addError(err)
	err = data.Set("enabled", mediaType.Disabled == 0)
	errors.addError(err)
	err = data.Set("description", mediaType.Description)
	errors.addError(err)
	return errors
}

func resourceZabbixDeleteMediaType(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	mediaType := &zabbix.MediaType{MediaID: data.Id()}
	err := api.DeleteAPIObject(mediaType)
	errors.addError(err)
	data.SetId("")
	return errors.getDiagnostics()
}
