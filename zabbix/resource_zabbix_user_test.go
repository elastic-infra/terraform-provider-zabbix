package zabbix

import (
	"fmt"
	"testing"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccZabbixUser_Basic(t *testing.T) {
	username := fmt.Sprintf("testuser_%s", acctest.RandString(5))
	var user zabbix.User

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixUserConfigBasic(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixUserExists("zabbix_user.test", &user),
					resource.TestCheckResourceAttr("zabbix_user.test", "username", username),
					resource.TestCheckResourceAttr("zabbix_user.test", "name", "Test"),
					resource.TestCheckResourceAttr("zabbix_user.test", "surname", "User"),
					resource.TestCheckResourceAttr("zabbix_user.test", "role_id", "1"),
					resource.TestCheckResourceAttrSet("zabbix_user.test", "user_id"),
				),
			},
		},
	})
}

func TestAccZabbixUser_WithMedia(t *testing.T) {
	username := fmt.Sprintf("testuser_%s", acctest.RandString(5))
	var user zabbix.User

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixUserConfigWithMedia(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixUserExists("zabbix_user.test", &user),
					resource.TestCheckResourceAttr("zabbix_user.test", "username", username),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.#", "1"),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.0.mediatypeid", "1"),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.0.sendto.#", "1"),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.0.sendto.0", "test@example.com"),
				),
			},
		},
	})
}

func TestAccZabbixUser_Update(t *testing.T) {
	username := fmt.Sprintf("testuser_%s", acctest.RandString(5))
	var user zabbix.User

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixUserConfigBasic(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixUserExists("zabbix_user.test", &user),
					resource.TestCheckResourceAttr("zabbix_user.test", "name", "Test"),
					resource.TestCheckResourceAttr("zabbix_user.test", "surname", "User"),
				),
			},
			{
				Config: testAccZabbixUserConfigUpdated(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixUserExists("zabbix_user.test", &user),
					resource.TestCheckResourceAttr("zabbix_user.test", "name", "Updated"),
					resource.TestCheckResourceAttr("zabbix_user.test", "surname", "UpdatedUser"),
					resource.TestCheckResourceAttr("zabbix_user.test", "theme", "blue-theme"),
				),
			},
		},
	})
}

func testAccCheckZabbixUserDestroy(s *terraform.State) error {
	api := testAccProvider.Meta().(*zabbix.API)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "zabbix_user" {
			continue
		}

		users, err := api.UsersGet(zabbix.Params{
			"userids": []string{rs.Primary.ID},
			"output":  []string{"userid"},
		})

		if err != nil {
			return nil
		}

		if len(users) > 0 {
			return fmt.Errorf("user still exists")
		}
	}

	return nil
}

func testAccCheckZabbixUserExists(name string, user *zabbix.User) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User ID is set")
		}

		api := testAccProvider.Meta().(*zabbix.API)
		users, err := api.UsersGet(zabbix.Params{
			"userids":      []string{rs.Primary.ID},
			"output":       "extend",
			"selectMedias": "extend",
		})

		if err != nil {
			return err
		}

		if len(users) != 1 {
			return fmt.Errorf("User not found")
		}

		*user = users[0]

		return nil
	}
}

func testAccZabbixUserConfigBasic(username string) string {
	return fmt.Sprintf(`
resource "zabbix_user" "test" {
	username   = "%s"
	name       = "Test"
	surname    = "User"
	password   = "ComplexPassword123!"
	role_id    = "1"
	lang       = "en_US"
	theme      = "default"
	autologin  = false
	autologout = "15m"
	refresh    = "30s"
	timezone   = "default"
	rows_per_page = 50
}
`, username)
}

func testAccZabbixUserConfigWithMedia(username string) string {
	return fmt.Sprintf(`
resource "zabbix_user" "test" {
	username   = "%s"
	name       = "Test"
	surname    = "User"
	password   = "ComplexPassword123!"
	role_id    = "1"

	medias {
		mediatypeid = "1"
		sendto      = ["test@example.com"]
		active      = 0
		severity    = 63
		period      = "1-7,00:00-24:00"
	}
}
`, username)
}

func testAccZabbixUserConfigUpdated(username string) string {
	return fmt.Sprintf(`
resource "zabbix_user" "test" {
	username   = "%s"
	name       = "Updated"
	surname    = "UpdatedUser"
	password   = "ComplexPassword123!"
	role_id    = "1"
	theme      = "blue-theme"
	lang       = "ja_JP"
	rows_per_page = 100
}
`, username)
}

func TestAccZabbixUser_MultipleMedias(t *testing.T) {
	username := fmt.Sprintf("testuser_%s", acctest.RandString(5))
	var user zabbix.User

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixUserConfigMultipleMedias(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixUserExists("zabbix_user.test", &user),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.#", "2"),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.0.mediatypeid", "1"),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.1.mediatypeid", "3"),
				),
			},
		},
	})
}

func TestAccZabbixUser_WithUserGroups(t *testing.T) {
	username := fmt.Sprintf("testuser_%s", acctest.RandString(5))
	var user zabbix.User

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixUserConfigWithUserGroups(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixUserExists("zabbix_user.test", &user),
					resource.TestCheckResourceAttr("zabbix_user.test", "user_groups.#", "1"),
					resource.TestCheckResourceAttr("zabbix_user.test", "user_groups.0", "Zabbix administrators"),
				),
			},
		},
	})
}

func TestAccZabbixUser_MediaUpdate(t *testing.T) {
	username := fmt.Sprintf("testuser_%s", acctest.RandString(5))
	var user zabbix.User

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckZabbixUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccZabbixUserConfigWithMedia(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixUserExists("zabbix_user.test", &user),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.#", "1"),
				),
			},
			{
				Config: testAccZabbixUserConfigMediaUpdated(username),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckZabbixUserExists("zabbix_user.test", &user),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.#", "1"),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.0.severity", "60"),
					resource.TestCheckResourceAttr("zabbix_user.test", "medias.0.active", "1"),
				),
			},
		},
	})
}

func testAccZabbixUserConfigMultipleMedias(username string) string {
	return fmt.Sprintf(`
resource "zabbix_user" "test" {
	username   = "%s"
	name       = "Test"
	surname    = "User"
	password   = "ComplexPassword123!"
	role_id    = "1"

	medias {
		mediatypeid = "1"
		sendto      = ["test@example.com"]
		active      = 0
		severity    = 63
		period      = "1-7,00:00-24:00"
	}

	medias {
		mediatypeid = "3"
		sendto      = ["#slack-channel"]
		active      = 0
		severity    = 60
		period      = "1-7,00:00-24:00"
	}
}
`, username)
}

func testAccZabbixUserConfigWithUserGroups(username string) string {
	return fmt.Sprintf(`
resource "zabbix_user" "test" {
	username   = "%s"
	name       = "Test"
	surname    = "User"
	password   = "ComplexPassword123!"
	role_id    = "1"

	user_groups = ["Zabbix administrators"]
}
`, username)
}

func testAccZabbixUserConfigMediaUpdated(username string) string {
	return fmt.Sprintf(`
resource "zabbix_user" "test" {
	username   = "%s"
	name       = "Test"
	surname    = "User"
	password   = "ComplexPassword123!"
	role_id    = "1"

	medias {
		mediatypeid = "1"
		sendto      = ["test@example.com"]
		active      = 1
		severity    = 60
		period      = "1-5,09:00-18:00"
	}
}
`, username)
}
