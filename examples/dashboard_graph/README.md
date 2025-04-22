# Zabbix Dashboard and Graph Resources

This example demonstrates how to create and manage Zabbix dashboards and graphs using Terraform.

## Resources Created

- A host group for Linux servers
- A Linux server host with Zabbix agent
- CPU load and memory usage monitoring items
- Two graphs for visualizing the metrics
- A dashboard containing both graphs as widgets

## Dashboard Resource

The `zabbix_dashboard` resource allows you to create and manage Zabbix dashboards.

### Example Usage

```hcl
resource "zabbix_dashboard" "example" {
  name           = "My Dashboard"
  display_period = 30
  auto_start     = 1
  private        = 1

  widgets {
    type   = "graph"
    name   = "CPU Load"
    x      = 0
    y      = 0
    width  = 12
    height = 5
    graph_ids = [zabbix_graph.cpu_graph.id]
  }
}
```

### Argument Reference

* `name` - (Required) Name of the dashboard.
* `display_period` - (Optional) Page display period in seconds. Default is 30.
* `auto_start` - (Optional) Auto start slideshow. 0 - no, 1 - yes. Default is 1.
* `private` - (Optional) Dashboard private state. 0 - no, 1 - yes. Default is 1.
* `widgets` - (Optional) List of dashboard widgets.

#### Widget Arguments

* `type` - (Required) Type of the dashboard widget.
* `name` - (Required) Name of the widget.
* `x` - (Required) X position of the widget.
* `y` - (Required) Y position of the widget.
* `width` - (Required) Width of the widget.
* `height` - (Required) Height of the widget.
* `graph_ids` - (Optional) List of graph IDs to display in the widget.
* `item_ids` - (Optional) List of item IDs to display in the widget.

## Graph Resource

The `zabbix_graph` resource allows you to create and manage Zabbix graphs.

### Example Usage

```hcl
resource "zabbix_graph" "example" {
  name           = "CPU Load Graph"
  width          = 900
  height         = 300
  graph_type     = 0
  show_legend    = 1
  
  graph_items {
    item_id    = zabbix_item.cpu_load.id
    color      = "00AA00"
    calc_fnc   = 2
    type       = 0
  }
}
```

### Argument Reference

* `name` - (Required) Name of the graph.
* `width` - (Optional) Width of the graph in pixels. Default is 900.
* `height` - (Optional) Height of the graph in pixels. Default is 200.
* `graph_type` - (Optional) Graph type. 0 - normal, 1 - stacked, 2 - pie, 3 - exploded. Default is 0.
* `show_legend` - (Optional) Whether to show the legend. 0 - hide, 1 - show. Default is 1.
* `show_work_period` - (Optional) Whether to show the working time. 0 - hide, 1 - show. Default is 1.
* `show_triggers` - (Optional) Whether to show triggers. 0 - hide, 1 - show. Default is 1.
* `yaxis_min` - (Optional) Minimum value of the Y axis. Default is "0".
* `yaxis_max` - (Optional) Maximum value of the Y axis. Default is "100".
* `percent_left` - (Optional) Left percentile. Default is "0".
* `percent_right` - (Optional) Right percentile. Default is "0".
* `ymin_type` - (Optional) Minimum value calculation method. 0 - calculated, 1 - fixed, 2 - item. Default is 0.
* `ymax_type` - (Optional) Maximum value calculation method. 0 - calculated, 1 - fixed, 2 - item. Default is 0.
* `graph_items` - (Required) List of items to display on the graph.

#### Graph Item Arguments

* `item_id` - (Required) ID of the monitored item.
* `color` - (Required) Line color (6 symbols, hex).
* `calc_fnc` - (Optional) Value calculation function. 1 - min, 2 - avg, 4 - max, 7 - all, 9 - last. Default is 2.
* `type` - (Optional) Draw style. 0 - line, 1 - filled region, 2 - bold line, 3 - dot, 4 - dashed line, 5 - gradient line. Default is 0.
* `yaxis_side` - (Optional) Side of the Y axis. 0 - left, 1 - right. Default is 0.
* `sortorder` - (Optional) Sort order. Default is 0. 