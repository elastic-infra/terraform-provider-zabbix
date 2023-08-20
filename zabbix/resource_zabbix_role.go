package zabbix

import (
	"context"
	"github.com/atypon/go-zabbix-api"
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceZabbixCreateRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	role, err := readRoleFromSchema(data)
	errors.addError(err)
	err = api.CreateAPIObject(&role)
	errors.addError(err)
	data.SetId(role.GetID())
	resourceZabbixReadRole(ctx, data, meta)
	return errors.getDiagnostics()
}

func resourceZabbixUpdateRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	role, err := readRoleFromSchema(data)
	errors.addError(err)
	err = api.UpdateAPIObject(&role)
	errors.addError(err)
	return errors.getDiagnostics()
}

func resourceZabbixReadRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	role := zabbix.Role{RoleID: data.Id()}
	err := api.ReadAPIObject(&role)
	errors.addError(err)
	err = data.Set("name", role.Name)
	errors.addError(err)
	roleType, err := role.GetType()
	errors.addError(err)
	err = data.Set("type", roleType)
	errors.addError(err)
	err = data.Set("read_only", role.ReadOnly)
	errors.addError(err)
	return errors.getDiagnostics()
}

func resourceZabbixDeleteRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	role := zabbix.Role{}
	role.SetID(data.Id())
	err := api.DeleteAPIObject(&role)
	errors.addError(err)
	data.SetId("")
	return errors.getDiagnostics()
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
