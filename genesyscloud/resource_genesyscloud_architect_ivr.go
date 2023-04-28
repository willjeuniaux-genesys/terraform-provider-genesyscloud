package genesyscloud

import (
	"context"
	"fmt"
	"log"
	"terraform-provider-genesyscloud/genesyscloud/consistency_checker"
	"terraform-provider-genesyscloud/genesyscloud/proxies/architect_api"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/mypurecloud/platform-client-sdk-go/v99/platformclientv2"
)

const maxDnisPerRequest = 50

var architectIvrProxy *architect_api.ArchitectIvrProxy

func init() {
	architectIvrProxy = architect_api.NewArchitectIvrProxy()
}

func getAllIvrConfigs(_ context.Context, clientConfig *platformclientv2.Configuration) (ResourceIDMetaMap, diag.Diagnostics) {
	resources := make(ResourceIDMetaMap)
	architectIvrProxy.ConfigureProxyApiInstance(clientConfig)

	for pageNum := 1; ; pageNum++ {
		const pageSize = 100
		ivrConfigs, _, getErr := architectIvrProxy.GetArchitectIvrs(architectIvrProxy.Api, pageNum, pageSize, "", "", "", "", "")
		if getErr != nil {
			return nil, diag.Errorf("Failed to get page of IVR configs: %v", getErr)
		}

		if ivrConfigs.Entities == nil || len(*ivrConfigs.Entities) == 0 {
			break
		}

		for _, ivrConfig := range *ivrConfigs.Entities {
			if ivrConfig.State != nil && *ivrConfig.State != "deleted" {
				resources[*ivrConfig.Id] = &ResourceMeta{Name: *ivrConfig.Name}
			}
		}
	}

	return resources, nil
}

func architectIvrExporter() *ResourceExporter {
	return &ResourceExporter{
		GetResourcesFunc: getAllWithPooledClient(getAllIvrConfigs),
		RefAttrs: map[string]*RefAttrSettings{
			"open_hours_flow_id":    {RefType: "genesyscloud_flow"},
			"closed_hours_flow_id":  {RefType: "genesyscloud_flow"},
			"holiday_hours_flow_id": {RefType: "genesyscloud_flow"},
			"schedule_group_id":     {RefType: "genesyscloud_architect_schedulegroups"},
			"division_id":           {RefType: "genesyscloud_auth_division"},
		},
	}
}

func resourceArchitectIvrConfig() *schema.Resource {
	return &schema.Resource{
		Description: "Genesys Cloud IVR config",

		CreateContext: createWithPooledClient(createIvrConfig),
		ReadContext:   readWithPooledClient(readIvrConfig),
		UpdateContext: updateWithPooledClient(updateIvrConfig),
		DeleteContext: deleteWithPooledClient(deleteIvrConfig),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name of the IVR config. Note: If the name changes, the existing Genesys Cloud IVR config will be dropped and recreated with a new ID. This can cause an Architect Flow to become invalid if the old flow is reference in the flow.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Description: "IVR Config description.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"dnis": {
				Description: fmt.Sprintf("The phone number(s) to contact the IVR by. (Note: An array with a length greater than %v will be broken into chunks and uploaded in subsequent PUT requests.)", maxDnisPerRequest),
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateDiagFunc: validatePhoneNumber},
			},
			"open_hours_flow_id": {
				Description: "ID of inbound call flow for open hours.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"closed_hours_flow_id": {
				Description: "ID of inbound call flow for closed hours.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"holiday_hours_flow_id": {
				Description: "ID of inbound call flow for holidays.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"schedule_group_id": {
				Description: "Schedule group ID.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"division_id": {
				Description: "Division ID.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func createIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	openHoursFlowId := buildSdkDomainEntityRef(d, "open_hours_flow_id")
	closedHoursFlowId := buildSdkDomainEntityRef(d, "closed_hours_flow_id")
	holidayHoursFlowId := buildSdkDomainEntityRef(d, "holiday_hours_flow_id")
	scheduleGroupId := buildSdkDomainEntityRef(d, "schedule_group_id")
	divisionId := d.Get("division_id").(string)
	dnis := buildSdkStringList(d, "dnis")

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectIvrProxy.ConfigureProxyApiInstance(sdkConfig)

	ivrBody := platformclientv2.Ivr{
		Name:             &name,
		OpenHoursFlow:    openHoursFlowId,
		ClosedHoursFlow:  closedHoursFlowId,
		HolidayHoursFlow: holidayHoursFlowId,
		ScheduleGroup:    scheduleGroupId,
	}

	if description != "" {
		ivrBody.Description = &description
	}

	if divisionId != "" {
		ivrBody.Division = &platformclientv2.Writabledivision{Id: &divisionId}
	}

	var dnisChunks [][]string
	if dnis != nil {
		dnisChunks = chunkSlice(*dnis, maxDnisPerRequest)
		if len(dnisChunks) == 1 {
			ivrBody.Dnis = &dnisChunks[0]
		}
	}

	// It might need to wait for a dependent did_pool to be created to avoid an eventual consistency issue which
	// would result in the error "Field 'didPoolId' is required and cannot be empty."
	if ivrBody.Dnis != nil {
		time.Sleep(3 * time.Second)
	}
	log.Printf("Creating IVR config %s", name)
	ivrConfig, _, err := architectIvrProxy.PostArchitectIvr(architectIvrProxy.Api, &ivrBody)
	if err != nil {
		return diag.Errorf("Failed to create IVR config %s: %s", name, err)
	}

	d.SetId(*ivrConfig.Id)

	if len(dnisChunks) > 1 {
		if _, _, err := architectIvrProxy.UploadIvrDnisChunks(architectIvrProxy, dnisChunks, d.Id()); err != nil {
			return diag.Errorf("%v", err)
		}
	}

	log.Printf("Created IVR config %s %s", name, *ivrConfig.Id)
	return readIvrConfig(ctx, d, meta)
}

func readIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectIvrProxy.ConfigureProxyApiInstance(sdkConfig)

	log.Printf("Reading IVR config %s", d.Id())
	return withRetriesForRead(ctx, d, func() *resource.RetryError {
		ivrConfig, resp, getErr := architectIvrProxy.GetArchitectIvr(architectIvrProxy.Api, d.Id())
		if getErr != nil {
			if isStatus404(resp) {
				return resource.RetryableError(fmt.Errorf("Failed to read IVR config %s: %s", d.Id(), getErr))
			}
			return resource.NonRetryableError(fmt.Errorf("Failed to read IVR config %s: %s", d.Id(), getErr))
		}

		if ivrConfig.State != nil && *ivrConfig.State == "deleted" {
			d.SetId("")
			return nil
		}

		cc := consistency_checker.NewConsistencyCheck(ctx, d, meta, resourceArchitectIvrConfig())
		d.Set("name", *ivrConfig.Name)
		d.Set("dnis", stringListToSetOrNil(ivrConfig.Dnis))

		if ivrConfig.Description != nil {
			d.Set("description", *ivrConfig.Description)
		} else {
			d.Set("description", nil)
		}

		if ivrConfig.OpenHoursFlow != nil {
			d.Set("open_hours_flow_id", *ivrConfig.OpenHoursFlow.Id)
		} else {
			d.Set("open_hours_flow_id", nil)
		}

		if ivrConfig.ClosedHoursFlow != nil {
			d.Set("closed_hours_flow_id", *ivrConfig.ClosedHoursFlow.Id)
		} else {
			d.Set("closed_hours_flow_id", nil)
		}

		if ivrConfig.HolidayHoursFlow != nil {
			d.Set("holiday_hours_flow_id", *ivrConfig.HolidayHoursFlow.Id)
		} else {
			d.Set("holiday_hours_flow_id", nil)
		}

		if ivrConfig.ScheduleGroup != nil {
			d.Set("schedule_group_id", *ivrConfig.ScheduleGroup.Id)
		} else {
			d.Set("schedule_group_id", nil)
		}

		if ivrConfig.Division != nil && ivrConfig.Division.Id != nil {
			d.Set("division_id", *ivrConfig.Division.Id)
		} else {
			d.Set("division_id", nil)
		}

		log.Printf("Read IVR config %s %s", d.Id(), *ivrConfig.Name)
		return cc.CheckState()
	})
}

func updateIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)
	description := d.Get("description").(string)
	openHoursFlowId := buildSdkDomainEntityRef(d, "open_hours_flow_id")
	closedHoursFlowId := buildSdkDomainEntityRef(d, "closed_hours_flow_id")
	holidayHoursFlowId := buildSdkDomainEntityRef(d, "holiday_hours_flow_id")
	scheduleGroupId := buildSdkDomainEntityRef(d, "schedule_group_id")
	divisionId := d.Get("division_id").(string)
	dnis := buildSdkStringList(d, "dnis")

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectIvrProxy.ConfigureProxyApiInstance(sdkConfig)

	diagErr := retryWhen(isVersionMismatch, func() (*platformclientv2.APIResponse, diag.Diagnostics) {
		// Get current version
		ivr, resp, getErr := architectIvrProxy.GetArchitectIvr(architectIvrProxy.Api, d.Id())
		if getErr != nil {
			return resp, diag.Errorf("Failed to read IVR config %s: %s", d.Id(), getErr)
		}

		ivrBody := platformclientv2.Ivr{
			Version:          ivr.Version,
			Name:             &name,
			OpenHoursFlow:    openHoursFlowId,
			ClosedHoursFlow:  closedHoursFlowId,
			HolidayHoursFlow: holidayHoursFlowId,
			ScheduleGroup:    scheduleGroupId,
		}

		if description != "" {
			ivrBody.Description = &description
		}

		if divisionId != "" {
			ivrBody.Division = &platformclientv2.Writabledivision{Id: &divisionId}
		}

		var dnisChunks [][]string
		if dnis != nil {
			dnisChunks = chunkSlice(*dnis, maxDnisPerRequest)
			if len(dnisChunks) == 1 {
				ivrBody.Dnis = &dnisChunks[0]
			} else {
				ivrBody.Dnis = nil
			}
		}

		// It might need to wait for a dependent did_pool to be created to avoid an eventual consistency issue which
		// would result in the error "Field 'didPoolId' is required and cannot be empty."
		if ivrBody.Dnis != nil {
			time.Sleep(3 * time.Second)
		}
		log.Printf("Updating IVR config %s", name)
		_, resp, putErr := architectIvrProxy.PutArchitectIvr(architectIvrProxy.Api, d.Id(), &ivrBody)

		if putErr != nil {
			return resp, diag.Errorf("Failed to update IVR config %s: %s", d.Id(), putErr)
		}

		if len(dnisChunks) > 1 {
			_, resp, err := architectIvrProxy.UploadIvrDnisChunks(architectIvrProxy, dnisChunks, d.Id())
			if err != nil {
				return resp, diag.Errorf("%v", err)
			}
		}

		return resp, nil
	})
	if diagErr != nil {
		return diagErr
	}

	log.Printf("Updated IVR config %s", d.Id())
	return readIvrConfig(ctx, d, meta)
}

func deleteIvrConfig(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	name := d.Get("name").(string)

	sdkConfig := meta.(*ProviderMeta).ClientConfig
	architectIvrProxy.ConfigureProxyApiInstance(sdkConfig)

	log.Printf("Deleting IVR config %s", name)
	if _, err := architectIvrProxy.DeleteArchitectIvr(architectIvrProxy.Api, d.Id()); err != nil {
		return diag.Errorf("Failed to delete IVR config %s: %s", name, err)
	}

	return withRetries(ctx, 30*time.Second, func() *resource.RetryError {
		ivr, resp, err := architectIvrProxy.GetArchitectIvr(architectIvrProxy.Api, d.Id())
		if err != nil {
			if isStatus404(resp) {
				// IVR config deleted
				log.Printf("Deleted IVR config %s", d.Id())
				return nil
			}
			return resource.NonRetryableError(fmt.Errorf("Error deleting IVR config %s: %s", d.Id(), err))
		}

		if ivr.State != nil && *ivr.State == "deleted" {
			// IVR config deleted
			log.Printf("Deleted IVR config %s", d.Id())
			return nil
		}

		return resource.RetryableError(fmt.Errorf("IVR config %s still exists", d.Id()))
	})
}
