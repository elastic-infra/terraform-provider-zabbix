package zabbix

import (
	"fmt"
	"log"
	"strings"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceZabbixGraph() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixGraphCreate,
		Read:   resourceZabbixGraphRead,
		Exists: resourceZabbixGraphExists,
		Update: resourceZabbixGraphUpdate,
		Delete: resourceZabbixGraphDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the graph.",
			},
			"width": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     900,
				Description: "Width of the graph in pixels.",
			},
			"height": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     200,
				Description: "Height of the graph in pixels.",
			},
			"graph_type": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Graph type. 0 - normal, 1 - stacked, 2 - pie, 3 - exploded.",
			},
			"show_legend": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Whether to show the legend on the graph. 0 - hide, 1 - show.",
			},
			"show_work_period": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Whether to show the working time on the graph. 0 - hide, 1 - show.",
			},
			"show_triggers": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Whether to show the trigger line on the graph. 0 - hide, 1 - show.",
			},
			"yaxis_min": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "0",
				Description: "Minimum value of the Y axis.",
			},
			"yaxis_max": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "100",
				Description: "Maximum value of the Y axis.",
			},
			"percent_left": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "0",
				Description: "Left percentile.",
			},
			"percent_right": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "0",
				Description: "Right percentile.",
			},
			"ymin_type": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Minimum value calculation method for the Y axis. 0 - calculated, 1 - fixed, 2 - item.",
			},
			"ymax_type": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Maximum value calculation method for the Y axis. 0 - calculated, 1 - fixed, 2 - item.",
			},
			"graph_items": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"item_id": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the item.",
						},
						"color": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "Line color for the item (6 symbols, hex).",
						},
						"calc_fnc": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     2,
							Description: "Value calculation function. 1 - minimum value, 2 - average value, 4 - maximum value, 7 - all values, 9 - last value.",
						},
						"type": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Draw style of the graph item. 0 - line, 1 - filled region, 2 - bold line, 3 - dot, 4 - dashed line, 5 - gradient line.",
						},
						"yaxis_side": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Side of the Y axis. 0 - left, 1 - right.",
						},
						"sortorder": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "Sort order. The lowest value is displayed first.",
						},
					},
				},
			},
		},
	}
}

func createGraphItems(d *schema.ResourceData) (GraphItems, error) {
	graphItems := make(GraphItems, 0)
	terraformGraphItems := d.Get("graph_items").([]interface{})
	
	if len(terraformGraphItems) == 0 {
		return graphItems, fmt.Errorf("At least one graph item is required")
	}

	for _, terraformItem := range terraformGraphItems {
		item := terraformItem.(map[string]interface{})

		graphItem := GraphItem{
			ItemID:    item["item_id"].(string),
			Color:     item["color"].(string),
			CalcFnc:   item["calc_fnc"].(int),
			Type:      item["type"].(int),
			YaxisSide: item["yaxis_side"].(int),
			SortOrder: item["sortorder"].(int),
		}

		graphItems = append(graphItems, graphItem)
	}

	return graphItems, nil
}

func createGraphObj(d *schema.ResourceData) (*Graph, error) {
	graph := Graph{
		Name:           d.Get("name").(string),
		Width:          d.Get("width").(int),
		Height:         d.Get("height").(int),
		GraphType:      d.Get("graph_type").(int),
		ShowLegend:     d.Get("show_legend").(int),
		ShowWorkPeriod: d.Get("show_work_period").(int),
		ShowTriggers:   d.Get("show_triggers").(int),
		YaxisMin:       d.Get("yaxis_min").(string),
		YaxisMax:       d.Get("yaxis_max").(string),
		PercentLeft:    d.Get("percent_left").(string),
		PercentRight:   d.Get("percent_right").(string),
		YminType:       d.Get("ymin_type").(int),
		YmaxType:       d.Get("ymax_type").(int),
	}

	items, err := createGraphItems(d)
	if err != nil {
		return nil, err
	}
	graph.GitItems = items

	return &graph, nil
}

func resourceZabbixGraphCreate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	graph, err := createGraphObj(d)
	if err != nil {
		return err
	}

	graphs := Graphs{*graph}
	err = GraphsCreate(api, graphs)
	if err != nil {
		return err
	}

	d.SetId(graphs[0].GraphID)
	return resourceZabbixGraphRead(d, meta)
}

func resourceZabbixGraphRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	params := zabbix.Params{
		"graphids":       d.Id(),
		"output":         "extend",
		"selectGraphItems": "extend",
	}
	graphs, err := GraphsGet(api, params)
	if err != nil {
		return err
	}
	if len(graphs) != 1 {
		return fmt.Errorf("Expected one graph with id %s and got %d graphs", d.Id(), len(graphs))
	}

	graph := graphs[0]
	d.Set("name", graph.Name)
	d.Set("width", graph.Width)
	d.Set("height", graph.Height)
	d.Set("graph_type", graph.GraphType)
	d.Set("show_legend", graph.ShowLegend)
	d.Set("show_work_period", graph.ShowWorkPeriod)
	d.Set("show_triggers", graph.ShowTriggers)
	d.Set("yaxis_min", graph.YaxisMin)
	d.Set("yaxis_max", graph.YaxisMax)
	d.Set("percent_left", graph.PercentLeft)
	d.Set("percent_right", graph.PercentRight)
	d.Set("ymin_type", graph.YminType)
	d.Set("ymax_type", graph.YmaxType)

	graphItems := make([]map[string]interface{}, len(graph.GitItems))
	for i, item := range graph.GitItems {
		graphItems[i] = map[string]interface{}{
			"item_id":    item.ItemID,
			"color":      item.Color,
			"calc_fnc":   item.CalcFnc,
			"type":       item.Type,
			"yaxis_side": item.YaxisSide,
			"sortorder":  item.SortOrder,
		}
	}
	d.Set("graph_items", graphItems)

	return nil
}

func resourceZabbixGraphExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	api := meta.(*zabbix.API)

	params := zabbix.Params{
		"graphids": d.Id(),
		"output":   "extend",
	}
	graphs, err := GraphsGet(api, params)
	if err != nil {
		if strings.Contains(err.Error(), "Expected exactly one result") {
			log.Printf("[DEBUG] Graph with id %s doesn't exist", d.Id())
			return false, nil
		}
		return false, err
	}
	if len(graphs) == 0 {
		return false, nil
	}
	return true, nil
}

func resourceZabbixGraphUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	graph, err := createGraphObj(d)
	if err != nil {
		return err
	}
	graph.GraphID = d.Id()

	graphs := Graphs{*graph}
	err = GraphsUpdate(api, graphs)
	if err != nil {
		return err
	}

	return resourceZabbixGraphRead(d, meta)
}

func resourceZabbixGraphDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)
	return GraphsDeleteByIds(api, []string{d.Id()})
} 