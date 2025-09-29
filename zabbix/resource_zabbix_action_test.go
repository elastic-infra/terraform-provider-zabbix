package zabbix

import (
	"fmt"
	"testing"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccZabbixAction_Basic(t *testing.T) {
	actionName := fmt.Sprintf("test_action_%s", acctest.RandString(5))
	var action zabbix.Action
	expectedAction := zabbix.Action{Name: actionName}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixActionConfig(actionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixActionExists("zabbix_action.zabbix", &action),
					testAccCheckZabbixActionAttributes(&action, expectedAction),
					resource.TestCheckResourceAttr("zabbix_action.zabbix", "name", actionName),
					resource.TestCheckResourceAttr("zabbix_action.zabbix", "event_source", "trigger"),
					resource.TestCheckResourceAttr("zabbix_action.zabbix", "enabled", "true"),
					resource.TestCheckResourceAttr("zabbix_action.zabbix", "operation.#", "1"),
					resource.TestCheckResourceAttr("zabbix_action.zabbix", "operation.0.type", "send_message"),
				),
			},
		},
	})
}

func TestAccZabbixAction_Update(t *testing.T) {
	actionName := fmt.Sprintf("test_action_%s", acctest.RandString(5))
	updatedActionName := fmt.Sprintf("updated_action_%s", acctest.RandString(5))
	var action zabbix.Action

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixActionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixActionConfig(actionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixActionExists("zabbix_action.zabbix", &action),
					resource.TestCheckResourceAttr("zabbix_action.zabbix", "name", actionName),
					resource.TestCheckResourceAttr("zabbix_action.zabbix", "enabled", "true"),
				),
			},
			{
				Config: testAccZabbixActionConfigUpdated(updatedActionName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixActionExists("zabbix_action.zabbix", &action),
					resource.TestCheckResourceAttr("zabbix_action.zabbix", "name", updatedActionName),
					resource.TestCheckResourceAttr("zabbix_action.zabbix", "enabled", "false"),
					resource.TestCheckResourceAttr("zabbix_action.zabbix", "operation.0.message.0.default_message", "true"),
				),
			},
		},
	})
}

func testAccCheckZabbixActionDestroy(s *terraform.State) error {
	api := testAccProvider.Meta().(*zabbix.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zabbix_action" {
			continue
		}

		_, err := api.ActionGetByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Action still exists")
		}
		expectedError := "Expected exactly one result, got 0."
		if err.Error() != expectedError {
			return fmt.Errorf("expected error : %s, got : %s", expectedError, err.Error())
		}
	}
	return nil
}

func testAccZabbixActionConfig(actionName string) string {
	return fmt.Sprintf(`
		resource "zabbix_action" "zabbix" {
			name         = "%s"
			event_source = "trigger"
			enabled      = true

			condition {
				type     = "trigger_name"
				operator = "contains"
				value    = "test"
			}

			operation {
				type = "send_message"

				message {
					default_message = false
					subject         = "Test Alert"
					message         = "Test message body"

					target {
						type  = "user_group"
						value = "Zabbix administrators"
					}
				}
			}
		}`, actionName)
}

func testAccZabbixActionConfigUpdated(actionName string) string {
	return fmt.Sprintf(`
		resource "zabbix_action" "zabbix" {
			name            = "%s"
			event_source    = "trigger"
			enabled         = false

			condition {
				type     = "trigger_name"
				operator = "contains"
				value    = "test"
			}

			operation {
				type = "send_message"

				message {
					default_message = true

					target {
						type  = "user_group"
						value = "Zabbix administrators"
					}
				}
			}
		}`, actionName)
}

func testAccCheckZabbixActionExists(resource string, action *zabbix.Action) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found; %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No record ID set")
		}

		api := testAccProvider.Meta().(*zabbix.API)
		act, err := api.ActionGetByID(rs.Primary.ID)
		if err != nil {
			return err
		}
		*action = *act
		return nil
	}
}

func testAccCheckZabbixActionAttributes(action *zabbix.Action, want zabbix.Action) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if action.Name != want.Name {
			return fmt.Errorf("got action name : %q, expected : %q", action.Name, want.Name)
		}
		return nil
	}
}
