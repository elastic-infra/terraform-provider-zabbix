package zabbix

import (
	"fmt"
	"testing"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccZabbixGraph_Basic(t *testing.T) {
	resourceName := "zabbix_graph.test"
	graphName := acctest.RandString(10)
	hostName := acctest.RandString(10)
	hostGroupName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixGraphDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixGraphConfigBasic(hostName, hostGroupName, graphName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixGraphExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", graphName),
					resource.TestCheckResourceAttr(resourceName, "width", "900"),
					resource.TestCheckResourceAttr(resourceName, "height", "200"),
					resource.TestCheckResourceAttr(resourceName, "graph_items.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccZabbixGraphConfigBasic(hostName, hostGroupName, graphName string) string {
	return fmt.Sprintf(`
resource "zabbix_host_group" "test" {
  name = "%s"
}

resource "zabbix_host" "test" {
  host = "%s"
  name = "%s"
  interfaces {
    ip = "127.0.0.1"
    main = true
    dns = "localhost"
    type = "agent"
    port = "10050"
  }
  groups = ["${zabbix_host_group.test.name}"]
  monitored = true
}

resource "zabbix_item" "test" {
  name = "Test Item"
  key = "system.hostname"
  delay = "60"
  host_id = "${zabbix_host.test.id}"
  interface_id = "${zabbix_host.test.interfaces[0].interface_id}"
  type = 0
}

resource "zabbix_graph" "test" {
  name = "%s"
  width = 900
  height = 200
  graph_type = 0
  show_legend = 1
  show_work_period = 1
  show_triggers = 1
  yaxis_min = "0"
  yaxis_max = "100"
  
  graph_items {
    item_id = "${zabbix_item.test.id}"
    color = "00AA00"
    calc_fnc = 2
    type = 0
    yaxis_side = 0
    sortorder = 0
  }
}
`, hostGroupName, hostName, hostName, graphName)
}

func testAccCheckZabbixGraphExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No graph ID is set")
		}

		return nil
	}
}

func testAccCheckZabbixGraphDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zabbix_graph" {
			continue
		}

		api := testAccProvider.Meta().(*zabbix.API)
		_, err := GraphGetByID(api, rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Graph still exists: %s", rs.Primary.ID)
		}
		
		// Check if the error is of type ErrorNotFound, which is expected
		if _, ok := err.(*ErrorNotFound); !ok {
			return fmt.Errorf("Expected ErrorNotFound but got: %v", err)
		}
	}

	return nil
} 