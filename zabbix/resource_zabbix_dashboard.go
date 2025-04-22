package zabbix

import (
	"fmt"
	"log"
	"strings"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceZabbixDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixDashboardCreate,
		Read:   resourceZabbixDashboardRead,
		Exists: resourceZabbixDashboardExists,
		Update: resourceZabbixDashboardUpdate,
		Delete: resourceZabbixDashboardDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the dashboard.",
			},
			"display_period": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     30,
				Description: "Page display period (in seconds).",
			},
			"auto_start": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Auto start slideshow. 0 - no, 1 - yes.",
			},
			"private": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Dashboard private state. 0 - no, 1 - yes.",
			},
			"widgets": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Type of the dashboard widget.",
						},
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Name of the widget.",
						},
						"x": &schema.Schema{
							Type:        schema.TypeInt,
							Required:    true,
							Description: "X position of the widget.",
						},
						"y": &schema.Schema{
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Y position of the widget.",
						},
						"width": &schema.Schema{
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Width of the widget.",
						},
						"height": &schema.Schema{
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Height of the widget.",
						},
						"graph_ids": &schema.Schema{
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
							Description: "List of graph IDs to display in the widget.",
						},
						"item_ids": &schema.Schema{
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Optional:    true,
							Description: "List of item IDs to display in the widget.",
						},
					},
				},
			},
		},
	}
}

func createDashboardWidgets(d *schema.ResourceData) (Widgets, error) {
	widgets := make(Widgets, 0)
	terraformWidgets := d.Get("widgets").([]interface{})
	
	if len(terraformWidgets) == 0 {
		return widgets, nil
	}

	for _, terraformWidget := range terraformWidgets {
		widget := terraformWidget.(map[string]interface{})
		
		widgetObj := Widget{
			Type:   widget["type"].(string),
			Name:   widget["name"].(string),
			X:      widget["x"].(int),
			Y:      widget["y"].(int),
			Width:  widget["width"].(int),
			Height: widget["height"].(int),
		}

		// Handle graph IDs if present
		if graphIds, ok := widget["graph_ids"].([]interface{}); ok && len(graphIds) > 0 {
			for _, graphId := range graphIds {
				widgetObj.Fields = append(widgetObj.Fields, WidgetField{
					Type:  "0",
					Name:  "graphid",
					Value: graphId.(string),
				})
			}
		}

		// Handle item IDs if present
		if itemIds, ok := widget["item_ids"].([]interface{}); ok && len(itemIds) > 0 {
			for _, itemId := range itemIds {
				widgetObj.Fields = append(widgetObj.Fields, WidgetField{
					Type:  "0",
					Name:  "itemid",
					Value: itemId.(string),
				})
			}
		}

		widgets = append(widgets, widgetObj)
	}

	return widgets, nil
}

func createDashboardObj(d *schema.ResourceData) (*Dashboard, error) {
	dashboard := Dashboard{
		Name:          d.Get("name").(string),
		DisplayPeriod: d.Get("display_period").(int),
		AutoStart:     d.Get("auto_start").(int),
		Private:       d.Get("private").(int),
	}

	widgets, err := createDashboardWidgets(d)
	if err != nil {
		return nil, err
	}
	
	// Wrap widgets in a single page
	if len(widgets) > 0 {
		dashboard.Pages = []DashboardPage{
			{
				// You might want to set a default page name or make it configurable
				Name:    "Page 1", 
				Widgets: widgets,
			},
		}
	} else {
		// Ensure Pages is at least an empty slice if required by API, 
		// or handle as per API specifics if no widgets are defined.
		// For now, setting to an empty slice.
		// API requires at least one page. Create a default page.
		dashboard.Pages = []DashboardPage{
			{
				Name:    "Page 1", // Default page name
				Widgets: make(Widgets, 0), // Empty widgets for this page
			},
		}
	}

	return &dashboard, nil
}

func resourceZabbixDashboardCreate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	dashboard, err := createDashboardObj(d)
	if err != nil {
		return err
	}

	dashboards := Dashboards{*dashboard}
	err = DashboardsCreate(api, dashboards)
	if err != nil {
		return err
	}

	d.SetId(dashboards[0].DashboardID)
	return resourceZabbixDashboardRead(d, meta)
}

func resourceZabbixDashboardRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	params := zabbix.Params{
		"dashboardids": d.Id(),
		"output":       "extend",
	}
	dashboards, err := DashboardsGet(api, params)
	if err != nil {
		return err
	}
	if len(dashboards) != 1 {
		return fmt.Errorf("Expected one dashboard with id %s and got %d dashboards", d.Id(), len(dashboards))
	}

	dashboard := dashboards[0]
	d.Set("name", dashboard.Name)
	d.Set("display_period", dashboard.DisplayPeriod)
	d.Set("auto_start", dashboard.AutoStart)
	d.Set("private", dashboard.Private)

	// Process widgets from the first page
	widgets := make([]map[string]interface{}, 0)
	
	// Check if there are any pages and if the first page has widgets
	if len(dashboard.Pages) > 0 && dashboard.Pages[0].Widgets != nil {
		for _, widget := range dashboard.Pages[0].Widgets { // Iterate over widgets in the first page
			widgetMap := make(map[string]interface{})
			widgetMap["type"] = widget.Type
			widgetMap["name"] = widget.Name
			widgetMap["x"] = widget.X
			widgetMap["y"] = widget.Y
			widgetMap["width"] = widget.Width
			widgetMap["height"] = widget.Height

			// Process fields
			graphIds := make([]string, 0)
			itemIds := make([]string, 0)
			for _, field := range widget.Fields {
				if field.Name == "graphid" {
					graphIds = append(graphIds, field.Value)
				} else if field.Name == "itemid" {
					itemIds = append(itemIds, field.Value)
				}
			}

			if len(graphIds) > 0 {
				widgetMap["graph_ids"] = graphIds
			}
			if len(itemIds) > 0 {
				widgetMap["item_ids"] = itemIds
			}

			widgets = append(widgets, widgetMap)
		}
	}
	
	d.Set("widgets", widgets)

	return nil
}

func resourceZabbixDashboardExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	api := meta.(*zabbix.API)

	params := zabbix.Params{
		"dashboardids": d.Id(),
		"output":       "extend",
	}
	dashboards, err := DashboardsGet(api, params)
	if err != nil {
		if strings.Contains(err.Error(), "Expected exactly one result") {
			log.Printf("[DEBUG] Dashboard with id %s doesn't exist", d.Id())
			return false, nil
		}
		return false, err
	}
	if len(dashboards) == 0 {
		return false, nil
	}
	return true, nil
}

func resourceZabbixDashboardUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	dashboard, err := createDashboardObj(d)
	if err != nil {
		return err
	}
	dashboard.DashboardID = d.Id()

	dashboards := Dashboards{*dashboard}
	err = DashboardsUpdate(api, dashboards)
	if err != nil {
		return err
	}

	return resourceZabbixDashboardRead(d, meta)
}

func resourceZabbixDashboardDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)
	return DashboardsDeleteByIds(api, []string{d.Id()})
} 