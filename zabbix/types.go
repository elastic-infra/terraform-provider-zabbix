package zabbix

import (
	"encoding/json"
	"fmt"

	"github.com/claranet/go-zabbix-api"
)

// Dashboard defines a Zabbix dashboard
type Dashboard struct {
	DashboardID   string          `json:"dashboardid,omitempty"`
	Name          string          `json:"name"`
	DisplayPeriod int             `json:"display_period,string"`
	AutoStart     int             `json:"auto_start,string"`
	Private       int             `json:"private,string"`
	Pages         []DashboardPage `json:"pages"`
}

// Dashboards is an array of Dashboard
type Dashboards []Dashboard

// DashboardPage defines a page within a Zabbix dashboard
type DashboardPage struct {
	DashboardPageID string `json:"dashboard_pageid,omitempty"`
	Name            string `json:"name,omitempty"` // Optional page name
	DisplayPeriod   int    `json:"display_period,string,omitempty"` // Optional page-specific period
	Widgets         Widgets `json:"widgets"`
}

// Widget defines a Zabbix dashboard widget
type Widget struct {
	WidgetID string       `json:"widgetid,omitempty"`
	Type     string       `json:"type"`
	Name     string       `json:"name"`
	X        int          `json:"x,string"`
	Y        int          `json:"y,string"`
	Width    int          `json:"width,string"`
	Height   int          `json:"height,string"`
	Fields   WidgetFields `json:"fields,omitempty"`
}

// Widgets is an array of Widget
type Widgets []Widget

// WidgetField defines a field for a dashboard widget
type WidgetField struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

// WidgetFields is an array of WidgetField
type WidgetFields []WidgetField

// Graph defines a Zabbix graph
type Graph struct {
	GraphID        string     `json:"graphid,omitempty"`
	Name           string     `json:"name"`
	Width          int        `json:"width,string,omitempty"`
	Height         int        `json:"height,string,omitempty"`
	GraphType      int        `json:"graphtype,string,omitempty"`
	ShowLegend     int        `json:"show_legend,string,omitempty"`
	ShowWorkPeriod int        `json:"show_work_period,string,omitempty"`
	ShowTriggers   int        `json:"show_triggers,string,omitempty"`
	YaxisMin       string     `json:"yaxismin,omitempty"`
	YaxisMax       string     `json:"yaxismax,omitempty"`
	PercentLeft    string     `json:"percent_left,omitempty"`
	PercentRight   string     `json:"percent_right,omitempty"`
	YminType       int        `json:"ymin_type,string,omitempty"`
	YmaxType       int        `json:"ymax_type,string,omitempty"`
	GitItems       GraphItems `json:"gitems,omitempty"`
}

// Graphs is an array of Graph
type Graphs []Graph

// GraphItem defines an item displayed on a graph
type GraphItem struct {
	ItemID    string `json:"itemid"`
	Color     string `json:"color"`
	CalcFnc   int    `json:"calc_fnc,string,omitempty"`
	Type      int    `json:"type,string,omitempty"`
	YaxisSide int    `json:"yaxisside,string,omitempty"`
	SortOrder int    `json:"sortorder,string,omitempty"`
}

// GraphItems is an array of GraphItem
type GraphItems []GraphItem

// ErrorNotFound custom error for not found objects
type ErrorNotFound struct {
	Message string
}

func (e ErrorNotFound) Error() string {
	return e.Message
}

// API method functions (these don't use receivers but take the API as a parameter)

// DashboardsGet gets dashboards by params
func DashboardsGet(api *zabbix.API, params zabbix.Params) (Dashboards, error) {
	response, err := api.CallWithError("dashboard.get", params)
	if err != nil {
		return nil, err
	}

	var dashboards Dashboards
	bytes, err := json.Marshal(response.Result)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &dashboards)
	return dashboards, err
}

// DashboardGetByID gets dashboard by ID
func DashboardGetByID(api *zabbix.API, id string) (Dashboard, error) {
	dashboards, err := DashboardsGet(api, zabbix.Params{"dashboardids": id})
	if err != nil {
		return Dashboard{}, err
	}
	if len(dashboards) == 0 {
		return Dashboard{}, &ErrorNotFound{Message: fmt.Sprintf("Dashboard with ID %s not found", id)}
	}
	return dashboards[0], nil
}

// DashboardsCreate creates new dashboards
func DashboardsCreate(api *zabbix.API, dashboards Dashboards) error {
	response, err := api.CallWithError("dashboard.create", dashboards)
	if err != nil {
		return err
	}

	// Extract the created dashboard IDs
	var result map[string][]interface{}
	bytes, err := json.Marshal(response.Result)
	if err != nil {
		return err
	}
	
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	// Update the dashboard IDs in the input slice
	if dashboardids, ok := result["dashboardids"]; ok && len(dashboardids) > 0 {
		for i, id := range dashboardids {
			if i < len(dashboards) {
				if strID, ok := id.(string); ok {
					dashboards[i].DashboardID = strID
				}
			}
		}
	}

	return nil
}

// DashboardsUpdate updates dashboards
func DashboardsUpdate(api *zabbix.API, dashboards Dashboards) error {
	_, err := api.CallWithError("dashboard.update", dashboards)
	return err
}

// DashboardsDeleteByIds deletes dashboards by ids
func DashboardsDeleteByIds(api *zabbix.API, ids []string) error {
	_, err := api.CallWithError("dashboard.delete", ids)
	return err
}

// GraphsGet gets graphs by params
func GraphsGet(api *zabbix.API, params zabbix.Params) (Graphs, error) {
	response, err := api.CallWithError("graph.get", params)
	if err != nil {
		return nil, err
	}

	var graphs Graphs
	bytes, err := json.Marshal(response.Result)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(bytes, &graphs)
	return graphs, err
}

// GraphGetByID gets graph by ID
func GraphGetByID(api *zabbix.API, id string) (Graph, error) {
	graphs, err := GraphsGet(api, zabbix.Params{"graphids": id})
	if err != nil {
		return Graph{}, err
	}
	if len(graphs) == 0 {
		return Graph{}, &ErrorNotFound{Message: fmt.Sprintf("Graph with ID %s not found", id)}
	}
	return graphs[0], nil
}

// GraphsCreate creates new graphs
func GraphsCreate(api *zabbix.API, graphs Graphs) error {
	response, err := api.CallWithError("graph.create", graphs)
	if err != nil {
		return err
	}

	// Extract the created graph IDs
	var result map[string][]interface{}
	bytes, err := json.Marshal(response.Result)
	if err != nil {
		return err
	}
	
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}

	// Update the graph IDs in the input slice
	if graphids, ok := result["graphids"]; ok && len(graphids) > 0 {
		for i, id := range graphids {
			if i < len(graphs) {
				if strID, ok := id.(string); ok {
					graphs[i].GraphID = strID
				}
			}
		}
	}

	return nil
}

// GraphsUpdate updates graphs
func GraphsUpdate(api *zabbix.API, graphs Graphs) error {
	_, err := api.CallWithError("graph.update", graphs)
	return err
}

// GraphsDeleteByIds deletes graphs by ids
func GraphsDeleteByIds(api *zabbix.API, ids []string) error {
	_, err := api.CallWithError("graph.delete", ids)
	return err
} 