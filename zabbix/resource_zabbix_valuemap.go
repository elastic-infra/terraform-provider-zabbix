package zabbix

import (
	"context"
	"github.com/atypon/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"log"
)

var mappingTypeMap = EnumMap{
	"exact_match":      int(zabbix.ExactMatchMapping),
	"greater_or_equal": int(zabbix.GreaterOrEqualMapping),
	"less_or_equal":    int(zabbix.LessOrEqualMapping),
	"regex_match":      int(zabbix.RegexMatchMapping),
	"default_match":    int(zabbix.DefaultValueMapping),
}

func resourceZabbixValueMap() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixValueMapCreate,
		ReadContext:   resourceZabbixValueMapRead,
		UpdateContext: resourceZabbixValueMapUpdate,
		DeleteContext: resourceZabbixValueMapDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"uuid": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Universal unique identifier, used for linking imported value maps to already existing ones. Used only for value maps on templates. Auto-generated, if not given",
			},
			"host_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mapping": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"new_value": {
							Type:     schema.TypeString,
							Required: true,
						},
						"type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(mappingTypeMap.types(), false)),
						},
					},
				},
			},
		},
	}
}

func resourceZabbixValueMapDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	valueMap := &zabbix.ValueMap{ValueMapID: data.Id()}
	err := api.DeleteAPIObject(valueMap)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId("")
	return nil
}

func resourceZabbixValueMapUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	valueMap := createValueMapObjectFromResourceData(data)
	/*
		HostID & UUID Should both be nulls on update
	*/
	valueMap.HostID = ""
	valueMap.UUID = ""
	err := api.UpdateAPIObject(valueMap)
	if err != nil {
		return diag.FromErr(err)
	}
	readDiags := resourceZabbixValueMapRead(ctx, data, meta)
	return readDiags
}

func resourceZabbixValueMapRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	valueMap := zabbix.ValueMap{ValueMapID: data.Id()}
	err := api.ReadAPIObject(&valueMap)
	errors.addError(err)
	err = data.Set("name", valueMap.Name)
	errors.addError(err)
	err = data.Set("uuid", valueMap.UUID)
	errors.addError(err)
	err = data.Set("host_id", valueMap.HostID)
	errors.addError(err)
	log.Printf("Read valuemap object: %+v\n", valueMap)
	var mappings []map[string]any
	for _, mapping := range valueMap.Mappings {
		mappingMap := map[string]any{
			"value":     mapping.Value,
			"new_value": mapping.Newvalue,
			"type":      mappingTypeMap.getStringType(int(mapping.Type)),
		}
		mappings = append(mappings, mappingMap)
	}
	err = data.Set("mapping", mappings)
	errors.addError(err)
	return errors.getDiagnostics()
}

func resourceZabbixValueMapCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	valueMap := createValueMapObjectFromResourceData(data)
	err := api.CreateAPIObject(valueMap)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(valueMap.GetID())
	readDiags := resourceZabbixValueMapRead(ctx, data, meta)
	return readDiags
}

func createValueMapObjectFromResourceData(d *schema.ResourceData) *zabbix.ValueMap {
	valueMap := zabbix.ValueMap{
		ValueMapID: d.Id(),
		Name:       d.Get("name").(string),
		UUID:       d.Get("uuid").(string),
		Mappings:   createValueMapMappingObject(d.Get("mapping").([]interface{})),
		HostID:     d.Get("host_id").(string),
	}
	return &valueMap
}

func createValueMapMappingObject(i []interface{}) (mappingsList []zabbix.ValueMapMapping) {
	for _, v := range i {
		m := v.(map[string]any)
		valueMap := zabbix.ValueMapMapping{
			Type:     zabbix.ValueMapMappingType(mappingTypeMap[m["type"].(string)]),
			Value:    m["value"].(string),
			Newvalue: m["new_value"].(string),
		}
		mappingsList = append(mappingsList, valueMap)
	}
	return mappingsList
}
