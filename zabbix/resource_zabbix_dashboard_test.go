package zabbix

import (
	"fmt"
	"testing"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccZabbixDashboard_Basic(t *testing.T) {
	resourceName := "zabbix_dashboard.test"
	dashboardName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixDashboardConfig(dashboardName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixDashboardExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", dashboardName),
					resource.TestCheckResourceAttr(resourceName, "display_period", "30"),
					resource.TestCheckResourceAttr(resourceName, "auto_start", "1"),
					resource.TestCheckResourceAttr(resourceName, "private", "1"),
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

func testAccZabbixDashboardConfig(name string) string {
	return fmt.Sprintf(`
resource "zabbix_dashboard" "test" {
  name = "%s"
  display_period = 30
  auto_start = 1
  private = 1
}
`, name)
}

func testAccCheckZabbixDashboardExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No dashboard ID is set")
		}

		return nil
	}
}

func testAccCheckZabbixDashboardDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zabbix_dashboard" {
			continue
		}

		api := testAccProvider.Meta().(*zabbix.API)
		_, err := DashboardGetByID(api, rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("Dashboard still exists: %s", rs.Primary.ID)
		}
		
		// Check if the error is of type ErrorNotFound, which is expected
		if _, ok := err.(*ErrorNotFound); !ok {
			return fmt.Errorf("Expected ErrorNotFound but got: %v", err)
		}
	}

	return nil
} 