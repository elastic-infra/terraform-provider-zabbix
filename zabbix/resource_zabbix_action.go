package zabbix

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/claranet/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var StringEventTypeMap = map[string]zabbix.EventType{
	"trigger":           zabbix.TriggerEvent,
	"discovery":         zabbix.DiscoveryRuleEvent,
	"auto-registration": zabbix.AutoRegistrationEvent,
	"internal":          zabbix.InternalEvent,
}

var EventTypeStringMap = map[zabbix.EventType]string{
	zabbix.TriggerEvent:          "trigger",
	zabbix.DiscoveryRuleEvent:    "discovery",
	zabbix.AutoRegistrationEvent: "auto-registration",
	zabbix.InternalEvent:         "internal",
}

var StringActionEvaluationTypeMap = map[string]zabbix.ActionEvaluationType{
	"and/or": zabbix.AndOr,
	"and":    zabbix.And,
	"or":     zabbix.Or,
	"custom": zabbix.Custom,
}

var ActionEvaluationTypeStringMap = map[zabbix.ActionEvaluationType]string{
	zabbix.AndOr:  "and/or",
	zabbix.And:    "and",
	zabbix.Or:     "or",
	zabbix.Custom: "custom",
}

var StringActionConditionTypeMap = map[string]zabbix.ActionConditionType{
	"host_group":                  zabbix.HostGroupCondition,
	"host":                        zabbix.HostCondition,
	"trigger":                     zabbix.TriggerCondition,
	"trigger_name":                zabbix.TriggerNameCondition,
	"trigger_severity":            zabbix.TriggerSeverityCondition,
	"time_period":                 zabbix.TimePeriodCondition,
	"host_ip":                     zabbix.HostIpCondition,
	"discovered_service_type":     zabbix.DiscoveredServiceTypeCondition,
	"discovered_service_port":     zabbix.DiscoveredServicePortCondition,
	"discovery_status":            zabbix.DiscoveryStatusCondition,
	"uptime_or_downtime_duration": zabbix.UptimeOrDowntimeDurationCondition,
	"received_value":              zabbix.ReceivedValueCondition,
	"host_template":               zabbix.HostTemplateCondition,
	"application":                 zabbix.ApplicationCondition,
	"problem_is_suppressed":       zabbix.ProblemIsSuppressedCondition,
	"discovery_rule":              zabbix.DiscoveryRuleCondition,
	"discovery_check":             zabbix.DiscoveryCheckCondition,
	"proxy":                       zabbix.ProxyCondition,
	"discovery_object":            zabbix.DiscoveryObjectCondition,
	"host_name":                   zabbix.HostNameCondition,
	"event_type":                  zabbix.EventTypeCondition,
	"host_metadata":               zabbix.HostMetadataCondition,
	"event_tag":                   zabbix.EventTagCondition,
	"event_tag_value":             zabbix.EventTagValueCondition,
}

var ActionConditionTypeStringMap = map[zabbix.ActionConditionType]string{
	zabbix.HostGroupCondition:                "host_group",
	zabbix.HostCondition:                     "host",
	zabbix.TriggerCondition:                  "trigger",
	zabbix.TriggerNameCondition:              "trigger_name",
	zabbix.TriggerSeverityCondition:          "trigger_severity",
	zabbix.TimePeriodCondition:               "time_period",
	zabbix.HostIpCondition:                   "host_ip",
	zabbix.DiscoveredServiceTypeCondition:    "discovered_service_type",
	zabbix.DiscoveredServicePortCondition:    "discovered_service_port",
	zabbix.DiscoveryStatusCondition:          "discovery_status",
	zabbix.UptimeOrDowntimeDurationCondition: "uptime_or_downtime_duration",
	zabbix.ReceivedValueCondition:            "received_value",
	zabbix.HostTemplateCondition:             "host_template",
	zabbix.ApplicationCondition:              "application",
	zabbix.ProblemIsSuppressedCondition:      "problem_is_suppressed",
	zabbix.DiscoveryRuleCondition:            "discovery_rule",
	zabbix.DiscoveryCheckCondition:           "discovery_check",
	zabbix.ProxyCondition:                    "proxy",
	zabbix.DiscoveryObjectCondition:          "discovery_object",
	zabbix.HostNameCondition:                 "host_name",
	zabbix.EventTypeCondition:                "event_type",
	zabbix.HostMetadataCondition:             "host_metadata",
	zabbix.EventTagCondition:                 "event_tag",
	zabbix.EventTagValueCondition:            "event_tag_value",
}

var StringActionFilterConditionOperatorMap = map[string]zabbix.ActionFilterConditionOperator{
	"equals":                    zabbix.Equals,
	"does_not_equal":            zabbix.DoesNotEqual,
	"contains":                  zabbix.Contains,
	"does_not_contain":          zabbix.DoesNotContains,
	"in":                        zabbix.In,
	"is_greater_than_or_equals": zabbix.IsGreaterThanOrEquals,
	"is_less_than_or_equals":    zabbix.IsLessThanOrEquals,
	"not_in":                    zabbix.NotIn,
	"matches":                   zabbix.Matches,
	"does_not_match":            zabbix.DoesNotMatch,
	"yes":                       zabbix.Yes,
	"no":                        zabbix.No,
}

var ActionFilterConditionOperatorStringMap = map[zabbix.ActionFilterConditionOperator]string{
	zabbix.Equals:                "equals",
	zabbix.DoesNotEqual:          "does_not_equal",
	zabbix.Contains:              "contains",
	zabbix.DoesNotContains:       "does_not_contain",
	zabbix.In:                    "in",
	zabbix.IsGreaterThanOrEquals: "is_greater_than_or_equals",
	zabbix.IsLessThanOrEquals:    "is_less_than_or_equals",
	zabbix.NotIn:                 "not_in",
	zabbix.Matches:               "matches",
	zabbix.DoesNotMatch:          "does_not_match",
	zabbix.Yes:                   "yes",
	zabbix.No:                    "no",
}

var StringActionOperationTypeMap = map[string]zabbix.ActionOperationType{
	"send_message":            zabbix.SendMessage,
	"remote_command":          zabbix.RemoteCommand,
	"add_host":                zabbix.AddHost,
	"remove_host":             zabbix.RemoveHost,
	"add_to_host_group":       zabbix.AddToHostGroup,
	"remove_from_host_group":  zabbix.RemoveFromHostGroup,
	"link_to_template":        zabbix.LinkToTemplate,
	"unlink_from_template":    zabbix.UnlinkFromTemplate,
	"enable_host":             zabbix.EnableHost,
	"disable_host":            zabbix.DisableHost,
	"set_host_inventory_mode": zabbix.SetHostInventoryMode,
	// "notify_all_involved": ActionOperationType is different between recovery operation and update operation
}

var ActionOperationTypeStringMap = map[zabbix.ActionOperationType]string{
	zabbix.SendMessage:               "send_message",
	zabbix.RemoteCommand:             "remote_command",
	zabbix.AddHost:                   "add_host",
	zabbix.RemoveHost:                "remove_host",
	zabbix.AddToHostGroup:            "add_to_host_group",
	zabbix.RemoveFromHostGroup:       "remove_from_host_group",
	zabbix.LinkToTemplate:            "link_to_template",
	zabbix.UnlinkFromTemplate:        "unlink_from_template",
	zabbix.EnableHost:                "enable_host",
	zabbix.DisableHost:               "disable_host",
	zabbix.SetHostInventoryMode:      "set_host_inventory_mode",
	zabbix.NotifyRecoveryAllInvolved: "notify_all_involved",
	zabbix.NotifyUpdateAllInvolved:   "notify_all_involved",
}

var StringActionOperationCommandTypeMap = map[string]zabbix.ActionOperationCommandType{
	"custom_script": zabbix.CustomScript,
	"ipmi":          zabbix.IpmiCommand,
	"ssh":           zabbix.SshCommand,
	"telnet":        zabbix.TelnetCommand,
	"global_script": zabbix.GlobalScript,
}

var ActionOperationCommandTypeStringMap = map[zabbix.ActionOperationCommandType]string{
	zabbix.CustomScript:  "custom_script",
	zabbix.IpmiCommand:   "ipmi",
	zabbix.SshCommand:    "ssh",
	zabbix.TelnetCommand: "telnet",
	zabbix.GlobalScript:  "global_script",
}

var StringActionOperationCommandAuthTypeMap = map[string]zabbix.ActionOperationCommandAuthType{
	"password":   zabbix.Password,
	"public_key": zabbix.PublicKey,
}

var ActionOperationCommandAuthTypeStringMap = map[zabbix.ActionOperationCommandAuthType]string{
	zabbix.Password:  "password",
	zabbix.PublicKey: "public_key",
}

var StringActionOperationCommandExecutorTypeMap = map[string]zabbix.ActionOperationCommandExecutorType{
	"agent":  zabbix.AgentExecutor,
	"server": zabbix.ServerExecutor,
	"proxy":  zabbix.ProxyExecutor,
}

var ActionOperationCommandExecutorTypeStringMap = map[zabbix.ActionOperationCommandExecutorType]string{
	zabbix.AgentExecutor:  "agent",
	zabbix.ServerExecutor: "server",
	zabbix.ProxyExecutor:  "proxy",
}

var StringActionOperationInventoryModeMap = map[string]string{
	"manual":    "0",
	"automatic": "1",
}

var ActionOperationInventoryModeStringMap = map[string]string{
	"0": "manual",
	"1": "automatic",
}

var actionOperationCommandSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"type": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: validation.StringInSlice(
				[]string{"custom_script", "ipmi", "ssh", "telnet", "global_script"},
				false,
			),
		},
		"auth_type": {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringInSlice(
				[]string{"password", "public_key"},
				false,
			),
		},
		"execute_on": {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringInSlice(
				[]string{"agent", "server", "proxy"},
				false,
			),
		},
		"command": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"username": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"password": {
			Type:      schema.TypeString,
			Optional:  true,
			Sensitive: true,
		},
		"port": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"private_key_file": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"public_key_file": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"script_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"target": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice(
							[]string{"host_group", "host", "current_host"},
							false,
						),
					},
					"value": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
			Set: OperationCommandHash,
		},
	},
}

func OperationCommandHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-%s",
		m["type"].(string),
		m["value"].(string),
	))
	return hashcode.String(buf.String())
}

var actionOperationMessageSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"default_message": {
			Type:     schema.TypeBool,
			Optional: true,
			// FIXME: default true on Zabbix 5.0 or later
			Default: false,
		},
		"media_type_id": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "0", // NOTE: ALL
		},
		"subject": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"message": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "",
		},
		"target": {
			Type:     schema.TypeSet,
			Required: true,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"type": {
						Type:     schema.TypeString,
						Required: true,
						ValidateFunc: validation.StringInSlice(
							[]string{"user_group", "user"},
							false,
						),
					},
					"value": {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			Set: OperationMessageHash,
		},
	},
}

func OperationMessageHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-%s",
		m["type"].(string),
		m["value"].(string),
	))
	return hashcode.String(buf.String())
}

func resourceZabbixAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceZabbixActionCreate,
		Read:   resourceZabbixActionRead,
		Exists: resourceZabbixActionExists,
		Update: resourceZabbixActionUpdate,
		Delete: resourceZabbixActionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"default_step_duration": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "1h",
			},
			"event_source": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Type of events that the action will handle.",
				ValidateFunc: validation.StringInSlice(
					[]string{"trigger", "discovery", "auto-registration", "internal"},
					false,
				),
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the action.",
			},
			"default_subject": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Problem message subject. This no longer works from Zabbix 5.0 onwards.",
			},
			"default_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Problem message text. This no longer works from Zabbix 5.0 onwards.",
			},
			"recovery_subject": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Recovery message subject. This no longer works from Zabbix 5.0 onwards.",
			},
			"recovery_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Recovery message text. This no longer works from Zabbix 5.0 onwards.",
			},
			"update_subject": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Update operation message subject. This no longer works from Zabbix 5.0 onwards.",
			},
			"update_message": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "Update operation message text. This no longer works from Zabbix 5.0 onwards.",
			},
			"enabled": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"pause_in_maintenance_periods": {
				Type:     schema.TypeBool,
				Default:  true,
				Optional: true,
			},
			"calculation": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "and/or",
				ValidateFunc: validation.StringInSlice(
					[]string{"and/or", "and", "or", "custom"},
					false,
				),
			},
			"formula": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"condition": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"condition_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									"host_group",
									"host",
									"trigger",
									"trigger_name",
									"trigger_severity",
									"time_period",
									"host_ip",
									"discovered_service_type",
									"discovered_service_port",
									"discovery_status",
									"uptime_or_downtime_duration",
									"received_value",
									"host_template",
									"application",
									"problem_is_suppressed",
									"discovery_rule",
									"discovery_check",
									"proxy",
									"discovery_object",
									"host_name",
									"event_type",
									"host_metadata",
									"event_tag",
									"event_tag_value",
								},
								false,
							),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
						//"value2": {
						//	Type:     schema.TypeString,
						//	Optional: true,
						//},
						"formula_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"operator": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "equals",
							ValidateFunc: validation.StringInSlice(
								[]string{
									"equals",
									"does_not_equal",
									"contains",
									"does_not_contain",
									"in",
									"is_greater_than_or_equals",
									"is_less_than_or_equals",
									"not_in",
									"matches",
									"does_not_match",
									"yes",
									"no",
								},
								false,
							),
						},
					},
				},
			},
			// FIXME: implement opconditions
			"operation": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"operation_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									"send_message",
									"remote_command",
									"add_host",
									"remove_host",
									"add_to_host_group",
									"remove_from_host_group",
									"link_to_template",
									"unlink_from_template",
									"enable_host",
									"disable_host",
									"set_host_inventory_mode",
								},
								false,
							),
						},
						"step_duration": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "0",
						},
						"step_from": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"step_to": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
						"command": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem:     actionOperationCommandSchema,
						},
						"host_groups": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"message": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem:     actionOperationMessageSchema,
						},
						"templates": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
						},
						"inventory_mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice(
								[]string{"manual", "automatic"},
								false,
							),
						},
					},
				},
			},
			"recovery_operation": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"operation_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									"send_message",
									"remote_command",
									"notify_all_involved",
								},
								false,
							),
						},
						"command": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem:     actionOperationCommandSchema,
						},
						"message": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem:     actionOperationMessageSchema,
						},
					},
				},
			},
			"update_operation": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"operation_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice(
								[]string{
									"send_message",
									"remote_command",
									"notify_all_involved",
								},
								false,
							),
						},
						"command": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem:     actionOperationCommandSchema,
						},
						"message": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem:     actionOperationMessageSchema,
						},
					},
				},
			},
		},
	}
}

func createActionObject(d *schema.ResourceData, api *zabbix.API) (*zabbix.Action, error) {
	status := zabbix.Disabled
	if d.Get("enabled").(bool) {
		status = zabbix.Enabled
	}

	ope, err := createActionOperationObject(d.Id(), d.Get("operation").([]interface{}), api)
	if err != nil {
		return nil, err
	}

	recOpe, err := createActionRecoveryOperationObject(d.Id(), d.Get("recovery_operation").([]interface{}), api)
	if err != nil {
		return nil, err
	}

	upOpe, err := createActionUpdateOperationObject(d.Get("update_operation").([]interface{}), api)
	if err != nil {
		return nil, err
	}

	eventSource := StringEventTypeMap[d.Get("event_source").(string)]

	action := zabbix.Action{
		ActionID:    d.Id(),
		EventSource: eventSource,
		Name:        d.Get("name").(string),
		Status:      status,
		Filter: zabbix.ActionFilter{
			Conditions:     createActionConditionObject(d.Get("condition").([]interface{})),
			EvaluationType: StringActionEvaluationTypeMap[d.Get("calculation").(string)],
			Formula:        d.Get("formula").(string),
		},
		Operations:         ope,
		RecoveryOperations: recOpe,
		UpdateOperations:   upOpe,
	}

	// NOTE: pause_suppressed set only TriggerEvent
	if eventSource == zabbix.TriggerEvent {
		if d.Get("pause_in_maintenance_periods").(bool) {
			action.PauseSuppressed = pauseType(zabbix.Pause)
		} else {
			action.PauseSuppressed = pauseType(zabbix.DontPause)
		}
	}

	return &action, nil
}

// FIXME: Several type need to convert from value to ID
func createActionConditionObject(lst []interface{}) (items zabbix.ActionFilterConditions) {
	for _, v := range lst {
		m := v.(map[string]interface{})

		item := zabbix.ActionFilterCondition{
			ConditionID:   m["condition_id"].(string),
			ConditionType: StringActionConditionTypeMap[m["type"].(string)],
			Value:         convertValuetoId(m["value"].(string), m["type"].(string)),
			//Value2:        m["value2"].(string),
			FormulaID: m["formula_id"].(string),
			Operator:  StringActionFilterConditionOperatorMap[m["operator"].(string)],
		}
		items = append(items, item)
	}

	return
}

func convertValuetoId(value string, condition string) string {

	if condition == "host_template" {
		value = ""
	}

	return value
}

func createActionOperationObject(id string, lst []interface{}, api *zabbix.API) (items zabbix.ActionOperations, err error) {
	for _, v := range lst {
		m := v.(map[string]interface{})
		opeId := m["operation_id"].(string)

		cmd, cmdHostGroups, cmdHosts, err := createActionOperationCommand(opeId, m["command"].([]interface{}), api)
		if err != nil {
			return nil, err
		}

		hostGroups, err := createActionOperationHostGroups(opeId, m["host_groups"].(*schema.Set).List(), api)
		if err != nil {
			return nil, err
		}

		msg, msgUserGroups, msgUsers, err := createActionOperationMessage(opeId, m["message"].([]interface{}), api)
		if err != nil {
			return nil, err
		}

		templates, err := createActionOperationTemplates(opeId, m["templates"].(*schema.Set).List(), api)
		if err != nil {
			return nil, err
		}

		var inventory *zabbix.ActionOperationInventory
		if mode := m["inventory_mode"].(string); mode != "" {
			inventory = &zabbix.ActionOperationInventory{
				OperationID:   opeId,
				InventoryMode: StringActionOperationInventoryModeMap[mode],
			}
		}
		// ONAIMAT
		item := zabbix.ActionOperation{
			//OperationID:   opeId,
			OperationType: StringActionOperationTypeMap[m["type"].(string)],
			//ActionID:      id,
			//Period:            m["step_duration"].(string),
			//StepFrom:          m["step_from"].(int),
			//StepTo:            m["step_to"].(int),
			Command:           cmd,
			CommandHostGroups: cmdHostGroups,
			CommandHosts:      cmdHosts,
			HostGroups:        hostGroups,
			Message:           msg,
			MessageUserGroups: msgUserGroups,
			MessageUsers:      msgUsers,
			Templates:         templates,
			Inventory:         inventory,
		}
		items = append(items, item)
	}

	return
}

func createActionRecoveryOperationObject(id string, lst []interface{}, api *zabbix.API) (items zabbix.ActionRecoveryOperations, err error) {
	for _, v := range lst {
		m := v.(map[string]interface{})
		opeId := m["operation_id"].(string)

		cmd, cmdHostGroups, cmdHosts, err := createActionOperationCommand(opeId, m["command"].([]interface{}), api)
		if err != nil {
			return nil, err
		}

		msg, msgUserGroups, msgUsers, err := createActionOperationMessage(opeId, m["message"].([]interface{}), api)
		if err != nil {
			return nil, err
		}

		t := m["type"].(string)
		opeType := StringActionOperationTypeMap[t]
		if t == "notify_all_involved" {
			opeType = zabbix.NotifyRecoveryAllInvolved
		}

		item := zabbix.ActionRecoveryOperation{
			OperationID:       opeId,
			OperationType:     opeType,
			ActionID:          id,
			Command:           cmd,
			CommandHostGroups: cmdHostGroups,
			CommandHosts:      cmdHosts,
			Message:           msg,
			MessageUserGroups: msgUserGroups,
			MessageUsers:      msgUsers,
		}
		items = append(items, item)
	}

	return
}

func createActionUpdateOperationObject(lst []interface{}, api *zabbix.API) (items zabbix.ActionUpdateOperations, err error) {
	for _, v := range lst {
		m := v.(map[string]interface{})
		opeId := m["operation_id"].(string)

		cmd, cmdHostGroups, cmdHosts, err := createActionOperationCommand(opeId, m["command"].([]interface{}), api)
		if err != nil {
			return nil, err
		}

		msg, msgUserGroups, msgUsers, err := createActionOperationMessage(opeId, m["message"].([]interface{}), api)
		if err != nil {
			return nil, err
		}

		t := m["type"].(string)
		opeType := StringActionOperationTypeMap[t]
		if t == "notify_all_involved" {
			opeType = zabbix.NotifyUpdateAllInvolved
		}

		item := zabbix.ActionUpdateOperation{
			//OperationID:       opeId,
			OperationType:     opeType,
			Command:           cmd,
			CommandHostGroups: cmdHostGroups,
			CommandHosts:      cmdHosts,
			Message:           msg,
			MessageUserGroups: msgUserGroups,
			MessageUsers:      msgUsers,
		}
		items = append(items, item)
	}

	return
}

func createActionOperationCommand(id string, lst []interface{}, api *zabbix.API) (
	cmd *zabbix.ActionOperationCommand,
	groups zabbix.ActionOperationCommandHostGroups,
	hosts zabbix.ActionOperationCommandHosts,
	err error) {
	if len(lst) == 0 {
		return
	}
	m := lst[0].(map[string]interface{})

	cmd = &zabbix.ActionOperationCommand{
		OperationID: id,
		Type:        StringActionOperationCommandTypeMap[m["type"].(string)],
		Command:     m["command"].(string),
		AuthType:    actionOperationCommandAuthType(m["auth_type"].(string)),
		ExecuteOn:   actionOperationCommandExecutorType(m["execute_on"].(string)),
		Username:    m["username"].(string),
		Password:    m["password"].(string),
		Port:        m["port"].(string),
		PrivateKey:  m["private_key_file"].(string),
		PublicKey:   m["public_key_file"].(string),
		ScriptID:    m["script_id"].(string),
	}

	targets := m["target"].(*schema.Set).List()

	var groupNames []string
	var hostNames []string
	for _, t := range targets {
		target := t.(map[string]interface{})
		switch target["type"].(string) {
		case "host_group":
			groupNames = append(groupNames, target["value"].(string))
		case "host":
			hostNames = append(hostNames, target["value"].(string))
		}
	}

	groupMap := map[string]string{}
	if len(groupNames) > 0 {
		params := zabbix.Params{
			"output": []string{"name"},
			"filter": map[string]interface{}{
				"name": groupNames,
			},
		}

		res, err := api.HostGroupsGet(params)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, g := range res {
			groupMap[g.Name] = g.GroupID
		}
	}

	hostMap := map[string]string{}
	if len(hostNames) > 0 {
		params := zabbix.Params{
			"output": []string{"host"},
			"filter": map[string]interface{}{
				"host": hostNames,
			},
		}

		res, err := api.HostsGet(params)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, h := range res {
			hostMap[h.Host] = h.HostID
		}
	}

	for _, t := range targets {
		target := t.(map[string]interface{})
		switch target["type"].(string) {
		case "host_group":
			groups = append(groups, zabbix.ActionOperationCommandHostGroup{
				OperationID: id,
				GroupID:     groupMap[target["value"].(string)],
			})
		case "host":
			hosts = append(hosts, zabbix.ActionOperationCommandHost{
				OperationID: id,
				HostID:      hostMap[target["value"].(string)],
			})
		case "current_host":
			hosts = append(hosts, zabbix.ActionOperationCommandHost{
				OperationID: id,
				HostID:      "0",
			})
		}
	}

	return
}

func createActionOperationHostGroups(id string, lst []interface{}, api *zabbix.API) (
	groups zabbix.ActionOperationHostGroups,
	err error) {
	if len(lst) == 0 {
		return
	}

	var groupNames []string
	for _, g := range lst {
		groupNames = append(groupNames, g.(string))
	}

	params := zabbix.Params{
		"output": []string{},
		"filter": map[string]interface{}{
			"name": groupNames,
		},
	}

	res, err := api.HostGroupsGet(params)
	if err != nil {
		return nil, err
	}
	for _, g := range res {
		groups = append(groups, zabbix.ActionOperationHostGroup{
			//OperationID: id,
			GroupID: g.GroupID,
		})
	}

	return
}

func createActionOperationMessage(id string, lst []interface{}, api *zabbix.API) (
	msg *zabbix.ActionOperationMessage,
	groups zabbix.ActionOperationMessageUserGroups,
	users zabbix.ActionOperationMessageUsers,
	err error) {
	if len(lst) == 0 {
		return
	}
	m := lst[0].(map[string]interface{})

	defMsg := "0"
	if m["default_message"].(bool) {
		defMsg = "1"
	}

	msg = &zabbix.ActionOperationMessage{
		OperationID:    id,
		DefaultMessage: defMsg,
		MediaTypeID:    m["media_type_id"].(string),
		Message:        m["message"].(string),
		Subject:        m["subject"].(string),
	}

	targets := m["target"].(*schema.Set).List()

	var groupNames []string
	var userNames []string
	for _, t := range targets {
		target := t.(map[string]interface{})
		switch target["type"].(string) {
		case "user_group":
			groupNames = append(groupNames, target["value"].(string))
		case "user":
			userNames = append(userNames, target["value"].(string))
		}
	}

	if len(groupNames) > 0 {
		params := zabbix.Params{
			"output": []string{},
			"filter": map[string]interface{}{
				"name": groupNames,
			},
		}

		res, err := api.UserGroupsGet(params)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, g := range res {
			groups = append(groups, zabbix.ActionOperationMessageUserGroup{
				OperationID: id,
				UserGroupID: g.GroupID,
			})
		}
	}

	if len(userNames) > 0 {
		params := zabbix.Params{
			"output": []string{},
			"filter": map[string]interface{}{
				"alias":    userNames,
				"username": userNames,
			},
		}

		res, err := api.UsersGet(params)
		if err != nil {
			return nil, nil, nil, err
		}
		for _, u := range res {
			users = append(users, zabbix.ActionOperationMessageUser{
				OperationID: id,
				UserID:      u.UserID,
			})
		}
	}

	return
}

func createActionOperationTemplates(id string, lst []interface{}, api *zabbix.API) (
	templates zabbix.ActionOperationTemplates,
	err error) {
	if len(lst) == 0 {
		return
	}

	var templateNames []string
	for _, g := range lst {
		templateNames = append(templateNames, g.(string))
	}

	params := zabbix.Params{
		"output": []string{},
		"filter": map[string]interface{}{
			"host": templateNames,
		},
	}

	res, err := api.TemplatesGet(params)
	if err != nil {
		return nil, err
	}
	for _, t := range res {
		templates = append(templates, zabbix.ActionOperationTemplate{
			//OperationID: id,
			TemplateID: t.TemplateID,
		})
	}
	return
}

func resourceZabbixActionCreate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	action, err := createActionObject(d, api)

	if err != nil {
		return err
	}

	actions := zabbix.Actions{*action}

	err = api.ActionsCreate(actions)

	if err != nil {
		return err
	}

	id := actions[0].ActionID
	d.SetId(id)

	return resourceZabbixActionRead(d, meta)
}

func resourceZabbixActionRead(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	action, err := api.ActionGetByID(d.Id())
	if err != nil {
		return err
	}

	//d.Set("default_step_duration", action.Period)
	d.Set("event_source", EventTypeStringMap[action.EventSource])
	d.Set("name", action.Name)
	//d.Set("default_subject", action.DefaultSubject)
	//d.Set("default_message", action.DefaultMessage)
	//d.Set("recovery_subject", action.RecoverySubject)
	//d.Set("recovery_message", action.RecoveryMessage)
	//d.Set("update_subject", action.AckSubject)
	//d.Set("update_message", action.AckMessage)
	d.Set("enabled", action.Status == zabbix.Enabled)
	d.Set("pause_in_maintenance_periods", pauseTypeValue(action.PauseSuppressed) == zabbix.Pause)
	d.Set("calculation", ActionEvaluationTypeStringMap[action.Filter.EvaluationType])
	d.Set("formula", action.Filter.EvaluationType)

	conditions, err := readActionConditions(action.Filter.Conditions, api)
	if err != nil {
		return err
	}
	d.Set("condition", conditions)

	operations, err := readActionOperations(action.Operations, api)
	if err != nil {
		return err
	}
	d.Set("operation", operations)

	recOpe, err := readActionRecoveryOperations(action.RecoveryOperations, api)
	if err != nil {
		return err
	}
	d.Set("recovery_operation", recOpe)

	upOpe, err := readActionUpdateOperations(action.UpdateOperations, api)
	if err != nil {
		return err
	}
	d.Set("update_operation", upOpe)

	log.Printf("[DEBUG] Action name is %s\n", action.Name)
	return nil
}

func readActionConditions(cds zabbix.ActionFilterConditions, api *zabbix.API) (lst []interface{}, err error) {
	for _, v := range cds {
		m := map[string]interface{}{}
		m["condition_id"] = v.ConditionID
		m["type"] = ActionConditionTypeStringMap[v.ConditionType]
		// FIXME: Several type need to convert from ID to value
		m["value"] = v.Value
		//m["value2"] = v.Value2
		m["formula_id"] = v.FormulaID
		m["operator"] = ActionFilterConditionOperatorStringMap[v.Operator]

		lst = append(lst, m)
	}
	return
}

func readActionOperations(ops zabbix.ActionOperations, api *zabbix.API) (lst []interface{}, err error) {
	for _, v := range ops {
		m := map[string]interface{}{}
		m["operation_id"] = v.OperationID
		m["type"] = ActionOperationTypeStringMap[v.OperationType]
		//m["step_duration"] = v.Period
		//m["step_from"] = v.StepFrom
		//m["step_to"] = v.StepTo

		commands, err := readActionOperationCommands(v.Command, v.CommandHostGroups, v.CommandHosts, api)
		if err != nil {
			return nil, err
		}
		m["command"] = commands

		hostGroups, err := readActionOperationHostGroups(v.HostGroups, api)
		if err != nil {
			return nil, err
		}
		m["host_groups"] = hostGroups

		messages, err := readActionOperationMessages(v.Message, v.MessageUserGroups, v.MessageUsers, api)
		if err != nil {
			return nil, err
		}
		m["message"] = messages

		templates, err := readActionOperationTemplates(v.Templates, api)
		if err != nil {
			return nil, err
		}
		m["templates"] = templates

		if v.Inventory != nil {
			m["inventory_mode"] = ActionOperationInventoryModeStringMap[v.Inventory.InventoryMode]
		}

		lst = append(lst, m)
	}

	return
}

func readActionRecoveryOperations(ops zabbix.ActionRecoveryOperations, api *zabbix.API) (lst []interface{}, err error) {
	for _, v := range ops {
		m := map[string]interface{}{}
		m["operation_id"] = v.OperationID
		m["type"] = ActionOperationTypeStringMap[v.OperationType]

		commands, err := readActionOperationCommands(v.Command, v.CommandHostGroups, v.CommandHosts, api)
		if err != nil {
			return nil, err
		}
		m["command"] = commands

		messages, err := readActionOperationMessages(v.Message, v.MessageUserGroups, v.MessageUsers, api)
		if err != nil {
			return nil, err
		}
		m["message"] = messages

		lst = append(lst, m)
	}

	return
}

func readActionUpdateOperations(ops zabbix.ActionUpdateOperations, api *zabbix.API) (lst []interface{}, err error) {
	for _, v := range ops {
		m := map[string]interface{}{}
		//m["operation_id"] = v.OperationID
		m["type"] = ActionOperationTypeStringMap[v.OperationType]

		commands, err := readActionOperationCommands(v.Command, v.CommandHostGroups, v.CommandHosts, api)
		if err != nil {
			return nil, err
		}
		m["command"] = commands

		messages, err := readActionOperationMessages(v.Message, v.MessageUserGroups, v.MessageUsers, api)
		if err != nil {
			return nil, err
		}
		m["message"] = messages

		lst = append(lst, m)
	}

	return
}

func readActionOperationCommands(
	cmd *zabbix.ActionOperationCommand,
	groups zabbix.ActionOperationCommandHostGroups,
	hosts zabbix.ActionOperationCommandHosts,
	api *zabbix.API) (lst []interface{}, err error) {
	if cmd == nil {
		return
	}

	m := map[string]interface{}{}
	m["type"] = ActionOperationCommandTypeStringMap[cmd.Type]
	m["auth_type"] = actionOperationCommandAuthTypeString(cmd.AuthType)
	m["execute_on"] = actionOperationCommandExecutorTypeString(cmd.ExecuteOn)
	m["command"] = cmd.Command
	m["username"] = cmd.Username
	m["password"] = cmd.Password
	m["port"] = cmd.Port
	m["private_key_file"] = cmd.PrivateKey
	m["public_key_file"] = cmd.PublicKey
	m["script_id"] = cmd.ScriptID

	target, err := readActionOperationCommandTargets(groups, hosts, api)
	if err != nil {
		return nil, err
	}
	m["target"] = target

	lst = append(lst, m)
	return
}

func readActionOperationCommandTargets(
	groups zabbix.ActionOperationCommandHostGroups,
	hosts zabbix.ActionOperationCommandHosts,
	api *zabbix.API) (lst []interface{}, err error) {
	if len(groups) > 0 {
		var groupIds []string
		for _, g := range groups {
			groupIds = append(groupIds, g.GroupID)
		}
		params := zabbix.Params{
			"groupids": groupIds,
			"output":   []string{"name"},
		}
		res, err := api.HostGroupsGet(params)
		if err != nil {
			return nil, err
		}
		groupMap := map[string]string{}
		for _, g := range res {
			groupMap[g.GroupID] = g.Name
		}
		for _, g := range groups {
			m := map[string]interface{}{}
			m["type"] = "host_group"
			m["value"] = groupMap[g.GroupID]
			lst = append(lst, m)
		}
	}

	if len(hosts) > 0 {
		var hostIds []string
		for _, h := range hosts {
			if h.HostID == "0" {
				continue
			}
			hostIds = append(hostIds, h.HostID)
		}
		params := zabbix.Params{
			"hostids": hostIds,
			"output":  []string{"host"},
		}
		res, err := api.HostsGet(params)
		if err != nil {
			return nil, err
		}
		hostMap := map[string]string{}
		for _, h := range res {
			hostMap[h.HostID] = h.Host
		}
		for _, h := range hosts {
			m := map[string]interface{}{}
			if h.HostID == "0" {
				m["type"] = "current_host"
			} else {
				m["type"] = "host"
				m["value"] = hostMap[h.HostID]
			}

			lst = append(lst, m)
		}
	}

	return
}

func readActionOperationHostGroups(grps zabbix.ActionOperationHostGroups, api *zabbix.API) (
	lst []interface{},
	err error) {
	if len(grps) == 0 {
		return
	}

	var groupIds []string
	for _, g := range grps {
		groupIds = append(groupIds, g.GroupID)
	}

	params := zabbix.Params{
		"groupids": groupIds,
		"output":   []string{"name"},
	}

	res, err := api.HostGroupsGet(params)
	if err != nil {
		return nil, err
	}

	for _, g := range res {
		lst = append(lst, g.Name)
	}

	return
}

func readActionOperationMessages(
	msg *zabbix.ActionOperationMessage,
	groups zabbix.ActionOperationMessageUserGroups,
	users zabbix.ActionOperationMessageUsers,
	api *zabbix.API) (lst []interface{}, err error) {
	if msg == nil {
		return
	}

	m := map[string]interface{}{}
	m["default_message"] = msg.DefaultMessage == "1"
	m["media_type_id"] = msg.MediaTypeID
	m["subject"] = msg.Subject
	m["message"] = msg.Message

	target, err := readActionOperationMessageTargets(groups, users, api)
	if err != nil {
		return nil, err
	}
	m["target"] = target

	lst = append(lst, m)
	return
}

func readActionOperationMessageTargets(
	groups zabbix.ActionOperationMessageUserGroups,
	users zabbix.ActionOperationMessageUsers,
	api *zabbix.API) (lst []interface{}, err error) {
	if len(groups) > 0 {
		var groupIds []string
		for _, g := range groups {
			groupIds = append(groupIds, g.UserGroupID)
		}
		params := zabbix.Params{
			"groupids": groupIds,
			"output":   []string{"name"},
		}
		res, err := api.UserGroupsGet(params)
		if err != nil {
			return nil, err
		}
		for _, g := range res {
			m := map[string]interface{}{}
			m["type"] = "user_group"
			m["value"] = g.Name
			lst = append(lst, m)
		}
	}

	if len(users) > 0 {
		var userIds []string
		for _, u := range users {
			userIds = append(userIds, u.UserID)
		}
		params := zabbix.Params{
			"userids": userIds,
			"output":  []string{"alias", "username"},
		}
		res, err := api.UsersGet(params)
		if err != nil {
			return nil, err
		}
		for _, u := range res {
			m := map[string]interface{}{}
			m["type"] = "user"
			if u.Alias != "" {
				m["value"] = u.Alias
			} else {
				m["value"] = u.Username
			}
			lst = append(lst, m)
		}
	}

	return
}

func readActionOperationTemplates(templates zabbix.ActionOperationTemplates, api *zabbix.API) (
	lst []interface{},
	err error) {

	if len(templates) == 0 {
		return
	}

	var templateIds []string
	for _, t := range templates {
		templateIds = append(templateIds, t.TemplateID)
	}

	params := zabbix.Params{
		"templateids": templateIds,
		"output":      []string{"host"},
	}

	res, err := api.TemplatesGet(params)
	if err != nil {
		return nil, err
	}

	for _, t := range res {
		lst = append(lst, t.Host)
	}

	return
}

func resourceZabbixActionExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	api := meta.(*zabbix.API)

	_, err := api.ActionGetByID(d.Id())
	if err != nil {
		if strings.Contains(err.Error(), "Expected exactly one result") {
			log.Printf("[DEBUG] Action with id %s doesn't exist", d.Id())
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func resourceZabbixActionUpdate(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	action, err := createActionObject(d, api)

	if err != nil {
		return err
	}

	// NOTE: EventSource can't be updated
	action.EventSource = ""

	actions := zabbix.Actions{*action}

	err = api.ActionsUpdate(actions)

	if err != nil {
		return err
	}

	return resourceZabbixActionRead(d, meta)
}

func resourceZabbixActionDelete(d *schema.ResourceData, meta interface{}) error {
	api := meta.(*zabbix.API)

	err := api.ActionsDeleteByIds([]string{d.Id()})

	if err != nil {
		return err
	}

	return nil
}

func pauseType(v zabbix.PauseType) *zabbix.PauseType {
	return &v
}

func pauseTypeValue(v *zabbix.PauseType) zabbix.PauseType {
	if v != nil {
		return *v
	}
	return zabbix.DontPause
}

func actionOperationCommandAuthType(v string) *zabbix.ActionOperationCommandAuthType {
	authType := StringActionOperationCommandAuthTypeMap[v]
	return &authType
}

func actionOperationCommandAuthTypeString(v *zabbix.ActionOperationCommandAuthType) string {
	var authType zabbix.ActionOperationCommandAuthType
	if v != nil {
		authType = *v
	} else {
		authType = zabbix.Password
	}
	return ActionOperationCommandAuthTypeStringMap[authType]
}

func actionOperationCommandExecutorType(v string) *zabbix.ActionOperationCommandExecutorType {
	executorType := StringActionOperationCommandExecutorTypeMap[v]
	return &executorType
}

func actionOperationCommandExecutorTypeString(v *zabbix.ActionOperationCommandExecutorType) string {
	var executorType zabbix.ActionOperationCommandExecutorType
	if v != nil {
		executorType = *v
	} else {
		executorType = zabbix.AgentExecutor
	}
	return ActionOperationCommandExecutorTypeStringMap[executorType]
}
