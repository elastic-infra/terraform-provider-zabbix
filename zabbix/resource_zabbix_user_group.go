package zabbix

import (
	"context"
	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceZabbixUserGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixCreateUserGroup,
		ReadContext:   resourceZabbixReadUserGroup,
		DeleteContext: resourceZabbixDeleteUserGroup,
		UpdateContext: resourceZabbixUpdateUserGroup,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"debug_mode": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"gui_access": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceZabbixCreateUserGroup(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	userGroup, err := createUserGroupFromResourceData(data)
	errors.addError(err)
	err = api.CreateAPIObject(&userGroup)
	errors.addError(err)
	data.SetId(userGroup.GetID())
	resourceZabbixReadUserGroup(ctx, data, meta)
	return errors.getDiagnostics()
}

func resourceZabbixReadUserGroup(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	userGroup := zabbix.UserGroup{GroupID: data.Id()}
	err := api.ReadAPIObject(&userGroup)
	errors.addError(err)
	err = data.Set("name", userGroup.Name)
	errors.addError(err)
	err = data.Set("debug_mode", userGroup.DebugMode)
	errors.addError(err)
	err = data.Set("gui_access", userGroup.GuiAccess)
	errors.addError(err)
	return errors.getDiagnostics()
}

func resourceZabbixUpdateUserGroup(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	userGroup, err := createUserGroupFromResourceData(data)
	errors.addError(err)
	err = api.UpdateAPIObject(&userGroup)
	errors.addError(err)
	return errors.getDiagnostics()
}

func resourceZabbixDeleteUserGroup(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	userGroup := zabbix.UserGroup{GroupID: data.Id()}
	err := api.DeleteAPIObject(&userGroup)
	errors.addError(err)
	data.SetId("")
	return errors.getDiagnostics()
}

func createUserGroupFromResourceData(data *schema.ResourceData) (userGroup zabbix.UserGroup, err error) {
	userGroup = zabbix.UserGroup{
		GroupID:   data.Id(),
		Name:      data.Get("name").(string),
		GuiAccess: data.Get("gui_access").(int),
		DebugMode: zabbix.DebugModeType(data.Get("debug_mode").(int)),
	}
	return
}
