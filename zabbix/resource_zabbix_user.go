package zabbix

import (
	"fmt"
	"log"
	"strings"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceZabbixUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixUserCreate,
		Read:   resourceZabbixUserRead,
		Exists: resourceZabbixUserExists,
		Update: resourceZabbixUserUpdate,
		Delete: resourceZabbixUserDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username for the user.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name (first name) of the user.",
			},
			"surname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Surname (last name) of the user.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "Password for the user.",
			},
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "URL of the page to redirect the user to after logging in.",
			},
			"autologin": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether to enable auto-login for the user.",
			},
			"autologout": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "0",
				Description: "User session life time. If set to 0, the session will never expire.",
			},
			"lang": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
				Description: "Language code of the user's language.",
			},
			"refresh": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "30s",
				Description: "Automatic refresh period.",
			},
			"theme": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
				Description: "User's theme.",
			},
			"rows_per_page": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      50,
				Description:  "Amount of object rows to show per page.",
				ValidateFunc: validation.IntBetween(1, 999999),
			},
			"timezone": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "default",
				Description: "User's timezone.",
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the user role.",
			},
			"user_groups": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "User groups to add the user to.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"medias": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "User media (notification methods).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mediatypeid": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "ID of the media type used by the media.",
						},
						"sendto": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "Address, user name or other identifier of the recipient.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"active": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"severity": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      63,
							Description:  "Trigger severities to send notifications about.",
							ValidateFunc: validation.IntBetween(1, 63),
						},
						"period": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "1-7,00:00-24:00",
							Description: "Time when the notifications can be sent as a time period.",
						},
				},
			},
		},
	},
}
}

func getUserGroups(d *schema.ResourceData, api *zabbix.API) (zabbix.UserGroups, error) {
	configGroups := d.Get("user_groups").(*schema.Set)
	if configGroups.Len() == 0 {
		return zabbix.UserGroups{}, nil
	}

	setUserGroups := make([]string, configGroups.Len())
	for i, g := range configGroups.List() {
		setUserGroups[i] = g.(string)
	}

	log.Printf("[DEBUG] User Groups %v\n", setUserGroups)

	groupParams := zabbix.Params{
		"output": "extend",
		"filter": map[string]interface{}{
			"name": setUserGroups,
		},
	}

	groups, err := api.UserGroupsGet(groupParams)
	if err != nil {
		return nil, err
	}

	if len(groups) < configGroups.Len() {
		log.Printf("[DEBUG] Not all of the specified user groups were found on zabbix server")

		for _, n := range configGroups.List() {
			found := false
			for _, g := range groups {
				if n.(string) == g.Name {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("user group %s does not exist", n.(string))
			}
		}
	}

	userGroups := make(zabbix.UserGroups, len(groups))
	for i, group := range groups {
		userGroups[i] = zabbix.UserGroup{
			GroupID: group.GroupID,
		}
	}

	return userGroups, nil
}

func resourceZabbixUserCreate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	// Handle medias
	var medias zabbix.Medias
	if mediasData, ok := d.GetOk("medias"); ok {
		mediasList := mediasData.([]interface{})
		medias = make(zabbix.Medias, len(mediasList))

		for i, mediaData := range mediasList {
			mediaMap := mediaData.(map[string]interface{})
			media := zabbix.Media{
				MediaTypeID: mediaMap["mediatypeid"].(string),
				Active:      zabbix.MediaStatus(map[bool]int{true: 0, false: 1}[mediaMap["active"].(bool)]),
				Severity:    mediaMap["severity"].(int),
				Period:      mediaMap["period"].(string),
			}

			// Handle sendto list
			if sendtoList, ok := mediaMap["sendto"].([]interface{}); ok {
				sendToStrings := make([]string, len(sendtoList))
				for j, sendto := range sendtoList {
					sendToStrings[j] = sendto.(string)
				}
				media.SendTo = sendToStrings
			}

			medias[i] = media
		}
	}

	// Handle user groups
	userGroups, err := getUserGroups(d, api)
	if err != nil {
		return fmt.Errorf("error getting user groups: %v", err)
	}

	// Prepare user object
	user := zabbix.User{
		Username:    d.Get("username").(string),
		RoleID:      d.Get("role_id").(string),
		Password:    d.Get("password").(string),
		Name:        d.Get("name").(string),
		Surname:     d.Get("surname").(string),
		Url:         d.Get("url").(string),
		Autologout:  d.Get("autologout").(string),
		Lang:        d.Get("lang").(string),
		Refresh:     d.Get("refresh").(string),
		Theme:       d.Get("theme").(string),
		Timezone:    d.Get("timezone").(string),
		Autologin:   map[bool]int{true: 1, false: 0}[d.Get("autologin").(bool)],
		RowsPerPage: d.Get("rows_per_page").(int),
		Medias:      medias,
		UsrGrps:     userGroups,
	}

	users := zabbix.Users{user}

	err = api.UsersCreate(users)
	if err != nil {
		return fmt.Errorf("error creating user: %v", err)
	}

	userID := users[0].UserID
	log.Printf("[DEBUG] Created user, id is %s", userID)

	d.SetId(userID)

	return resourceZabbixUserRead(d, meta)
}

func resourceZabbixUserRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	users, err := api.UsersGet(zabbix.Params{
		"userids":       []string{d.Id()},
		"output":        "extend",
		"selectMedias":  "extend",
		"selectUsrgrps": []string{"name"},
	})

	if err != nil {
		return fmt.Errorf("error reading user: %v", err)
	}

	if len(users) == 0 {
		d.SetId("")
		return nil
	}

	user := users[0]

	d.Set("username", user.Username)
	d.Set("name", user.Name)
	d.Set("surname", user.Surname)
	d.Set("url", user.Url)
	d.Set("autologin", user.Autologin == 1)
	d.Set("autologout", user.Autologout)
	d.Set("lang", user.Lang)
	d.Set("refresh", user.Refresh)
	d.Set("theme", user.Theme)
	d.Set("rows_per_page", user.RowsPerPage)
	d.Set("timezone", user.Timezone)
	d.Set("role_id", user.RoleID)

	// Set medias
	mediasList := make([]interface{}, len(user.Medias))
	for i, media := range user.Medias {
		// Handle SendTo which can be either string or []string
		var sendTo []string
		switch v := media.SendTo.(type) {
		case string:
			sendTo = []string{v}
		case []interface{}:
			for _, item := range v {
				if str, ok := item.(string); ok {
					sendTo = append(sendTo, str)
				}
			}
		case []string:
			sendTo = v
		}

		mediaMap := map[string]interface{}{
			"mediatypeid": media.MediaTypeID,
			"sendto":      sendTo,
			"active":      int(media.Active) == 0,
			"severity":    media.Severity,
			"period":      media.Period,
		}
		mediasList[i] = mediaMap
	}
	d.Set("medias", mediasList)

	// Set user groups (use group names instead of IDs)
	userGroupsList := make([]string, len(user.UsrGrps))
	for i, userGroup := range user.UsrGrps {
		userGroupsList[i] = userGroup.Name
	}
	d.Set("user_groups", userGroupsList)

	return nil
}

func resourceZabbixUserExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	api := meta.(*zabbix.API)

	users, err := api.UsersGet(zabbix.Params{
		"userids": []string{d.Id()},
		"output":  []string{"userid"},
	})

	if err != nil {
		if strings.Contains(err.Error(), "Expected exactly one result") {
			log.Printf("[DEBUG] User with id %s doesn't exist", d.Id())
			return false, nil
		}
		return false, err
	}

	return len(users) > 0, nil
}

func resourceZabbixUserUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	user := zabbix.User{
		UserID: d.Id(),
	}

	// Update fields that have changed
	if d.HasChange("username") {
		user.Username = d.Get("username").(string)
	}
	if d.HasChange("name") {
		user.Name = d.Get("name").(string)
	}
	if d.HasChange("surname") {
		user.Surname = d.Get("surname").(string)
	}
	if d.HasChange("password") {
		// Password is required for updates if changed
		user.Password = d.Get("password").(string)
	}
	if d.HasChange("url") {
		user.Url = d.Get("url").(string)
	}
	if d.HasChange("autologin") {
		if d.Get("autologin").(bool) {
			user.Autologin = 1
		} else {
			user.Autologin = 0
		}
	}
	if d.HasChange("autologout") {
		user.Autologout = d.Get("autologout").(string)
	}
	if d.HasChange("lang") {
		user.Lang = d.Get("lang").(string)
	}
	if d.HasChange("refresh") {
		user.Refresh = d.Get("refresh").(string)
	}
	if d.HasChange("theme") {
		user.Theme = d.Get("theme").(string)
	}
	if d.HasChange("rows_per_page") {
		user.RowsPerPage = d.Get("rows_per_page").(int)
	}
	if d.HasChange("timezone") {
		user.Timezone = d.Get("timezone").(string)
	}
	if d.HasChange("role_id") {
		user.RoleID = d.Get("role_id").(string)
	}

	// Handle medias update
	if d.HasChange("medias") {
		if mediasData, ok := d.GetOk("medias"); ok {
			mediasList := mediasData.([]interface{})
			user.Medias = make(zabbix.Medias, len(mediasList))

			for i, mediaData := range mediasList {
				mediaMap := mediaData.(map[string]interface{})
				media := zabbix.Media{
					MediaTypeID: mediaMap["mediatypeid"].(string),
					Active:      zabbix.MediaStatus(map[bool]int{true: 0, false: 1}[mediaMap["active"].(bool)]),
					Severity:    mediaMap["severity"].(int),
					Period:      mediaMap["period"].(string),
				}

				// Handle sendto list
				if sendtoList, ok := mediaMap["sendto"].([]interface{}); ok {
					sendToStrings := make([]string, len(sendtoList))
					for j, sendto := range sendtoList {
						sendToStrings[j] = sendto.(string)
					}
					media.SendTo = sendToStrings
				}

				user.Medias[i] = media
			}
		} else {
			// Clear medias if not specified
			user.Medias = zabbix.Medias{}
		}
	}

	// Handle user groups update
	if d.HasChange("user_groups") {
		userGroups, err := getUserGroups(d, api)
		if err != nil {
			return fmt.Errorf("error getting user groups: %v", err)
		}
		user.UsrGrps = userGroups
	}

	users := zabbix.Users{user}
	err := api.UsersUpdate(users)
	if err != nil {
		return fmt.Errorf("error updating user: %v", err)
	}

	return resourceZabbixUserRead(d, meta)
}

func resourceZabbixUserDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	err := api.UsersDeleteByIds([]string{d.Id()})
	if err != nil {
		return fmt.Errorf("error deleting user: %v", err)
	}

	return nil
}
