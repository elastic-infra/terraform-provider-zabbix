package zabbix

import (
	"context"
	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceZabbixRole() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixCreateRole,
		ReadContext:   resourceZabbixReadRole,
		DeleteContext: resourceZabbixDeleteRole,
		UpdateContext: resourceZabbixUpdateRole,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(zabbix.ValidRoleTypes, false)),
			},
			"read_only": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceZabbixCreateRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	api := meta.(*zabbix.API)
	role, err := readRoleFromSchema(data)
	if err != nil {
		return diag.FromErr(err)
	}
	roles := zabbix.Roles{role}
	err = api.RolesCreateAndSetIDs(roles)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId(roles[0].RoleID)
	resourceZabbixReadRole(ctx, data, meta)
	return diags
}

func resourceZabbixUpdateRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	api := meta.(*zabbix.API)
	role, err := readRoleFromSchema(data)
	if err != nil {
		return diag.FromErr(err)
	}
	err = api.RolesUpdate(zabbix.Roles{role})
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceZabbixReadRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	api := meta.(*zabbix.API)
	role, err := api.RoleGetByID(data.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = data.Set("name", role.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	roleType, err := role.GetType()
	if err != nil {
		return diag.FromErr(err)
	}
	err = data.Set("type", roleType)
	if err != nil {
		return diag.FromErr(err)
	}
	err = data.Set("read_only", role.ReadOnly)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags
}

func resourceZabbixDeleteRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	api := meta.(*zabbix.API)
	roleId := data.Id()
	err := api.RolesDeleteByID(roleId)
	if err != nil {
		return diag.FromErr(err)
	}
	data.SetId("")
	return diags
}

func readRoleFromSchema(data *schema.ResourceData) (role zabbix.Role, err error) {
	roleType, err := zabbix.NewRoleType(data.Get("type").(string))
	role = zabbix.Role{
		RoleID: data.Id(),
		Type:   roleType,
		Name:   data.Get("name").(string),
	}
	return
}
