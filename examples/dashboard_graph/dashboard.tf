# Example of how to create a Zabbix dashboard with graphs

# Configure the host group
resource "zabbix_host_group" "linux_servers" {
  name = "Linux Servers"
}

# Configure a host
resource "zabbix_host" "linux_server1" {
  host = "linux-server1"
  name = "Linux Server 1"
  interfaces {
    ip   = "192.168.1.10"
    main = true
    type = "agent"
    port = "10050"
  }
  groups    = [zabbix_host_group.linux_servers.name]
  monitored = true
}

# Create monitoring items
resource "zabbix_item" "cpu_load" {
  name        = "CPU Load"
  key         = "system.cpu.load[percpu,avg1]"
  delay       = "60"
  history     = "90d"
  trends      = "365d"
  host_id     = zabbix_host.linux_server1.id
  interface_id = zabbix_host.linux_server1.interfaces[0].interface_id
  type        = 0 # Zabbix agent
  value_type  = 0 # Numeric float
}

resource "zabbix_item" "memory_usage" {
  name        = "Memory Usage"
  key         = "vm.memory.size[pused]"
  delay       = "60"
  history     = "90d"
  trends      = "365d"
  host_id     = zabbix_host.linux_server1.id
  interface_id = zabbix_host.linux_server1.interfaces[0].interface_id
  type        = 0 # Zabbix agent
  value_type  = 0 # Numeric float
}

# Create a graph for CPU load
resource "zabbix_graph" "cpu_load_graph" {
  name           = "CPU Load - ${zabbix_host.linux_server1.name}"
  width          = 900
  height         = 300
  graph_type     = 0 # Normal
  show_legend    = 1
  show_work_period = 1
  show_triggers   = 1
  
  graph_items {
    item_id    = zabbix_item.cpu_load.id
    color      = "00AA00" # Green
    calc_fnc   = 2 # Average
    type       = 0 # Simple
    yaxis_side = 0 # Left
    sortorder  = 0
  }
}

# Create a graph for memory usage
resource "zabbix_graph" "memory_usage_graph" {
  name           = "Memory Usage - ${zabbix_host.linux_server1.name}"
  width          = 900
  height         = 300
  graph_type     = 0 # Normal
  show_legend    = 1
  show_work_period = 1
  show_triggers   = 1
  
  graph_items {
    item_id    = zabbix_item.memory_usage.id
    color      = "AA0000" # Red
    calc_fnc   = 2 # Average
    type       = 0 # Simple
    yaxis_side = 0 # Left
    sortorder  = 0
  }
}

# Create a dashboard with both graphs
resource "zabbix_dashboard" "linux_dashboard" {
  name           = "Linux Server Dashboard"
  display_period = 30
  auto_start     = 1
  private        = 1

  # Add the CPU load graph widget
  widgets {
    type   = "graph"
    name   = "CPU Load"
    x      = 0
    y      = 0
    width  = 12
    height = 5
    graph_ids = [zabbix_graph.cpu_load_graph.id]
  }

  # Add the memory usage graph widget
  widgets {
    type   = "graph"
    name   = "Memory Usage"
    x      = 0
    y      = 5
    width  = 12
    height = 5
    graph_ids = [zabbix_graph.memory_usage_graph.id]
  }
} 