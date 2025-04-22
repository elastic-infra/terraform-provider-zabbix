package zabbix

import (
	"fmt"
	"testing"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccZabbixHost_Basic(t *testing.T) {
	var getHost zabbix.Host
	randName := acctest.RandString(5)
	host := fmt.Sprintf("host_%s", randName)
	name := fmt.Sprintf("name_%s", randName)
	hostGroup := fmt.Sprintf("host_group_%s", randName)
	expectedHost1 := zabbix.Host{
		Host:       host,
		Name:       name,
		Interfaces: zabbix.HostInterfaces{zabbix.HostInterface{IP: "127.0.0.1", Main: 1, Port: "10050", Type: zabbix.Agent}},
	}
	expectedHost2 := zabbix.Host{
		Host:       host,
		Name:       name,
		Interfaces: zabbix.HostInterfaces{zabbix.HostInterface{DNS: "localhost", Main: 1, Port: "10050", Type: zabbix.Agent}},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixHostConfig(host, name, hostGroup),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixHostExists("zabbix_host.zabbix1", &getHost),
					testAccCheckZabbixHostAttributes(&getHost, expectedHost1, []string{hostGroup}),
					resource.TestCheckResourceAttr("zabbix_host.zabbix1", "host", host),
					resource.TestCheckResourceAttr("zabbix_host.zabbix1", "name", name),
				),
			},
			{
				Config: testAccZabbixHostUpdateConfig(host, name, hostGroup),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixHostExists("zabbix_host.zabbix1", &getHost),
					testAccCheckZabbixHostAttributes(&getHost, expectedHost2, []string{hostGroup}),
				),
			},
		},
	})
}

func TestAccZabbixHost_Templates(t *testing.T) {
	var getHost zabbix.Host
	randName := acctest.RandString(5)
	host := fmt.Sprintf("host_%s", randName)
	name := fmt.Sprintf("name_%s", randName)
	hostGroup := fmt.Sprintf("host_group_%s", randName)
	templateGroup := fmt.Sprintf("template_group_%s", randName)
	parentTemplate := fmt.Sprintf("template_%s", randName)
	expectedHost := zabbix.Host{
		Host:       host,
		Name:       name,
		Interfaces: zabbix.HostInterfaces{zabbix.HostInterface{IP: "127.0.0.1", Main: 1, Port: "10050", Type: zabbix.Agent}},
		Templates:  zabbix.Templates{zabbix.Template{Host: parentTemplate}},
		UserMacros: zabbix.Macros{zabbix.Macro{MacroName: "{$MACRO1}", Value: "value3"}},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixHostTemplatesConfig(host, name, hostGroup, templateGroup, parentTemplate),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixHostExists("zabbix_host.zabbix1", &getHost),
					testAccCheckZabbixHostAttributes(&getHost, expectedHost, []string{hostGroup}),
				),
			},
		},
	})
}

func testAccCheckZabbixHostDestroy(s *terraform.State) error {
	api := testAccProvider.Meta().(*zabbix.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zabbix_host" {
			continue
		}

		_, err := api.HostGroupGetByID(rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Host still exists")
		}
		expectedError := "Expected exactly one result, got 0."
		if err.Error() != expectedError {
			return fmt.Errorf("expected error : %s, got : %s", expectedError, err.Error())
		}
	}
	return nil
}

func testAccZabbixHostConfig(host string, name string, hostGroup string) string {
	return fmt.Sprintf(`
	  	resource "zabbix_host" "zabbix1" {
			host = "%s"
			name = "%s"
			interfaces {
		  		ip = "127.0.0.1"
				main = true
			}
			groups = ["${zabbix_host_group.zabbix.name}"]
	  	}

	  	resource "zabbix_host_group" "zabbix" {
			name = "%s"
	  	}`, host, name, hostGroup,
	)
}

func testAccZabbixHostUpdateConfig(host string, name string, hostGroup string) string {
	return fmt.Sprintf(`
	  	resource "zabbix_host" "zabbix1" {
			host = "%s"
			name = "%s"
			interfaces {
		  		dns = "localhost"
				main = true
			}
			groups = ["${zabbix_host_group.zabbix.name}"]
	  	}

	  	resource "zabbix_host_group" "zabbix" {
			name = "%s"
	  	}`, host, name, hostGroup,
	)
}

func testAccZabbixHostTemplatesConfig(host string, name string, hostGroup string, templateGroup string, parentTemplate string) string {
	return fmt.Sprintf(`
	  	resource "zabbix_host" "zabbix1" {
			host = "%s"
			name = "%s"
			interfaces {
		  		ip = "127.0.0.1"
				main = true
			}
			groups    = ["${zabbix_host_group.zabbix.name}"]
			templates = ["${zabbix_template.zabbix.host}"]
			macro = {
			  MACRO1 = "value3"
			}
	  	}

	  	resource "zabbix_host_group" "zabbix" {
			name = "%s"
	  	}

	  	resource "zabbix_template_group" "zabbix" {
			name = "%s"
	  	}

		resource "zabbix_template" "zabbix" {
			host = "%s"
			groups = ["${zabbix_template_group.zabbix.name}"]
			description = "test_template_description"
			macro = {
			  MACRO1 = "value1"
			  MACRO2 = "value2"
			}
		}`, host, name, hostGroup, templateGroup, parentTemplate,
	)
}

func testAccCheckZabbixHostExists(resource string, host *zabbix.Host) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No record ID id set")
		}

		api := testAccProvider.Meta().(*zabbix.API)
		hosts, err := api.HostsGet(zabbix.Params{
			"hostids":               rs.Primary.ID,
			"selectInterfaces":      "extend",
			"selectParentTemplates": []string{"name"},
			"selectMacros":          "extend",
		})

		if err != nil {
			return err
		}
		*host = hosts[0]
		return nil
	}
}

func testAccCheckZabbixHostAttributes(host *zabbix.Host, want zabbix.Host, groupNames []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		api := testAccProvider.Meta().(*zabbix.API)

		if host.Host != want.Host {
			return fmt.Errorf("Got host name: %q, expected: %q", host.Host, want.Host)
		}
		if host.Name != want.Name {
			return fmt.Errorf("Got name: %q, expected: %q", host.Name, want.Name)
		}

		if err := interfaceEquals(host.Interfaces, want.Interfaces); err != nil {
			return err
		}

		param := zabbix.Params{
			"output": "extend",
			"hostids": []string{
				host.HostID,
			},
		}
		groups, err := api.HostGroupsGet(param)
		if err != nil {
			return err
		}
		if len(groups) != len(groupNames) {
			return fmt.Errorf("Got %d groups, but expected %d groups", len(groups), len(groupNames))
		}
		for _, groupName := range groupNames {
			if !containGroup(groups, groupName) {
				return fmt.Errorf("Group not found: %s", groupName)
			}
		}

		if len(host.Templates) != len(want.Templates) {
			return fmt.Errorf("Got %d templates, but expected %d templates", len(host.Templates), len(want.Templates))
		}
		for _, template := range want.Templates {
			if !containTemplate(host.Templates, template.Host) {
				return fmt.Errorf("Template not found : %s", template.Host)
			}
		}

		if len(host.UserMacros) != len(want.UserMacros) {
			return fmt.Errorf("Got %d macros, but expected %d macros", len(host.UserMacros), len(want.UserMacros))
		}
		for _, macro := range want.UserMacros {
			if !containMacro(host.UserMacros, macro) {
				return fmt.Errorf("Macro not found : %s = %s", macro.MacroName, macro.Value)
			}
		}
		return nil
	}
}

func interfaceEquals(got zabbix.HostInterfaces, want zabbix.HostInterfaces) error {
	if len(got) != len(want) {
		return fmt.Errorf("Got %d interfaces, but expected %d interfaces", len(got), len(want))
	}
	for idx, iface := range want {
		if got[idx].IP != iface.IP {
			return fmt.Errorf("Got interface ip: %q, expected: %q", got[idx].IP, iface.IP)
		}
		if got[idx].DNS != iface.DNS {
			return fmt.Errorf("Got interface dns: %q, expected: %q", got[idx].DNS, iface.DNS)
		}
		if got[idx].Main != iface.Main {
			return fmt.Errorf("Got interface main: %q, expected: %q", got[idx].Main, iface.Main)
		}
		if got[idx].Port != iface.Port {
			return fmt.Errorf("Got interface port: %q, expected: %q", got[idx].Port, iface.Port)
		}
		if got[idx].Type != iface.Type {
			return fmt.Errorf("Got interface type: %q, expected: %q", got[idx].Type, iface.Type)
		}
	}
	return nil
}

func containGroup(groupNames zabbix.HostGroups, name string) bool {
	for _, group := range groupNames {
		if name == group.Name {
			return true
		}
	}
	return false
}

func containTemplate(templateNames zabbix.Templates, name string) bool {
	for _, template := range templateNames {
		if name == template.Name {
			return true
		}
	}
	return false
}

func containMacro(got zabbix.Macros, want zabbix.Macro) bool {
	for _, macro := range got {
		if macro.MacroName == want.MacroName && macro.Value == want.Value {
			return true
		}
	}
	return false
}
