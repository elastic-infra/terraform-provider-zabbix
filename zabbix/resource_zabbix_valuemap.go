package zabbix

import (
	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

/* Add a converter from Type String to Map
Mapping match type. For type equal 0,1,2,3,4 value field cannot be empty, for type 5 value field should be empty.

Possible values:
0 - (default) exact match ;
1 - mapping will be applied if value is greater or equal1;
2 - mapping will be applied if value is less or equal1;
3 - mapping will be applied if value is in range (ranges are inclusive), allow to define multiple ranges separated by comma character1;
4 - mapping will be applied if value match regular expression2;
5 - default value, mapping will be applied if no other match were found.


*/
type ValueMapType struct {
	ValueMapID       string       `json:"valuemapid,omitempty"`
	ValueMapName     string       `json:"name"`
	ValueMapMappings mappingsList `json:"mappings"`
	ValueMapHostID   string       `json:"hostid"`
	ValueMapUUID     string       `json:"uuid,omitempty"`
}
type mappings struct {
	Type     int    `json:"type,omitempty"`
	Value    string `json:"value"`
	Newvalue string `json:"newvalue"`
}

type mappingsList []mappings

// ValueMaps is an array of ValueMapType
type ValueMaps []ValueMapType

func resourceZabbixValueMap() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixValueMapCreate,
		Read:   resourceZabbixValueMapRead,
		Exists: resourceZabbixValueMapExists,
		Update: resourceZabbixValueMapUpdate,
		Delete: resourceZabbixValueMapDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"uuid": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Universal unique identifier, used for linking imported value maps to already existing ones. Used only for value maps on templates. Auto-generated, if not given",
			},
			"valuemap_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"mapping": {
				Type:     schema.TypeList,
				Optional: true,
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
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceZabbixValueMapDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceZabbixValueMapUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceZabbixValueMapExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return true, nil
}

func resourceZabbixValueMapRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceZabbixValueMapCreate(d *schema.ResourceData, meta interface{}) error {
	valueMap := createValueMapObject(d)

	return createRetry(d, meta, createValueMap, *valueMap, resourceZabbixValueMapRead)

}

func createValueMap(valemap interface{}, api *zabbix.API) (id string, err error) {
	valueMaps := zabbix.ValueMaps{valemap.(zabbix.ValueMapType)}

	err = api.ValueMapCreate(valueMaps)
	if err != nil {
		return
	}
	id = valueMaps[0].ValueMapID
	return
}
func createValueMapObject(d *schema.ResourceData) *zabbix.ValueMapType {

	valueMap := zabbix.ValueMapType{
		ValueMapName:     d.Get("name").(string),
		ValueMapUUID:     d.Get("uuid").(string),
		ValueMapMappings: createValueMapMappingObject(d.Get("mapping").([]interface{})),
		ValueMapHostID:   d.Get("valuemap_id").(string),
	}

	return &valueMap
}

func createValueMapMappingObject(i []interface{}) (mappingsList zabbix.MappingsList) {
	for _, v := range i {
		m := v.(map[string]interface{})
		valueMap := zabbix.Mapping{
			Type:     0,
			Value:    m["value"].(string),
			Newvalue: m["new_value"].(string),
		}
		mappingsList = append(mappingsList, valueMap)
	}

	return mappingsList
}
