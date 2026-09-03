// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	environmentsettings "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/environment_settings"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &EnvironmentResource{}
	_ resource.ResourceWithConfigure        = &EnvironmentResource{}
	_ resource.ResourceWithImportState      = &EnvironmentResource{}
	_ resource.ResourceWithConfigValidators = &EnvironmentResource{}
	_ resource.ResourceWithModifyPlan       = &EnvironmentResource{}
)

// NewEnvironmentResource is a helper function to simplify the provider implementation.
func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

// EnvironmentResource is the resource implementation.
type EnvironmentResource struct {
	client *client.Client
}

// EnvironmentResourceModel describes the resource data model.
type EnvironmentResourceModel struct {
	ID                         types.String                    `tfsdk:"id"`
	Name                       types.String                    `tfsdk:"name"`
	ApplicationFamily          types.String                    `tfsdk:"application_family"`
	Type                       types.String                    `tfsdk:"type"`
	CountryCode                types.String                    `tfsdk:"country_code"`
	RingName                   types.String                    `tfsdk:"ring_name"`
	ApplicationVersion         types.String                    `tfsdk:"application_version"`
	IgnoreUpdateWindow         types.Bool                      `tfsdk:"ignore_update_window"`
	AzureRegion                types.String                    `tfsdk:"azure_region"`
	Status                     types.String                    `tfsdk:"status"`
	WebClientLoginURL          types.String                    `tfsdk:"web_client_login_url"`
	WebServiceURL              types.String                    `tfsdk:"web_service_url"`
	AppInsightsKey             types.String                    `tfsdk:"app_insights_key"`
	PlatformVersion            types.String                    `tfsdk:"platform_version"`
	AADTenantID                types.String                    `tfsdk:"aad_tenant_id"`
	PendingUpgradeVersion      types.String                    `tfsdk:"pending_upgrade_version"`
	PendingUpgradeScheduledFor types.String                    `tfsdk:"pending_upgrade_scheduled_for"`
	Settings                   *EnvironmentSettingsNestedModel `tfsdk:"settings"`
	Timeouts                   types.Object                    `tfsdk:"timeouts"`
}

// EnvironmentSettingsNestedModel describes the optional settings nested block within the environment resource.
type EnvironmentSettingsNestedModel struct {
	UpdateWindowStartTime   types.String `tfsdk:"update_window_start_time"`
	UpdateWindowEndTime     types.String `tfsdk:"update_window_end_time"`
	UpdateWindowTimeZone    types.String `tfsdk:"update_window_timezone"`
	AppInsightsKey          types.String `tfsdk:"app_insights_key"`
	SecurityGroupID         types.String `tfsdk:"security_group_id"`
	AccessWithM365Licenses  types.Bool   `tfsdk:"access_with_m365_licenses"`
	AppUpdateCadence        types.String `tfsdk:"app_update_cadence"`
	PartnerAccessStatus     types.String `tfsdk:"partner_access_status"`
	AllowedPartnerTenantIDs types.List   `tfsdk:"allowed_partner_tenant_ids"`
}

// settingsTimeFormatRegex validates time in HH:mm format.
var settingsTimeFormatRegex = regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)

// Metadata returns the resource type name.
func (r *EnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

// Schema defines the schema for the resource.
func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Business Central environment in the Admin Center.\n\n" +
			"This resource creates and manages Business Central environments (Production or Sandbox). " +
			"Environment creation is an asynchronous operation that can take several minutes to complete.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ARM-like resource ID (format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName})",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the environment. Must be between 1 and 30 characters. Changing this forces a new Business Central Environment to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 30),
				},
			},
			"application_family": schema.StringAttribute{
				MarkdownDescription: "The application family for the environment. Defaults to 'BusinessCentral'. Changing this forces a new Business Central Environment to be created.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("BusinessCentral"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of environment. Must be either 'Production' or 'Sandbox'. Changing this forces a new Business Central Environment to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("Production", "Sandbox"),
				},
			},
			"country_code": schema.StringAttribute{
				MarkdownDescription: "The country/region code for the environment (e.g., 'US', 'GB', 'DK'). Changing this forces a new Business Central Environment to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ring_name": schema.StringAttribute{
				MarkdownDescription: "The release ring for the environment. Must be one of 'PROD', 'PREVIEW', or 'FAST'. Defaults to 'PROD'. Changing this forces a new Business Central Environment to be created.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("PROD"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("PROD", "PREVIEW", "FAST"),
				},
			},
			"application_version": schema.StringAttribute{
				MarkdownDescription: "The desired application version for the environment (e.g. `\"26.1\"`). " +
					"When set at creation, the version is passed to the Create API. " +
					"When changed after creation, the provider schedules an in-place upgrade via the Admin Center Updates API. " +
					"When not set, the API assigns the version based on the ring. " +
					"During a scheduled or running upgrade, this attribute reflects the target version and does not cause drift. " +
					"If the upgrade fails, this attribute reflects the currently running version, causing drift and triggering a retry on next apply. " +
					"Do not use this alongside `bcadmincenter_environment_update_schedule` for the same environment.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					utils.NoDowngradeVersion(),
				},
			},
			"ignore_update_window": schema.BoolAttribute{
				MarkdownDescription: "When `true`, the version upgrade scheduled via `application_version` may start immediately " +
					"without waiting for the environment's configured update window. " +
					"When `false` (default), the upgrade waits for the next update window. " +
					"This setting applies only to platform/environment version updates — it has no effect on app installations or updates.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"azure_region": schema.StringAttribute{
				MarkdownDescription: "The Azure region where the environment should be created. If not specified, a default region will be used. Changing this forces a new Business Central Environment to be created.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the environment (e.g., 'Active', 'Creating').",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"web_client_login_url": schema.StringAttribute{
				MarkdownDescription: "The URL for accessing the web client.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"web_service_url": schema.StringAttribute{
				MarkdownDescription: "The URL for web service access.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_insights_key": schema.StringAttribute{
				MarkdownDescription: "The Application Insights instrumentation key for the environment.",
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"platform_version": schema.StringAttribute{
				MarkdownDescription: "The platform version of the environment.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aad_tenant_id": schema.StringAttribute{
				MarkdownDescription: "The Azure AD tenant ID for the environment. If not specified, the value is read from the API response. Changing this forces a new resource to be created.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					// The tenant selects which API the resource is read from and written
					// to. Without this, editing it made Terraform plan an in-place update
					// that talked to the old tenant while recording the new one.
					stringplanmodifier.RequiresReplace(),
				},
			},
			"pending_upgrade_version": schema.StringAttribute{
				MarkdownDescription: "The target version of a currently selected/scheduled or running upgrade. " +
					"Empty when no upgrade is in progress. " +
					"While non-empty, `application_version` is suppressed to this value so no drift is reported.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pending_upgrade_scheduled_for": schema.StringAttribute{
				MarkdownDescription: "The RFC3339 datetime at which the pending upgrade is scheduled to run. " +
					"Empty when the upgrade will run at the next update window or when no upgrade is pending.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"timeouts": schema.SingleNestedAttribute{
				MarkdownDescription: "Timeout configuration for the resource operations.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"create": schema.StringAttribute{
						MarkdownDescription: "Timeout for create operations. Defaults to 60 minutes.",
						Optional:            true,
					},
					"delete": schema.StringAttribute{
						MarkdownDescription: "Timeout for delete operations. Defaults to 60 minutes.",
						Optional:            true,
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"settings": schema.SingleNestedBlock{
				MarkdownDescription: "Optional environment settings block. When specified, the settings are applied to the environment after creation and managed inline.",
				Attributes: map[string]schema.Attribute{
					"update_window_start_time": schema.StringAttribute{
						MarkdownDescription: "Start time for the update window in HH:mm format (24-hour). Requires `update_window_timezone` to be set.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								settingsTimeFormatRegex,
								"must be in HH:mm format (e.g., '22:00')",
							),
						},
					},
					"update_window_end_time": schema.StringAttribute{
						MarkdownDescription: "End time for the update window in HH:mm format (24-hour). Requires `update_window_timezone` to be set. Must be at least 6 hours after start time.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.RegexMatches(
								settingsTimeFormatRegex,
								"must be in HH:mm format (e.g., '06:00')",
							),
						},
					},
					"update_window_timezone": schema.StringAttribute{
						MarkdownDescription: "Windows time zone identifier for the update window (e.g., 'Pacific Standard Time', 'Eastern Standard Time'). Required if `update_window_start_time` or `update_window_end_time` are set.",
						Optional:            true,
					},
					"app_insights_key": schema.StringAttribute{
						MarkdownDescription: "Application Insights connection string or instrumentation key for environment telemetry. Warning: Setting this triggers an automatic environment restart.",
						Optional:            true,
						Sensitive:           true,
					},
					"security_group_id": schema.StringAttribute{
						MarkdownDescription: "Microsoft Entra (Azure AD) security group object ID to restrict environment access.",
						Optional:            true,
					},
					"access_with_m365_licenses": schema.BoolAttribute{
						MarkdownDescription: "Whether users can access the environment with Microsoft 365 licenses (requires environment version 21.1+).",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"app_update_cadence": schema.StringAttribute{
						MarkdownDescription: "How frequently AppSource apps should be updated. Valid values: `Default`, `DuringMajorUpgrade`, `DuringMajorMinorUpgrade`.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("Default", "DuringMajorUpgrade", "DuringMajorMinorUpgrade"),
						},
					},
					"partner_access_status": schema.StringAttribute{
						MarkdownDescription: "Partner access configuration. Valid values: `Disabled`, `AllowAllPartnerTenants`, `AllowSelectedPartnerTenants`. Note: Only internal global administrators can modify this setting.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("Disabled", "AllowAllPartnerTenants", "AllowSelectedPartnerTenants"),
						},
					},
					"allowed_partner_tenant_ids": schema.ListAttribute{
						MarkdownDescription: "List of partner tenant IDs allowed to access the environment. Only used when `partner_access_status` is `AllowSelectedPartnerTenants`.",
						Optional:            true,
						ElementType:         types.StringType,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *EnvironmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating BC Admin Center environment", map[string]interface{}{
		"name":               plan.Name.ValueString(),
		"application_family": plan.ApplicationFamily.ValueString(),
		"type":               plan.Type.ValueString(),
	})

	// Create environment service, targeting the specified tenant if aad_tenant_id is set.
	tenantID := r.client.GetTenantID()
	if !plan.AADTenantID.IsNull() && !plan.AADTenantID.IsUnknown() {
		tenantID = plan.AADTenantID.ValueString()
	}
	svc := NewService(r.client.ForTenant(tenantID))

	// Prepare create request.
	createReq := &CreateEnvironmentRequest{
		EnvironmentType: plan.Type.ValueString(),
		Name:            plan.Name.ValueString(),
		CountryCode:     plan.CountryCode.ValueString(),
		RingName:        plan.RingName.ValueString(), // API expects "PROD", "PREVIEW", "FAST"
		AzureRegion:     plan.AzureRegion.ValueString(),
	}

	// Include ApplicationVersion only when explicitly set by the user.
	// Save it now so we can restore the short form after the API returns the full form.
	configuredApplicationVersion := plan.ApplicationVersion
	if !plan.ApplicationVersion.IsNull() && !plan.ApplicationVersion.IsUnknown() && plan.ApplicationVersion.ValueString() != "" {
		createReq.ApplicationVersion = plan.ApplicationVersion.ValueString()
	}

	// Create the environment.
	operation, err := svc.Create(ctx, plan.ApplicationFamily.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating environment",
			fmt.Sprintf("Could not create environment: %s", err),
		)
		return
	}

	// Log the operation response for debugging.
	tflog.Debug(ctx, "Create operation response", map[string]interface{}{
		"operation_id":       operation.ID,
		"operation_type":     operation.Type,
		"product_family":     operation.ProductFamily,
		"application_family": operation.ApplicationFamily,
		"environment_name":   operation.EnvironmentName,
		"destination_env":    operation.DestinationEnvironment,
		"source_env":         operation.SourceEnvironment,
	})

	timeout := utils.OperationTimeout(ctx, plan.Timeouts, "create")

	// Always use the application_family from the plan when constructing API paths.
	// The operation response fields (productFamily, applicationFamily) are internal API
	// concepts that differ from the applicationFamily URL parameter (e.g. the API may
	// return productFamily="Financials" while the correct path segment is "BusinessCentral").
	appFamily := plan.ApplicationFamily.ValueString()

	envName := operation.EnvironmentName
	if envName == "" {
		envName = operation.DestinationEnvironment
	}
	if envName == "" {
		envName = plan.Name.ValueString()
	}

	tflog.Debug(ctx, "Waiting for environment creation to complete", map[string]interface{}{
		"operation_id":       operation.ID,
		"timeout":            timeout.String(),
		"application_family": appFamily,
		"environment_name":   envName,
	})

	// The PUT has been accepted, so the environment now exists in the tenant. Every
	// failure from here on records what is known before reporting the error, so
	// Terraform does not forget a live environment.
	if err := svc.WaitForOperation(ctx, appFamily, envName, operation.ID, timeout); err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for environment creation",
			fmt.Sprintf("Environment creation failed: %s. The environment has been recorded in state; "+
				"run `terraform plan` to reconcile it once provisioning settles.", err),
		)
		r.savePartialCreateState(ctx, resp, &plan, tenantID, appFamily, envName)
		return
	}

	// Log what we're about to use for the Get call.
	tflog.Debug(ctx, "Reading created environment", map[string]interface{}{
		"application_family": appFamily,
		"environment_name":   envName,
	})

	// Wait for the environment to become Active.
	// The operation succeeds when the create request is accepted, but the environment.
	// may still be in "Preparing" status. We need to poll until it's "Active".
	tflog.Debug(ctx, "Waiting for environment to become Active", map[string]interface{}{
		"application_family": appFamily,
		"environment_name":   envName,
	})

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	envTimeout, envCancel := context.WithTimeout(ctx, timeout)
	defer envCancel()

	for {
		env, err := svc.Get(ctx, appFamily, envName)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading created environment",
				fmt.Sprintf("Could not read environment after creation: %s. The environment has been "+
					"recorded in state; run `terraform plan` to reconcile it.", err),
			)
			r.savePartialCreateState(ctx, resp, &plan, tenantID, appFamily, envName)
			return
		}

		tflog.Debug(ctx, "Environment status check", map[string]interface{}{
			"status": env.Status,
		})

		if utils.StatusIs(env.Status, EnvironmentStatusActive) {
			// Environment is ready, update state and return.
			r.updateModelFromEnvironment(&plan, env)
			// Preserve the user-configured short version (e.g. "27.1") if the API
			// returned the full build version (e.g. "27.1.41698.41831").
			plan.ApplicationVersion = types.StringValue(
				normalizeApplicationVersion(configuredApplicationVersion.ValueString(), plan.ApplicationVersion.ValueString()))

			// Apply inline settings block if configured.
			if plan.Settings != nil {
				settingsSvc := environmentsettings.NewService(r.client.ForTenant(tenantID))
				if err := r.applyEnvironmentSettings(ctx, settingsSvc, plan.ApplicationFamily.ValueString(), envName, plan.Settings); err != nil {
					resp.Diagnostics.AddError(
						"Error applying environment settings",
						"Could not apply settings after environment creation: "+err.Error()+
							". The environment exists and has been recorded in state; re-run to apply the settings.",
					)
					r.savePartialCreateState(ctx, resp, &plan, tenantID, appFamily, envName)
					return
				}
				// Read back readable settings (update_window, security_group, m365 access).
				if err := r.readEnvironmentSettings(ctx, settingsSvc, plan.ApplicationFamily.ValueString(), envName, plan.Settings); err != nil {
					resp.Diagnostics.AddError(
						"Error reading environment settings",
						"Could not read settings after applying: "+err.Error()+
							". The environment exists and has been recorded in state.",
					)
					r.savePartialCreateState(ctx, resp, &plan, tenantID, appFamily, envName)
					return
				}
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}

		// Check for failed states.
		if utils.StatusIs(env.Status, "Failed", "Suspended") {
			resp.Diagnostics.AddError(
				"Environment creation failed",
				fmt.Sprintf("Environment entered %s state during creation. It has been recorded in "+
					"state so it can be inspected or destroyed with Terraform.", env.Status),
			)
			r.updateModelFromEnvironment(&plan, env)
			r.savePartialCreateState(ctx, resp, &plan, tenantID, appFamily, envName)
			return
		}

		// Wait for next tick or timeout.
		select {
		case <-envTimeout.Done():
			resp.Diagnostics.AddError(
				"Timeout waiting for environment",
				fmt.Sprintf("Environment did not become Active within %v (current status: %s). It has "+
					"been recorded in state; run `terraform plan` once provisioning completes.", timeout, env.Status),
			)
			r.updateModelFromEnvironment(&plan, env)
			r.savePartialCreateState(ctx, resp, &plan, tenantID, appFamily, envName)
			return
		case <-ticker.C:
			// Continue polling.
			continue
		}
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentResourceModel

	// Read current state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save the prior version before it may be overwritten by updateModelFromEnvironment.
	// This allows us to preserve a user-configured short form (e.g. "27.1") when the
	// API returns the full build version (e.g. "27.1.41698.41831").
	priorApplicationVersion := state.ApplicationVersion

	tflog.Debug(ctx, "Reading BC Admin Center environment", map[string]interface{}{
		"name":               state.Name.ValueString(),
		"application_family": state.ApplicationFamily.ValueString(),
	})

	// Create environment service, targeting the tenant from state.
	svc := NewService(r.client.ForTenant(state.AADTenantID.ValueString()))

	// Get the environment.
	env, err := svc.Get(ctx, state.ApplicationFamily.ValueString(), state.Name.ValueString())
	if err != nil {
		// Deleted out of band (for example in the Admin Center portal). Removing it from
		// state lets the next plan recreate it; raising an error instead made every
		// subsequent plan, apply and destroy fail until the user ran `terraform state rm`.
		if isEnvironmentNotFoundError(err) {
			tflog.Warn(ctx, "Environment no longer exists; removing from state", map[string]interface{}{
				"name":               state.Name.ValueString(),
				"application_family": state.ApplicationFamily.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading environment",
			fmt.Sprintf("Could not read environment: %s", err),
		)
		return
	}

	// Update state with current environment data.
	r.updateModelFromEnvironment(&state, env)

	// Drift detection: check pending/running/failed updates.
	updates, err := svc.GetUpdates(ctx, state.ApplicationFamily.ValueString(), state.Name.ValueString())
	if err != nil {
		// Non-fatal: if the updates endpoint fails, fall back to environment version.
		tflog.Warn(ctx, "Failed to get environment updates for drift detection; using environment version", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		r.applyUpdatesDriftDetection(&state, env, updates)
	}

	// Normalize application_version: preserve the prior short form if the API returned
	// the full build version starting with it (e.g. keep "27.1" when API says "27.1.41698.41831").
	if !priorApplicationVersion.IsNull() && !priorApplicationVersion.IsUnknown() && !state.ApplicationVersion.IsNull() {
		state.ApplicationVersion = types.StringValue(
			normalizeApplicationVersion(priorApplicationVersion.ValueString(), state.ApplicationVersion.ValueString()))
	}

	// Read inline settings if the settings block is configured in state.
	if state.Settings != nil {
		settingsSvc := environmentsettings.NewService(r.client.ForTenant(state.AADTenantID.ValueString()))
		if err := r.readEnvironmentSettings(ctx, settingsSvc, state.ApplicationFamily.ValueString(), state.Name.ValueString(), state.Settings); err != nil {
			resp.Diagnostics.AddError(
				"Error reading environment settings",
				"Could not read inline settings: "+err.Error(),
			)
			return
		}
	}

	// Set refreshed state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state EnvironmentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Only application_version and ignore_update_window support in-place updates.
	versionChanged := shouldScheduleVersionUpgrade(plan, state)
	windowChanged := !plan.IgnoreUpdateWindow.Equal(state.IgnoreUpdateWindow)
	settingsChanged := settingsBlockChanged(plan.Settings, state.Settings)

	if !versionChanged && !windowChanged && !settingsChanged {
		// Nothing to do; copy plan to state.
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Apply inline settings changes if the block was added, modified, or removed.
	if settingsChanged {
		// Removing the block means "revert these settings", not "leave them alone".
		// Guarding this whole branch on plan.Settings != nil made removal a silent no-op:
		// the apply reported success, state showed no settings, and the environment stayed
		// locked to its security group. An all-null model drives the existing clearing
		// paths (ClearSecurityGroup and friends) in applyEnvironmentSettingsChanges.
		planSettings := plan.Settings
		blockRemoved := planSettings == nil
		if blockRemoved {
			planSettings = clearedSettingsModel()
		}

		settingsSvc := environmentsettings.NewService(r.client.ForTenant(state.AADTenantID.ValueString()))
		if err := r.applyEnvironmentSettingsChanges(ctx, settingsSvc, state.ApplicationFamily.ValueString(), state.Name.ValueString(), planSettings, state.Settings); err != nil {
			resp.Diagnostics.AddError(
				"Error updating environment settings",
				"Could not update inline settings: "+err.Error(),
			)
			return
		}
		// Read back readable settings to keep state consistent. Skipped when the block was
		// removed: the config has no settings block, so writing one back into state would
		// fail the apply with an inconsistent result.
		if !blockRemoved {
			if err := r.readEnvironmentSettings(ctx, settingsSvc, state.ApplicationFamily.ValueString(), state.Name.ValueString(), planSettings); err != nil {
				resp.Diagnostics.AddError(
					"Error reading environment settings",
					"Could not read settings after update: "+err.Error(),
				)
				return
			}
		}
	}

	if !versionChanged {
		// Only settings and/or ignore_update_window changed; persist the plan (with refreshed settings).
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	if plan.ApplicationVersion.IsNull() || plan.ApplicationVersion.IsUnknown() || plan.ApplicationVersion.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Cannot update without application_version",
			"application_version must be set to schedule a version upgrade.",
		)
		return
	}

	svc := NewService(r.client.ForTenant(state.AADTenantID.ValueString()))

	targetVersion := plan.ApplicationVersion.ValueString()
	ignoreUpdateWindow := plan.IgnoreUpdateWindow.ValueBool()

	tflog.Debug(ctx, "Scheduling environment version upgrade", map[string]interface{}{
		"application_family":   state.ApplicationFamily.ValueString(),
		"environment_name":     state.Name.ValueString(),
		"target_version":       targetVersion,
		"ignore_update_window": ignoreUpdateWindow,
	})

	if err := svc.SelectUpdateVersion(ctx, state.ApplicationFamily.ValueString(), state.Name.ValueString(), targetVersion, ignoreUpdateWindow); err != nil {
		resp.Diagnostics.AddError(
			"Error scheduling environment upgrade",
			fmt.Sprintf("Could not schedule upgrade to version %s: %s", targetVersion, err),
		)
		// Any settings changes above were already applied remotely. Returning without
		// saving left state holding the *old* settings, so state and the environment
		// disagreed until the next successful apply. Persist everything except the version,
		// which is the part that failed.
		plan.ApplicationVersion = state.ApplicationVersion
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Store the target version in state immediately; drift detection in Read will resolve it.
	plan.ApplicationVersion = types.StringValue(targetVersion)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentResourceModel

	// Read current state.
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting BC Admin Center environment", map[string]interface{}{
		"name":               state.Name.ValueString(),
		"application_family": state.ApplicationFamily.ValueString(),
	})

	// Create environment service, targeting the tenant from state.
	svc := NewService(r.client.ForTenant(state.AADTenantID.ValueString()))

	// Delete the environment.
	operation, err := svc.Delete(ctx, state.ApplicationFamily.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting environment",
			fmt.Sprintf("Could not delete environment: %s", err),
		)
		return
	}

	// If operation is nil, the environment was already deleted.
	if operation == nil {
		return
	}

	timeout := utils.OperationTimeout(ctx, state.Timeouts, "delete")

	// Always use the application_family from state when constructing API paths, matching
	// Create. The operation response's productFamily/applicationFamily are internal API
	// concepts that differ from the applicationFamily URL segment (the API may return
	// productFamily="Financials" where the path segment must be "BusinessCentral").
	// Preferring them built a path that 404s with a code other than EnvironmentNotFound,
	// so the delete reported failure and Terraform kept the resource in state even though
	// the deletion was running normally.
	appFamily := state.ApplicationFamily.ValueString()

	envName := operation.EnvironmentName
	if envName == "" {
		envName = operation.SourceEnvironment
	}
	if envName == "" {
		envName = state.Name.ValueString()
	}

	tflog.Debug(ctx, "Waiting for environment deletion to complete", map[string]interface{}{
		"operation_id":       operation.ID,
		"timeout":            timeout.String(),
		"application_family": appFamily,
		"environment_name":   envName,
	})

	if err := svc.WaitForOperation(ctx, appFamily, envName, operation.ID, timeout); err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for environment deletion",
			fmt.Sprintf("Environment deletion failed: %s", err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform state.
func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the ARM-like ID.
	tenantID, applicationFamily, environmentName, err := ParseEnvironmentID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected ARM-like resource ID in format '/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}', got: %s\nError: %s",
				req.ID, err.Error()),
		)
		return
	}

	// Set the attributes.
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_family"), applicationFamily)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), environmentName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aad_tenant_id"), tenantID)...)
}

// ConfigValidators enforces that an update window is configured completely or not at all.
//
// The three attributes are sent as one payload whose fields are `*string` with
// `omitempty`, so a nil field is simply absent from the request and the API keeps its
// previous value. Clearing one of three therefore looked like it worked and silently did
// not. Requiring them together removes the partial state that cannot be expressed.
func (r *EnvironmentResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.RequiredTogether(
			path.MatchRoot("settings").AtName("update_window_start_time"),
			path.MatchRoot("settings").AtName("update_window_end_time"),
			path.MatchRoot("settings").AtName("update_window_timezone"),
		),
	}
}

// ModifyPlan refuses, at plan time, to "remove" a setting the Admin Center offers no way
// to unset.
//
// applyEnvironmentSettingsChanges can only send values the API accepts. For these
// settings there is no clear operation, so a removal previously produced no request at
// all: Terraform reported success, recorded the attribute as gone, and left the
// environment unchanged. None of them are read back from the API either — they are
// write-only or need elevated permissions — so no later plan could detect the difference.
// The result was permanent, invisible drift.
//
// Erroring here keeps the failure in `terraform plan`, before anything is applied, and
// names the value to set instead.
func (r *EnvironmentResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Nothing to compare on create (no prior state) or destroy (no plan).
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state EnvironmentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(checkUnclearableSettings(&plan, &state)...)
}

// checkUnclearableSettings reports removals the Admin Center cannot carry out. Kept
// separate from ModifyPlan so the rules can be tested directly on models.
func checkUnclearableSettings(plan, state *EnvironmentResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics
	if state.Settings == nil {
		return diags
	}

	// Removing the whole block is the same request as clearing every attribute in it.
	planned := plan.Settings
	if planned == nil {
		planned = clearedSettingsModel()
	}

	settingsPath := path.Root("settings")

	for _, c := range []struct {
		attribute string
		wasSet    bool
		nowUnset  bool
		remedy    string
	}{
		{
			attribute: "app_update_cadence",
			wasSet:    !state.Settings.AppUpdateCadence.IsNull(),
			nowUnset:  planned.AppUpdateCadence.IsNull(),
			remedy:    `Set it explicitly to "Default" to return to the standard cadence.`,
		},
		{
			attribute: "partner_access_status",
			wasSet:    !state.Settings.PartnerAccessStatus.IsNull(),
			nowUnset:  planned.PartnerAccessStatus.IsNull(),
			remedy:    `Set it explicitly to "Disabled" to withdraw partner access.`,
		},
	} {
		if c.wasSet && c.nowUnset {
			diags.AddAttributeError(
				settingsPath.AtName(c.attribute),
				fmt.Sprintf("Cannot remove %s", c.attribute),
				fmt.Sprintf("The Admin Center API has no operation to unset %s, so removing it from the "+
					"configuration would leave the environment unchanged while Terraform recorded it as "+
					"removed — and this setting is not read back, so the difference would never surface "+
					"in a later plan.\n\n%s", c.attribute, c.remedy),
			)
		}
	}

	windowWasSet := !state.Settings.UpdateWindowStartTime.IsNull() ||
		!state.Settings.UpdateWindowEndTime.IsNull() ||
		!state.Settings.UpdateWindowTimeZone.IsNull()
	windowNowUnset := planned.UpdateWindowStartTime.IsNull() &&
		planned.UpdateWindowEndTime.IsNull() &&
		planned.UpdateWindowTimeZone.IsNull()

	if windowWasSet && windowNowUnset {
		diags.AddAttributeError(
			settingsPath.AtName("update_window_start_time"),
			"Cannot remove the update window",
			"The Admin Center API has no operation to clear a configured update window: the request "+
				"omits absent fields, so the previous window would stay in force while Terraform "+
				"recorded it as removed.\n\nSet the window you want instead, or keep the current "+
				"values to leave it as it is.",
		)
	}

	return diags
}

// clearedSettingsModel returns a settings model with every attribute null, representing
// the user having removed the `settings` block. Passing it to
// applyEnvironmentSettingsChanges reverts each setting through the same code paths that
// handle an individual attribute being cleared.
func clearedSettingsModel() *EnvironmentSettingsNestedModel {
	return &EnvironmentSettingsNestedModel{
		UpdateWindowStartTime:   types.StringNull(),
		UpdateWindowEndTime:     types.StringNull(),
		UpdateWindowTimeZone:    types.StringNull(),
		AppInsightsKey:          types.StringNull(),
		SecurityGroupID:         types.StringNull(),
		AccessWithM365Licenses:  types.BoolNull(),
		AppUpdateCadence:        types.StringNull(),
		PartnerAccessStatus:     types.StringNull(),
		AllowedPartnerTenantIDs: types.ListNull(types.StringType),
	}
}

// resolveUnknownComputed replaces any attribute still carrying an unknown value with
// null.
//
// Terraform rejects a state object that contains unknown values after apply, so a model
// can only be written to state once every unknown has been resolved. Create fills these
// from the API on the happy path; this is the fallback for saving partial state when a
// step after the environment already exists fails.
func resolveUnknownComputed(model *EnvironmentResourceModel) {
	for _, attr := range []*types.String{
		&model.ID,
		&model.ApplicationFamily,
		&model.RingName,
		&model.ApplicationVersion,
		&model.AzureRegion,
		&model.Status,
		&model.WebClientLoginURL,
		&model.WebServiceURL,
		&model.AppInsightsKey,
		&model.PlatformVersion,
		&model.AADTenantID,
		&model.PendingUpgradeVersion,
		&model.PendingUpgradeScheduledFor,
	} {
		if attr.IsUnknown() {
			*attr = types.StringNull()
		}
	}
	if model.IgnoreUpdateWindow.IsUnknown() {
		model.IgnoreUpdateWindow = types.BoolNull()
	}
	if model.Settings != nil && model.Settings.AccessWithM365Licenses.IsUnknown() {
		model.Settings.AccessWithM365Licenses = types.BoolNull()
	}
}

// savePartialCreateState records an environment that exists remotely but whose creation
// could not be completed or read back.
//
// Once the PUT succeeds the environment is being provisioned in the tenant. Returning an
// error without writing state made Terraform forget it entirely: the environment
// finished provisioning, the next apply re-issued the PUT and failed with "already
// exists", and the only way out was to delete it by hand in the Admin Center. Recording
// what is known lets the next plan reconcile it instead.
func (r *EnvironmentResource) savePartialCreateState(ctx context.Context, resp *resource.CreateResponse, plan *EnvironmentResourceModel, tenantID, appFamily, envName string) {
	if plan.ID.IsNull() || plan.ID.IsUnknown() || plan.ID.ValueString() == "" {
		plan.ID = types.StringValue(BuildEnvironmentID(tenantID, appFamily, envName))
	}
	if plan.Name.IsNull() || plan.Name.IsUnknown() {
		plan.Name = types.StringValue(envName)
	}
	resolveUnknownComputed(plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// updateModelFromEnvironment updates the Terraform model with data from the API.
// It sets application_version as a baseline from the environment GET response.
// applyUpdatesDriftDetection may override application_version based on pending updates.
func (r *EnvironmentResource) updateModelFromEnvironment(model *EnvironmentResourceModel, env *Environment) {
	// Build ARM-like ID using tenant ID from aad_tenant_id field.
	tenantID := env.AADTenantID
	if tenantID == "" {
		// Fallback to provider tenant if not available in response.
		tenantID = r.client.GetTenantID()
	}

	model.ID = types.StringValue(BuildEnvironmentID(tenantID, env.ApplicationFamily, env.Name))
	model.Name = types.StringValue(env.Name)
	model.ApplicationFamily = types.StringValue(env.ApplicationFamily)
	model.Type = types.StringValue(env.Type)
	model.CountryCode = types.StringValue(env.CountryCode)
	model.Status = types.StringValue(env.Status)
	model.WebClientLoginURL = types.StringValue(env.WebClientLoginURL)
	model.AADTenantID = types.StringValue(env.AADTenantID)

	if env.WebServiceURL != "" {
		model.WebServiceURL = types.StringValue(env.WebServiceURL)
	} else {
		model.WebServiceURL = types.StringNull()
	}

	if env.AppInsightsKey != "" {
		model.AppInsightsKey = types.StringValue(env.AppInsightsKey)
	} else {
		model.AppInsightsKey = types.StringNull()
	}

	// azure_region is deliberately left untouched. The API accepts it on create but never
	// returns it, so there is nothing to refresh from, and overwriting the model with null
	// discarded the configured value. Because the attribute is Optional + Computed +
	// RequiresReplace, that made every apply fail with "Provider produced inconsistent
	// result after apply" and every later plan propose replacing the environment.

	// Normalize ring name from API response format to Terraform format.
	// API accepts "PROD", "PREVIEW", "FAST" on input but returns "Production", "Preview", "Fast" on output.
	if env.RingName != "" {
		normalizedRing := normalizeRingName(env.RingName)
		model.RingName = types.StringValue(normalizedRing)
	} else {
		model.RingName = types.StringNull()
	}

	// Set application_version from environment response as baseline.
	// applyUpdatesDriftDetection may override this with the target version.
	if env.ApplicationVersion != "" {
		model.ApplicationVersion = types.StringValue(env.ApplicationVersion)
	} else {
		model.ApplicationVersion = types.StringNull()
	}

	if env.PlatformVersion != "" {
		model.PlatformVersion = types.StringValue(env.PlatformVersion)
	} else {
		model.PlatformVersion = types.StringNull()
	}

	// Clear pending upgrade attrs; applyUpdatesDriftDetection will populate them if an upgrade is in flight.
	model.PendingUpgradeVersion = types.StringValue("")
	model.PendingUpgradeScheduledFor = types.StringValue("")
}

// applyUpdatesDriftDetection applies drift detection logic based on the environment updates list.
//
// Suppression table:
//
//	| selected | updateStatus          | behavior                                              |
//	|----------|-----------------------|-------------------------------------------------------|
//	| true     | "" / "scheduled" / "running" | Suppress drift; set application_version = targetVersion; populate pending_ attrs |
//	| true     | "failed"              | Report drift; clear pending_ attrs                    |
//	| true     | "succeeded" / other   | No suppression; clear pending_ attrs                  |
//	| false    | any                   | No suppression; clear pending_ attrs                  |
//
// The API may return selected:true with an empty updateStatus immediately after scheduling
// (before the upgrade transitions to "scheduled"). We treat that the same as "scheduled" to
// avoid false-positive drift during the window between PATCH and status propagation.
func (r *EnvironmentResource) applyUpdatesDriftDetection(model *EnvironmentResourceModel, env *Environment, updates []EnvironmentUpdate) {
	// Find the selected update.
	var selectedUpdate *EnvironmentUpdate
	for i := range updates {
		if updates[i].Selected {
			selectedUpdate = &updates[i]
			break
		}
	}

	if selectedUpdate == nil {
		// No selected update: use applicationVersion from environment GET (no drift if versions match).
		// pending_ attrs are already cleared by updateModelFromEnvironment.
		return
	}

	switch selectedUpdate.UpdateStatus {
	case UpdateStatusFailed:
		// Drift: store the currently running version so Terraform detects a change and retries.
		// pending_ attrs remain empty (upgrade is not in progress).
		if env.ApplicationVersion != "" {
			model.ApplicationVersion = types.StringValue(env.ApplicationVersion)
		} else {
			model.ApplicationVersion = types.StringNull()
		}
	case UpdateStatusScheduled, UpdateStatusRunning, "":
		// Suppress drift: the upgrade is selected, in-progress, or just scheduled (status not yet
		// propagated). Store the target version and surface the pending upgrade attributes.
		model.ApplicationVersion = types.StringValue(selectedUpdate.TargetVersion)
		model.PendingUpgradeVersion = types.StringValue(selectedUpdate.TargetVersion)
		if selectedUpdate.ScheduleDetails != nil && selectedUpdate.ScheduleDetails.SelectedDateTime != "" {
			model.PendingUpgradeScheduledFor = types.StringValue(selectedUpdate.ScheduleDetails.SelectedDateTime)
		} else {
			model.PendingUpgradeScheduledFor = types.StringValue("")
		}
	default:
		// For other statuses (e.g., "succeeded"), fall through to the environment version (already set).
		// pending_ attrs remain empty.
	}
}

// normalizeRingName converts API ring name format to Terraform format.
// API returns "Production", "Preview", "Fast" but Terraform expects "PROD", "PREVIEW", "FAST".
func normalizeRingName(apiRingName string) string {
	switch apiRingName {
	case "Production":
		return "PROD"
	case "Preview":
		return "PREVIEW"
	case "Fast":
		return "FAST"
	default:
		// Return as-is if unknown.
		return apiRingName
	}
}

// normalizeApplicationVersion returns priorVersion when the API-reported version should
// not cause drift in Terraform state. Two cases are handled:
//
//  1. Short-form preservation: user configured "27.1" and API returned the full build
//     version "27.1.41698.41831" — keep "27.1" to avoid spurious drift.
//
//  2. External auto-upgrade suppression: Microsoft upgraded the environment from "27.5"
//     to "28.0" without Terraform's involvement. In that case the API version is higher
//     than what the user configured, but no config change was made, so we preserve the
//     user-configured prior version in state. This prevents the plan from showing a
//     false "28.0 → 27.5" downgrade diff on the next run.
//
// A "." separator check is used in case 1 to avoid incorrectly matching "27.1" against
// "27.10.xxx".
func normalizeApplicationVersion(priorVersion, apiVersion string) string {
	if priorVersion == "" || apiVersion == "" {
		return apiVersion
	}
	// Case 1: full-build form of the same major.minor.
	if apiVersion == priorVersion || strings.HasPrefix(apiVersion, priorVersion+".") {
		return priorVersion
	}
	// Case 2: API version is higher at the major.minor level — external auto-upgrade.
	// Preserve the user's configured version to suppress spurious drift.
	if isAPIVersionHigher(apiVersion, priorVersion) {
		return priorVersion
	}
	return apiVersion
}

// isAPIVersionHigher returns true when apiVersion is strictly greater than priorVersion
// at the major.minor level. Unparseable versions return false (do not suppress).
func isAPIVersionHigher(apiVersion, priorVersion string) bool {
	apiParts := strings.SplitN(apiVersion, ".", 3)
	priorParts := strings.SplitN(priorVersion, ".", 3)
	if len(apiParts) < 2 || len(priorParts) < 2 {
		return false
	}
	apiMajor, err := strconv.Atoi(apiParts[0])
	if err != nil {
		return false
	}
	priorMajor, err := strconv.Atoi(priorParts[0])
	if err != nil {
		return false
	}
	if apiMajor != priorMajor {
		return apiMajor > priorMajor
	}
	// Same major — compare minor (take only the first numeric segment).
	apiMinor, err := strconv.Atoi(strings.SplitN(apiParts[1], ".", 2)[0])
	if err != nil {
		return false
	}
	priorMinor, err := strconv.Atoi(strings.SplitN(priorParts[1], ".", 2)[0])
	if err != nil {
		return false
	}
	return apiMinor > priorMinor
}

func shouldScheduleVersionUpgrade(plan, state EnvironmentResourceModel) bool {
	// An unknown plan value for application_version means the user did not set it; treat
	// it as no version change so that settings-only updates do not block on a missing version.
	return !plan.ApplicationVersion.IsUnknown() && !plan.ApplicationVersion.Equal(state.ApplicationVersion)
}

// settingsBlockChanged returns true if the settings block differs between plan and state.
// Both nil means no change. One nil and one non-nil means change (block added/removed).
// For Computed-only fields like access_with_m365_licenses, an unknown plan value means
// the provider will resolve it during apply — treat it as unchanged to avoid false positives.
func settingsBlockChanged(plan, state *EnvironmentSettingsNestedModel) bool {
	if plan == nil && state == nil {
		return false
	}
	if plan == nil || state == nil {
		return true
	}
	// For access_with_m365_licenses, skip the comparison when the plan value is unknown.
	// Unknown values occur when Terraform cannot determine the plan value from config — most
	// commonly on the first apply when the settings block is added (no prior state), but also
	// when importing or after state corruption. UseStateForUnknown() eliminates this for all
	// subsequent plans once the state has a value.
	m365Changed := !plan.AccessWithM365Licenses.IsUnknown() && !plan.AccessWithM365Licenses.Equal(state.AccessWithM365Licenses)
	return !plan.UpdateWindowStartTime.Equal(state.UpdateWindowStartTime) ||
		!plan.UpdateWindowEndTime.Equal(state.UpdateWindowEndTime) ||
		!plan.UpdateWindowTimeZone.Equal(state.UpdateWindowTimeZone) ||
		!plan.AppInsightsKey.Equal(state.AppInsightsKey) ||
		!plan.SecurityGroupID.Equal(state.SecurityGroupID) ||
		m365Changed ||
		!plan.AppUpdateCadence.Equal(state.AppUpdateCadence) ||
		!plan.PartnerAccessStatus.Equal(state.PartnerAccessStatus) ||
		!plan.AllowedPartnerTenantIDs.Equal(state.AllowedPartnerTenantIDs)
}

// applyEnvironmentSettings applies all settings from the nested block to the environment via the settings service.
func (r *EnvironmentResource) applyEnvironmentSettings(ctx context.Context, svc *environmentsettings.Service, applicationFamily, environmentName string, settings *EnvironmentSettingsNestedModel) error {
	// Apply update window if any component is set.
	if !settings.UpdateWindowStartTime.IsNull() || !settings.UpdateWindowEndTime.IsNull() || !settings.UpdateWindowTimeZone.IsNull() {
		us := &environmentsettings.UpdateSettings{}
		if !settings.UpdateWindowStartTime.IsNull() {
			v := settings.UpdateWindowStartTime.ValueString()
			us.PreferredStartTime = &v
		}
		if !settings.UpdateWindowEndTime.IsNull() {
			v := settings.UpdateWindowEndTime.ValueString()
			us.PreferredEndTime = &v
		}
		if !settings.UpdateWindowTimeZone.IsNull() {
			v := settings.UpdateWindowTimeZone.ValueString()
			us.TimeZoneID = &v
		}
		if _, err := svc.SetUpdateSettings(ctx, applicationFamily, environmentName, us); err != nil {
			return fmt.Errorf("setting update window: %w", err)
		}
	}

	// Apply Application Insights key if provided.
	if !settings.AppInsightsKey.IsNull() {
		if err := svc.SetAppInsightsKey(ctx, applicationFamily, environmentName, settings.AppInsightsKey.ValueString()); err != nil {
			return fmt.Errorf("setting app insights key: %w", err)
		}
	}

	// Apply security group if provided.
	if !settings.SecurityGroupID.IsNull() {
		if err := svc.SetSecurityGroup(ctx, applicationFamily, environmentName, settings.SecurityGroupID.ValueString()); err != nil {
			return fmt.Errorf("setting security group: %w", err)
		}
	}

	// Apply M365 license access if explicitly provided (not null and not unknown).
	if !settings.AccessWithM365Licenses.IsNull() && !settings.AccessWithM365Licenses.IsUnknown() {
		if err := svc.SetAccessWithM365Licenses(ctx, applicationFamily, environmentName, settings.AccessWithM365Licenses.ValueBool()); err != nil {
			return fmt.Errorf("setting M365 license access: %w", err)
		}
	}

	// Apply app update cadence if provided.
	if !settings.AppUpdateCadence.IsNull() {
		if err := svc.SetAppUpdateCadence(ctx, applicationFamily, environmentName, settings.AppUpdateCadence.ValueString()); err != nil {
			return fmt.Errorf("setting app update cadence: %w", err)
		}
	}

	// Apply partner access if provided.
	if !settings.PartnerAccessStatus.IsNull() {
		pa := &environmentsettings.PartnerAccessRequest{
			Status: settings.PartnerAccessStatus.ValueString(),
		}
		if settings.PartnerAccessStatus.ValueString() == "AllowSelectedPartnerTenants" && !settings.AllowedPartnerTenantIDs.IsNull() {
			var tenantIDs []string
			diags := settings.AllowedPartnerTenantIDs.ElementsAs(ctx, &tenantIDs, false)
			if diags.HasError() {
				return fmt.Errorf("reading allowed_partner_tenant_ids: %s", diags)
			}
			pa.AllowedPartnerTenantIDs = tenantIDs
		}
		if err := svc.SetPartnerAccess(ctx, applicationFamily, environmentName, pa); err != nil {
			return fmt.Errorf("setting partner access: %w", err)
		}
	}

	return nil
}

// applyEnvironmentSettingsChanges applies only the settings that changed between plan and state.
func (r *EnvironmentResource) applyEnvironmentSettingsChanges(ctx context.Context, svc *environmentsettings.Service, applicationFamily, environmentName string, plan, state *EnvironmentSettingsNestedModel) error {
	// For a nil state (block was just added), apply everything.
	if state == nil {
		return r.applyEnvironmentSettings(ctx, svc, applicationFamily, environmentName, plan)
	}

	// Update window settings if changed.
	if !plan.UpdateWindowStartTime.Equal(state.UpdateWindowStartTime) ||
		!plan.UpdateWindowEndTime.Equal(state.UpdateWindowEndTime) ||
		!plan.UpdateWindowTimeZone.Equal(state.UpdateWindowTimeZone) {
		us := &environmentsettings.UpdateSettings{}
		if !plan.UpdateWindowStartTime.IsNull() {
			v := plan.UpdateWindowStartTime.ValueString()
			us.PreferredStartTime = &v
		}
		if !plan.UpdateWindowEndTime.IsNull() {
			v := plan.UpdateWindowEndTime.ValueString()
			us.PreferredEndTime = &v
		}
		if !plan.UpdateWindowTimeZone.IsNull() {
			v := plan.UpdateWindowTimeZone.ValueString()
			us.TimeZoneID = &v
		}
		if _, err := svc.SetUpdateSettings(ctx, applicationFamily, environmentName, us); err != nil {
			return fmt.Errorf("updating update window: %w", err)
		}
	}

	// Update Application Insights key if changed, including clearing it.
	//
	// The `&& !plan.AppInsightsKey.IsNull()` guard that used to be here made removing the
	// attribute a silent no-op: no request was sent, so the key stayed on the environment
	// while state recorded it as gone. Because this setting is never read back from the
	// API, no later plan could detect the difference either. An empty key is the API's
	// representation of "not set", so the removal is expressible.
	if !plan.AppInsightsKey.Equal(state.AppInsightsKey) {
		if err := svc.SetAppInsightsKey(ctx, applicationFamily, environmentName, plan.AppInsightsKey.ValueString()); err != nil {
			return fmt.Errorf("updating app insights key: %w", err)
		}
	}

	// Update security group if changed.
	if !plan.SecurityGroupID.Equal(state.SecurityGroupID) {
		if plan.SecurityGroupID.IsNull() {
			if err := svc.ClearSecurityGroup(ctx, applicationFamily, environmentName); err != nil {
				return fmt.Errorf("clearing security group: %w", err)
			}
		} else {
			if err := svc.SetSecurityGroup(ctx, applicationFamily, environmentName, plan.SecurityGroupID.ValueString()); err != nil {
				return fmt.Errorf("updating security group: %w", err)
			}
		}
	}

	// Update M365 license access if explicitly changed (not null and not unknown).
	if !plan.AccessWithM365Licenses.Equal(state.AccessWithM365Licenses) && !plan.AccessWithM365Licenses.IsNull() && !plan.AccessWithM365Licenses.IsUnknown() {
		if err := svc.SetAccessWithM365Licenses(ctx, applicationFamily, environmentName, plan.AccessWithM365Licenses.ValueBool()); err != nil {
			return fmt.Errorf("updating M365 license access: %w", err)
		}
	}

	// Update app update cadence if changed.
	if !plan.AppUpdateCadence.Equal(state.AppUpdateCadence) && !plan.AppUpdateCadence.IsNull() {
		if err := svc.SetAppUpdateCadence(ctx, applicationFamily, environmentName, plan.AppUpdateCadence.ValueString()); err != nil {
			return fmt.Errorf("updating app update cadence: %w", err)
		}
	}

	// Update partner access if changed.
	if !plan.PartnerAccessStatus.Equal(state.PartnerAccessStatus) || !plan.AllowedPartnerTenantIDs.Equal(state.AllowedPartnerTenantIDs) {
		if !plan.PartnerAccessStatus.IsNull() {
			pa := &environmentsettings.PartnerAccessRequest{
				Status: plan.PartnerAccessStatus.ValueString(),
			}
			if plan.PartnerAccessStatus.ValueString() == "AllowSelectedPartnerTenants" && !plan.AllowedPartnerTenantIDs.IsNull() {
				var tenantIDs []string
				diags := plan.AllowedPartnerTenantIDs.ElementsAs(ctx, &tenantIDs, false)
				if diags.HasError() {
					return fmt.Errorf("reading allowed_partner_tenant_ids: %s", diags)
				}
				pa.AllowedPartnerTenantIDs = tenantIDs
			}
			if err := svc.SetPartnerAccess(ctx, applicationFamily, environmentName, pa); err != nil {
				return fmt.Errorf("updating partner access: %w", err)
			}
		}
	}

	return nil
}

// readEnvironmentSettings reads readable settings from the API and updates the nested model in place.
// Write-only fields (AppInsightsKey, AppUpdateCadence, PartnerAccess) are preserved from the current model.
func (r *EnvironmentResource) readEnvironmentSettings(ctx context.Context, svc *environmentsettings.Service, applicationFamily, environmentName string, settings *EnvironmentSettingsNestedModel) error {
	// Read update window settings.
	updateSettings, err := svc.GetUpdateSettings(ctx, applicationFamily, environmentName)
	if err != nil {
		return fmt.Errorf("reading update settings: %w", err)
	}
	if updateSettings != nil {
		if updateSettings.PreferredStartTime != nil {
			settings.UpdateWindowStartTime = types.StringValue(*updateSettings.PreferredStartTime)
		} else {
			settings.UpdateWindowStartTime = types.StringNull()
		}
		if updateSettings.PreferredEndTime != nil {
			settings.UpdateWindowEndTime = types.StringValue(*updateSettings.PreferredEndTime)
		} else {
			settings.UpdateWindowEndTime = types.StringNull()
		}
		if updateSettings.TimeZoneID != nil {
			settings.UpdateWindowTimeZone = types.StringValue(*updateSettings.TimeZoneID)
		} else {
			settings.UpdateWindowTimeZone = types.StringNull()
		}
	}

	// Read security group.
	securityGroup, err := svc.GetSecurityGroup(ctx, applicationFamily, environmentName)
	if err != nil {
		// Log but don't fail — security group may not be configured.
		tflog.Warn(ctx, "Could not read security group for inline settings", map[string]interface{}{
			"error": err.Error(),
		})
	} else if securityGroup != nil {
		settings.SecurityGroupID = types.StringValue(securityGroup.ID)
	} else {
		settings.SecurityGroupID = types.StringNull()
	}

	// Read M365 license access.
	m365Access, err := svc.GetAccessWithM365Licenses(ctx, applicationFamily, environmentName)
	if err != nil {
		// Leave the existing value in place. Nulling it on a transient failure destroyed a
		// value the user had configured: with a planned value of true and a 503 on the
		// read-back, the apply failed with "inconsistent result after apply" even though
		// the write had succeeded. The SecurityGroupID branch above does the same.
		tflog.Warn(ctx, "Could not read M365 license access for inline settings; keeping the current value", map[string]interface{}{
			"error": err.Error(),
		})
	} else if m365Access != nil {
		settings.AccessWithM365Licenses = types.BoolValue(m365Access.Enabled)
	} else {
		settings.AccessWithM365Licenses = types.BoolNull()
	}

	// AppInsightsKey, AppUpdateCadence, and PartnerAccess are write-only / require elevated permissions.
	// They are not read back from the API; the current state values are preserved by the caller.

	return nil
}
