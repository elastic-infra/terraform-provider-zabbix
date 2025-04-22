package zabbix

import (
	"fmt"
	"log"
	"strings"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceZabbixTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixTemplateCreate,
		Read:   resourceZabbixTemplateRead,
		Exists: resourceZabbixTemplateExists,
		Update: resourceZabbixTemplateUpdate,
		Delete: resourceZabbixTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Technical name of the template.",
			},
			"groups": &schema.Schema{
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: "ID of the Host Group.",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Visible name of the template.",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the template.",
			},
			"macro": &schema.Schema{
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "User macros for the template.",
			},
			"linked_template": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
		},
	}
}

func createZabbixMacro(d *schema.ResourceData) zabbix.Macros {
	var macros zabbix.Macros

	terraformMacros := d.Get("macro").(map[string]interface{})
	for i, terraformMacro := range terraformMacros {
		macro := zabbix.Macro{
			MacroName: fmt.Sprintf("{$%s}", i),
			Value:     terraformMacro.(string),
		}
		macros = append(macros, macro)
	}
	return macros
}

func createLinkedTemplate(d *schema.ResourceData) zabbix.TemplateIDs {
	var templates zabbix.TemplateIDs

	terraformTemplates := d.Get("linked_template").(*schema.Set)
	for _, terraformTemplate := range terraformTemplates.List() {
		zabbixTemplate := zabbix.TemplateID{
			TemplateID: terraformTemplate.(string),
		}
		templates = append(templates, zabbixTemplate)
	}
	return templates
}

func createTemplateObj(d *schema.ResourceData, api *zabbix.API) (*zabbix.Template, error) {
	template := zabbix.Template{
		Host:            d.Get("host").(string),
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		UserMacros:      createZabbixMacro(d),
		LinkedTemplates: createLinkedTemplate(d),
	}

	var groupIds zabbix.HostGroupIDs
	var err error
	if api.ServerVersion.GreaterThanOrEqual(version.Must(version.NewVersion("6.2"))) {
		groupIds, err = getTemplateGroups(d, api)
	} else {
		groupIds, err = getHostGroups(d, api)
	}

	if err != nil {
		return nil, err
	}
	template.Groups = groupIds
	if template.UserMacros == nil {
		template.UserMacros = zabbix.Macros{}
	}
	return &template, nil
}

func resourceZabbixTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	template, err := createTemplateObj(d, api)
	if err != nil {
		return err
	}

	return createRetry(d, meta, createTemplate, *template, resourceZabbixTemplateRead)
}

func resourceZabbixTemplateRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	params := zabbix.Params{
		"templateids":  d.Id(),
		"output":       "extend",
		"selectMacros": "extend",
	}
	templates, err := api.TemplatesGet(params)
	if err != nil {
		return err
	}
	if len(templates) != 1 {
		log.Printf("[DEBUG] Expected one template with id %s and got %#v", d.Id(), templates)
		return fmt.Errorf("Expected one template with id %s and got %d templates", d.Id(), len(templates))
	}

	template := templates[0]
	d.Set("host", template.Host)
	if template.Host != template.Name && d.Get("name").(string) == "" {
		d.Set("name", template.Name)
	}
	d.Set("description", template.Description)

	terraformMacros, err := createTerraformMacro(template)
	if err != nil {
		return err
	}
	d.Set("macro", terraformMacros)

	terraformGroups, err := createTerraformTemplateGroup(d, api)
	if err != nil {
		return err
	}
	d.Set("groups", terraformGroups)
	return nil
}

func resourceZabbixTemplateExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	api := meta.(*zabbix.API)

	_, err := api.TemplateGetByID(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "Expected exactly one result") {
			log.Printf("[DEBUG] Template with id %s doesn't exist", d.Id())
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func resourceZabbixTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	template, err := createTemplateObj(d, api)
	if err != nil {
		return err
	}
	template.TemplatesClear = getUnlinkedTemplate(d)
	template.TemplateID = d.Id()

	// Check if macros need to be cleared
	macrosCleared := false
	if d.HasChange("macro") {
		_, newMacros := d.GetChange("macro")
		if len(newMacros.(map[string]interface{})) == 0 {
			macrosCleared = true
		}
	}

	// If macros were cleared, we need a special update call payload to bypass omitempty
	if macrosCleared {
		log.Printf("[DEBUG] Clearing macros for template ID %s via custom payload", d.Id())
		// Construct a payload containing only the template ID and the empty macros array
		// The API should only update the macros field when receiving this.
		updatePayload := map[string]interface{}{
			"templateid": template.TemplateID,
			"macros":        []zabbix.Macro{},
		}
		// Use a retry mechanism for the specific macro clearing update
		updateFunc := func(payload interface{}, a *zabbix.API) (string, error) {
			_, err := a.CallWithError("template.update", payload.(map[string]interface{}))
			if err != nil {
				return template.TemplateID, err // Return ID even on error for retry context if needed
			}
			return template.TemplateID, nil
		}
		return createRetry(d, meta, updateFunc, updatePayload, resourceZabbixTemplateRead)
		
	} else {
		// Standard update if macros were not cleared, using the helper function
		log.Printf("[DEBUG] Updating template ID %s via standard helper", d.Id())
		return createRetry(d, meta, updateTemplate, *template, resourceZabbixTemplateRead)
	}
}

func resourceZabbixTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	return api.TemplatesDeleteByIds([]string{d.Id()})
}

func createTerraformMacro(template zabbix.Template) (map[string]interface{}, error) {
	terraformMacros := make(map[string]interface{}, len(template.UserMacros))

	for _, macro := range template.UserMacros {
		var name string
		if noPrefix := strings.Split(macro.MacroName, "{$"); len(noPrefix) == 2 {
			name = noPrefix[1]
		} else {
			return nil, fmt.Errorf("Invalid macro name \"%s\"", macro.MacroName)
		}
		if noSuffix := strings.Split(name, "}"); len(noSuffix) == 2 {
			name = noSuffix[0]
		} else {
			return nil, fmt.Errorf("Invalid macro name \"%s\"", macro.MacroName)
		}
		terraformMacros[name] = macro.Value
	}
	return terraformMacros, nil
}

func createTerraformTemplateGroup(d *schema.ResourceData, api *zabbix.API) ([]string, error) {
	if api.ServerVersion.GreaterThanOrEqual(version.Must(version.NewVersion("6.2"))) {
		params := zabbix.Params{
			"output": "extend",
			"templateids": []string{
				d.Id(),
			},
		}
		groups, err := api.TemplateGroupsGet(params)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}

		groupNames := make([]string, len(groups))
		for i, g := range groups {
			groupNames[i] = g.Name
		}
		return groupNames, nil
	} else {
		params := zabbix.Params{
			"output": "extend",
			"hostids": []string{
				d.Id(),
			},
		}
		groups, err := api.HostGroupsGet(params)
		if err != nil {
			return nil, err
		}

		groupNames := make([]string, len(groups))
		for i, g := range groups {
			groupNames[i] = g.Name
		}
		return groupNames, nil
	}
}

func createTerraformLinkedTemplate(template zabbix.Template) []string {
	var terraformTemplates []string

	for _, linkedTemplate := range template.LinkedTemplates {
		terraformTemplates = append(terraformTemplates, linkedTemplate.TemplateID)
	}
	return terraformTemplates
}

func getUnlinkedTemplate(d *schema.ResourceData) zabbix.TemplateIDs {
	before, after := d.GetChange("linked_template")
	beforeID := before.(*schema.Set).List()
	afterID := after.(*schema.Set).List()
	var unlinkID zabbix.TemplateIDs

	for _, l := range beforeID {
		present := false
		for _, k := range afterID {
			if l == k {
				present = true
			}
		}
		if !present {
			unlinkID = append(unlinkID, zabbix.TemplateID{TemplateID: l.(string)})
		}
	}
	return unlinkID
}

func createTemplate(template interface{}, api *zabbix.API) (id string, err error) {
	templates := zabbix.Templates{template.(zabbix.Template)}

	err = api.TemplatesCreate(templates)
	if err != nil {
		return
	}
	id = templates[0].TemplateID
	return
}

func updateTemplate(template interface{}, api *zabbix.API) (id string, err error) {
	templates := zabbix.Templates{template.(zabbix.Template)}

	err = api.TemplatesUpdate(templates)
	if err != nil {
		return
	}
	id = templates[0].TemplateID
	return
}

func getTemplateGroups(d *schema.ResourceData, api *zabbix.API) (zabbix.HostGroupIDs, error) {
	configGroups := d.Get("groups").(*schema.Set)
	setTemplateGroups := make([]string, configGroups.Len())

	for i, g := range configGroups.List() {
		setTemplateGroups[i] = g.(string)
	}

	log.Printf("[DEBUG] Groups %v\n", setTemplateGroups)

	groupParams := zabbix.Params{
		"output": "extend",
		"filter": map[string]interface{}{
			"name": setTemplateGroups,
		},
	}

	groups, err := api.TemplateGroupsGet(groupParams)

	if err != nil {
		return nil, err
	}

	if len(groups) < configGroups.Len() {
		log.Printf("[DEBUG] Not all of the specified groups were found on zabbix server")

		for _, n := range configGroups.List() {
			found := false

			for _, g := range groups {
				if n == g.Name {
					found = true
					break
				}
			}

			if !found {
				return nil, fmt.Errorf("template group %s doesnt exist in zabbix server", n)
			}
			log.Printf("[DEBUG] %s exists on zabbix server", n)
		}
	}

	hostGroups := make(zabbix.HostGroupIDs, len(groups))

	for i, g := range groups {
		hostGroups[i] = zabbix.HostGroupID{
			GroupID: g.GroupID,
		}
	}

	return hostGroups, nil
}
