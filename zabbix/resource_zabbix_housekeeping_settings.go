package zabbix

import (
	"context"
	"github.com/atypon/go-zabbix-api"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceZabbixHousekeepingSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceZabbixHousekeepingCreate,
		ReadContext:   resourceZabbixHousekeepingRead,
		UpdateContext: resourceZabbixHousekeepingUpdate,
		DeleteContext: resourceZabbixHousekeepingDelete,
		Schema: map[string]*schema.Schema{
			"events_mode": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"events_trigger": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"events_internal": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			//"events_service": {
			//	Type:     schema.TypeString,
			//	Optional: true,
			//},
			"events_discovery": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"events_autoreg": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"services_mode": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"services": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"audit_mode": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"audit": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"sessions_mode": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"sessions": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"history_mode": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"history_global": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"history": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"trends_mode": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"trends_global": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"trends": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"db_extension": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"compression_status": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"compress_older": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"compression_availability": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceZabbixHousekeepingUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*zabbix.API)
	housekeepingSettings := createHousekeepingObjectFromResourceData(data)
	err := api.HousekeepingSet(housekeepingSettings)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceZabbixHousekeepingDelete(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	warning := diag.Diagnostic{
		Summary: "The housekeeping_settings resource is a singleton that can only be read and updated, " +
			"this will only delete the resource from terraform state",
		Severity: diag.Warning,
	}
	data.SetId("")
	return diag.Diagnostics{warning}
}

func resourceZabbixHousekeepingCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	warning := diag.Diagnostic{
		Summary: "The housekeeping_settings resource is a singleton that can only be read and updated, " +
			"this will read the resource into terraform state and update remote values to match your code",
		Severity: diag.Warning,
	}
	readDiags := resourceZabbixHousekeepingRead(ctx, data, meta)
	if readDiags.HasError() {
		return readDiags
	}
	updateDiags := resourceZabbixHousekeepingUpdate(ctx, data, meta)
	data.SetId("housekeeping_settings")
	return append(updateDiags, warning)
}

func resourceZabbixHousekeepingRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var errors TerraformErrors
	api := meta.(*zabbix.API)
	housekeeping, err := api.HousekeepingGet()
	if err != nil {
		return diag.FromErr(err)
	}
	errors.addError(data.Set("events_mode", housekeeping.EventsMode))
	errors.addError(data.Set("events_internal", housekeeping.EventsDataStoragePeriod))
	errors.addError(data.Set("events_trigger", housekeeping.EventsTriggerStoragePeriod))
	//errors.addError(data.Set("events_service", housekeeping.EventsService))
	errors.addError(data.Set("events_discovery", housekeeping.EventsDiscoveryPeriod))
	errors.addError(data.Set("events_autoreg", housekeeping.EventsAutoregPeriod))
	errors.addError(data.Set("services_mode", housekeeping.ServicesMode))
	errors.addError(data.Set("services", housekeeping.ServicesDataStoragePeriod))
	errors.addError(data.Set("audit_mode", housekeeping.AuditMode))
	errors.addError(data.Set("audit", housekeeping.AuditStoragePeriod))
	errors.addError(data.Set("sessions_mode", housekeeping.SessionsMode))
	errors.addError(data.Set("sessions", housekeeping.SessionsStoragePeriod))
	errors.addError(data.Set("history_mode", housekeeping.HistoryMode))
	errors.addError(data.Set("history_global", housekeeping.HistoryGlobal))
	errors.addError(data.Set("history", housekeeping.HistoryStoragePeriod))
	errors.addError(data.Set("trends_mode", housekeeping.TrendsMode))
	errors.addError(data.Set("trends_global", housekeeping.TrendsGlobal))
	errors.addError(data.Set("trends", housekeeping.TrendsStoragePeriod))
	errors.addError(data.Set("db_extension", housekeeping.DBExtension))
	errors.addError(data.Set("compression_status", housekeeping.CompressionStatus))
	errors.addError(data.Set("compress_older", housekeeping.CompressOlderThan))
	errors.addError(data.Set("compression_availability", housekeeping.CompressionAvailability))
	return errors.getDiagnostics()
}

func createHousekeepingObjectFromResourceData(data *schema.ResourceData) *zabbix.HousekeepingSettings {
	return &zabbix.HousekeepingSettings{
		EventsMode:                 data.Get("events_mode").(int),
		EventsDataStoragePeriod:    data.Get("events_internal").(string),
		EventsTriggerStoragePeriod: data.Get("events_trigger").(string),
		EventsDiscoveryPeriod:      data.Get("events_discovery").(string),
		EventsAutoregPeriod:        data.Get("events_autoreg").(string),
		ServicesMode:               data.Get("services_mode").(int),
		ServicesDataStoragePeriod:  data.Get("services").(string),
		AuditMode:                  data.Get("audit_mode").(int),
		AuditStoragePeriod:         data.Get("audit").(string),
		SessionsMode:               data.Get("sessions_mode").(int),
		SessionsStoragePeriod:      data.Get("sessions").(string),
		HistoryMode:                data.Get("history_mode").(int),
		HistoryGlobal:              data.Get("history_global").(int),
		HistoryStoragePeriod:       data.Get("history").(string),
		TrendsMode:                 data.Get("trends_mode").(int),
		TrendsGlobal:               data.Get("trends_global").(int),
		TrendsStoragePeriod:        data.Get("trends").(string),
		DBExtension:                data.Get("db_extension").(string),
		CompressionStatus:          data.Get("compression_status").(int),
		CompressOlderThan:          data.Get("compress_older").(string),
		CompressionAvailability:    data.Get("compression_availability").(int),
	}
}
