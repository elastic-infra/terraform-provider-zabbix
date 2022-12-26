package zabbix

import (
	"fmt"
	"log"
	"strings"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var ItemTypeInventoryMap = map[string]int{
	"Zabbix agent":          0,
	"Zabbix trapper":        2,
	"Simple check":          3,
	"Zabbix internal":       5,
	"Zabbix agent (active)": 7,
	"Web item":              9,
	"External check":        10,
	"Database monitor":      11,
	"IPMI agent":            12,
	"SSH agent":             13,
	"Telnet agent":          14,
	"Calculated":            15,
	"JMX agent":             16,
	"SNMP trap":             17,
	"Dependent item":        18,
	"HTTP agent":            19,
	"SNMP_AGENT":            20,
	"Script":                21,
}

var ValueTypeInventoryMap = map[string]int{
	"float":     0,
	"character": 1,
	"log":       2,
	"unsigned":  3,
	"text":      4,
}

func resourceZabbixItem() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixItemCreate,
		Read:   resourceZabbixItemRead,
		Exists: resourceZabbixItemExists,
		Update: resourceZabbixItemUpdate,
		Delete: resourceZabbixItemDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"delay": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"host_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the host or template that the item belongs to.",
			},
			"interface_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "0",
			},
			"key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Item key.",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the item.",
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  0,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					value := ItemTypeInventoryMap[val.(string)]
					if value < 0 || value > 21 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 16 inclusive, got %d", key, value))
					}
					return
				},
			},
			"value_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  0,
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					value := ValueTypeInventoryMap[val.(string)]
					if value < 0 || value > 4 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 4 inclusive, got %d", key, value))
					}
					return
				},
			},
			"data_type": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Data type of the item (Removed in Zabbix 3.4).",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 || v > 3 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 3 inclusive, got %d", key, v))
					}
					return
				},
			},
			"delta": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Value that will be stored (Removed in Zabbix 3.4).",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					v := val.(int)
					if v < 0 || v > 2 {
						errs = append(errs, fmt.Errorf("%q, must be between 0 and 2 inclusive, got %d", key, v))
					}
					return
				},
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the item.",
				Default:     "",
			},
			"history": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Number of days to keep item's history data. From 3.4 version, string is required instead of integer. Default: 90 (90d for 3.4+).",
			},
			"trends": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Optional:    true,
				Description: "Number of days to keep item's trends data. From 3.4 version, string is required instead of interger. Default: 365 (365d for 3.4+).",
			},
			"trapper_host": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allowed hosts. Used only by trapper items.",
			},
			"units": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SNMP OID , Used only with SNMP.",
			},
			"snmp_oid": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "SNMP OID , Used only with SNMP.",
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tag": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true},
					},
				},
			},
			"valuemap_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Value Map of Item",
			},
			"preprocessor": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Item Preprocessors",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"step": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Optional: true},
								},
							},
						},
					},
				},
			},
			//"preprocessor": {
			//	Type:     schema.TypeList,
			//	Optional: true,
			//	Elem: &schema.Resource{
			//		Schema: map[string]*schema.Schema{
			//			"type": {
			//				Type:     schema.TypeString,
			//				Required: true,
			//			},
			//			"value": {
			//				Type:     schema.TypeString,
			//				Optional: true},
			//		},
			//	},
			//},
		},
	}
}

func createItemObject(d *schema.ResourceData) *zabbix.Item {

	item := zabbix.Item{
		Delay:        d.Get("delay").(string),
		HostID:       d.Get("host_id").(string),
		InterfaceID:  d.Get("interface_id").(string),
		Key:          d.Get("key").(string),
		Name:         d.Get("name").(string),
		Type:         zabbix.ItemType(ItemTypeInventoryMap[d.Get("type").(string)]),
		ValueType:    zabbix.ValueType(ValueTypeInventoryMap[d.Get("value_type").(string)]),
		DataType:     zabbix.DataType(d.Get("data_type").(int)),
		Delta:        zabbix.DeltaType(d.Get("delta").(int)),
		Description:  d.Get("description").(string),
		History:      d.Get("history").(string),
		Trends:       d.Get("trends").(string),
		TrapperHosts: d.Get("trapper_host").(string),
		SnmpOid:      d.Get("snmp_oid").(string),
		Units:        d.Get("units").(string),
		Tags:         createItemTagsObject(d.Get("tags").([]interface{})),
		ValueMapID:   d.Get("valuemap_id").(string), // get its ID or create a new ValueMap
		Preprocessor: createPreProcessorObject(d.Get("preprocessor").([]interface{})),
	}

	return &item
}

func createPreProcessorObject(lst []interface{}) (PreprocessorList zabbix.PreprocessorList) {
	for _, v := range lst {
		stepMap := v.(map[string]interface{})["step"].([]interface{})
		for _, v := range stepMap {
			key := v.(map[string]interface{})["type"].(string)
			value := v.(map[string]interface{})["value"].(string)
			println(key, value)
		}
	}

	return PreprocessorList
}

func resourceZabbixItemCreate(d *schema.ResourceData, meta interface{}) error {
	item := createItemObject(d)

	return createRetry(d, meta, createItem, *item, resourceZabbixItemRead)
}

func resourceZabbixItemRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	item, err := api.ItemGetByID(d.Id())
	if err != nil {
		return err
	}

	d.Set("delay", item.Delay)
	d.Set("host_id", item.HostID)
	d.Set("interface_id", item.InterfaceID)
	d.Set("key", item.Key)
	d.Set("name", item.Name)
	d.Set("type", item.Type)
	d.Set("value_type", item.ValueType)
	d.Set("data_type", item.DataType)
	d.Set("delta", item.Delta)
	d.Set("description", item.Description)
	d.Set("history", item.History)
	d.Set("trends", item.Trends)
	d.Set("trapper_host", item.TrapperHosts)
	d.Set("snmp_oid", item.SnmpOid)
	d.Set("units", item.Units)
	d.Set("tags", createItemTagsObject(d.Get("tags").([]interface{})))
	log.Printf("[DEBUG] Item name is %s\n", item.Name)
	return nil
}

func resourceZabbixItemExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	api := meta.(*zabbix.API)

	_, err := api.ItemGetByID(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "Expected exactly one result") {
			log.Printf("[DEBUG] Item with id %s doesn't exist", d.Id())
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func resourceZabbixItemUpdate(d *schema.ResourceData, meta interface{}) error {
	item := createItemObject(d)

	item.ItemID = d.Id()
	return createRetry(d, meta, updateItem, *item, resourceZabbixItemRead)

}

func resourceZabbixItemDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	return deleteRetry(d.Id(), getItemParentID, api.ItemsDeleteIDs, api)
}

func getItemParentID(api *zabbix.API, id string) (string, error) {
	items, err := api.ItemsGet(zabbix.Params{
		"output":      "extend",
		"selectHosts": "extend",
		"itemids":     id,
	})
	if err != nil {
		return "", fmt.Errorf("%s, with item %s", err.Error(), id)
	}
	if len(items) != 1 {
		return "", fmt.Errorf("Expected one item and got %d items", len(items))
	}
	if len(items[0].ItemParent) != 1 {
		return "", fmt.Errorf("Expected one parent for item %s and got %d", id, len(items[0].ItemParent))
	}
	return items[0].ItemParent[0].HostID, nil
}

func createItem(item interface{}, api *zabbix.API) (id string, err error) {
	items := zabbix.Items{item.(zabbix.Item)}

	err = api.ItemsCreate(items)
	if err != nil {
		return
	}
	id = items[0].ItemID
	return
}

func updateItem(item interface{}, api *zabbix.API) (id string, err error) {
	items := zabbix.Items{item.(zabbix.Item)}

	err = api.ItemsUpdate(items)
	if err != nil {
		return
	}
	id = items[0].ItemID
	return
}

func readItemTagsObject(tags zabbix.TagsList) (lst []interface{}, err error) {
	for _, v := range tags {
		m := map[string]interface{}{}
		m["tag"] = v.Tag
		m["value"] = v.Value
		lst = append(lst, m)
	}
	return
}

func createItemTagsObject(lst []interface{}) (tags zabbix.TagsList) {
	for _, v := range lst {
		m := v.(map[string]interface{})

		tag := zabbix.Tag{
			Tag:   m["tag"].(string),
			Value: m["value"].(string),
		}
		tags = append(tags, tag)
	}

	return
}
