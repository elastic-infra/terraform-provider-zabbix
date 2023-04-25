package zabbix

import (
	"context"
	"github.com/atypon/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceZabbixUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixCreateUser,
		ReadContext:   resourceZabbixReadUser,
		DeleteContext: resourceZabbixDeleteUser,
		UpdateContext: resourceZabbixUpdateUser,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"surname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"groups": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceZabbixReadUser(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	user := zabbix.User{UserID: data.Id()}
	err := api.ReadAPIObject(&user)
	errors.addError(err)
	err = data.Set("username", user.Username)
	errors.addError(err)
	err = data.Set("name", user.Name)
	errors.addError(err)
	err = data.Set("surname", user.Surname)
	errors.addError(err)
	err = data.Set("role_id", user.RoleID)
	errors.addError(err)
	err = data.Set("groups", user.Groups)
	errors.addError(err)
	return errors.getDiagnostics()
}

func resourceZabbixCreateUser(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	user, err := createUserFromResourceData(data)
	errors.addError(err)
	err = api.CreateAPIObject(&user)
	errors.addError(err)
	data.SetId(user.GetID())
	resourceZabbixReadUser(ctx, data, meta)
	return errors.getDiagnostics()
}

func resourceZabbixUpdateUser(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	user, err := createUserFromResourceData(data)
	errors.addError(err)
	err = api.UpdateAPIObject(&user)
	errors.addError(err)
	return errors.getDiagnostics()
}

func resourceZabbixDeleteUser(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	user := zabbix.User{UserID: data.Id()}
	err := api.DeleteAPIObject(&user)
	errors.addError(err)
	data.SetId("")
	return errors.getDiagnostics()
}

func createUserFromResourceData(data *schema.ResourceData) (userGroup zabbix.User, err error) {
	var groups []zabbix.UserGroupID
	for _, v := range data.Get("groups").([]any) {
		group := zabbix.UserGroupID(v.(string))
		groups = append(groups, group)
	}
	userGroup = zabbix.User{
		UserID:   data.Id(),
		Name:     data.Get("name").(string),
		Username: data.Get("username").(string),
		Surname:  data.Get("surname").(string),
		RoleID:   data.Get("role_id").(string),
		Groups:   groups,
	}
	return
}
