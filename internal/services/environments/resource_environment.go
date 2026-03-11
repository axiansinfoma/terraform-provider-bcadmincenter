// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	_ resource.Resource                = &EnvironmentResource{}
	_ resource.ResourceWithConfigure   = &EnvironmentResource{}
	_ resource.ResourceWithImportState = &EnvironmentResource{}
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
				MarkdownDescription: "The Azure AD tenant ID for the environment. If not specified, the value is read from the API response.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"pending_upgrade_version": schema.StringAttribute{
				MarkdownDescription: "The target version of a currently selected/scheduled or running upgrade. " +
					"Empty when no upgrade is in progress. " +
					"While non-empty, `application_version` is suppressed to this value so no drift is reported.",
				Computed: true,
			},
			"pending_upgrade_scheduled_for": schema.StringAttribute{
				MarkdownDescription: "The RFC3339 datetime at which the pending upgrade is scheduled to run. " +
					"Empty when the upgrade will run at the next update window or when no upgrade is pending.",
				Computed: true,
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Optional environment settings block. When specified, the settings are applied to the environment after creation and managed inline.",
				Optional:            true,
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

	// Determine timeout.
	// TODO: Parse timeout from plan.Timeouts if needed.
	timeout := 60 * time.Minute // default

	// Wait for the operation to complete.
	// Use ProductFamily from operation response if available, otherwise use the plan value.
	appFamily := operation.ProductFamily
	if appFamily == "" {
		appFamily = operation.ApplicationFamily
	}
	if appFamily == "" {
		appFamily = plan.ApplicationFamily.ValueString()
	}

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

	if err := svc.WaitForOperation(ctx, appFamily, envName, operation.ID, timeout); err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for environment creation",
			fmt.Sprintf("Environment creation failed: %s", err),
		)
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
				fmt.Sprintf("Could not read environment after creation: %s", err),
			)
			return
		}

		tflog.Debug(ctx, "Environment status check", map[string]interface{}{
			"status": env.Status,
		})

		if env.Status == "Active" {
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
						"Could not apply settings after environment creation: "+err.Error(),
					)
					return
				}
				// Read back readable settings (update_window, security_group, m365 access).
				if err := r.readEnvironmentSettings(ctx, settingsSvc, plan.ApplicationFamily.ValueString(), envName, plan.Settings); err != nil {
					resp.Diagnostics.AddError(
						"Error reading environment settings",
						"Could not read settings after applying: "+err.Error(),
					)
					return
				}
			}

			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}

		// Check for failed states.
		if env.Status == "Failed" || env.Status == "Suspended" {
			resp.Diagnostics.AddError(
				"Environment creation failed",
				fmt.Sprintf("Environment entered %s state during creation", env.Status),
			)
			return
		}

		// Wait for next tick or timeout.
		select {
		case <-envTimeout.Done():
			resp.Diagnostics.AddError(
				"Timeout waiting for environment",
				fmt.Sprintf("Environment did not become Active within %v (current status: %s)", timeout, env.Status),
			)
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
	versionChanged := !plan.ApplicationVersion.Equal(state.ApplicationVersion)
	windowChanged := !plan.IgnoreUpdateWindow.Equal(state.IgnoreUpdateWindow)
	settingsChanged := settingsBlockChanged(plan.Settings, state.Settings)

	if !versionChanged && !windowChanged && !settingsChanged {
		// Nothing to do; copy plan to state.
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Apply inline settings changes if the block was added, modified, or removed.
	if settingsChanged && plan.Settings != nil {
		settingsSvc := environmentsettings.NewService(r.client.ForTenant(state.AADTenantID.ValueString()))
		if err := r.applyEnvironmentSettingsChanges(ctx, settingsSvc, state.ApplicationFamily.ValueString(), state.Name.ValueString(), plan.Settings, state.Settings); err != nil {
			resp.Diagnostics.AddError(
				"Error updating environment settings",
				"Could not update inline settings: "+err.Error(),
			)
			return
		}
		// Read back readable settings to keep state consistent.
		if err := r.readEnvironmentSettings(ctx, settingsSvc, state.ApplicationFamily.ValueString(), state.Name.ValueString(), plan.Settings); err != nil {
			resp.Diagnostics.AddError(
				"Error reading environment settings",
				"Could not read settings after update: "+err.Error(),
			)
			return
		}
	}

	if !versionChanged && !windowChanged {
		// Only settings changed; persist the plan (with refreshed settings).
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

	// Determine timeout.
	// TODO: Parse timeout from state.Timeouts if needed.
	timeout := 60 * time.Minute // default

	// Wait for the operation to complete.
	// Use ProductFamily from operation response if available, otherwise fall back to state.
	appFamily := operation.ProductFamily
	if appFamily == "" {
		appFamily = operation.ApplicationFamily
	}
	if appFamily == "" {
		appFamily = state.ApplicationFamily.ValueString()
	}

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
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
	resp.State.SetAttribute(ctx, path.Root("application_family"), applicationFamily)
	resp.State.SetAttribute(ctx, path.Root("name"), environmentName)
	resp.State.SetAttribute(ctx, path.Root("aad_tenant_id"), tenantID)
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

	// Azure region is not returned by the API, so always set to null.
	model.AzureRegion = types.StringNull()

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

// normalizeApplicationVersion returns the short form of priorVersion when the API returned
// the full build version. This prevents spurious drift when users configure versions in
// "major.minor" format (e.g., "27.1") while the API stores the full build version
// (e.g., "27.1.41698.41831").
//
// A "." separator check is used to avoid incorrectly matching "27.1" against "27.10.xxx".
func normalizeApplicationVersion(priorVersion, apiVersion string) string {
	if priorVersion == "" || apiVersion == "" {
		return apiVersion
	}
	if apiVersion == priorVersion || strings.HasPrefix(apiVersion, priorVersion+".") {
		return priorVersion
	}
	return apiVersion
}

// settingsBlockChanged returns true if the settings block differs between plan and state.
// Both nil means no change. One nil and one non-nil means change (block added/removed).
func settingsBlockChanged(plan, state *EnvironmentSettingsNestedModel) bool {
	if plan == nil && state == nil {
		return false
	}
	if plan == nil || state == nil {
		return true
	}
	return !plan.UpdateWindowStartTime.Equal(state.UpdateWindowStartTime) ||
		!plan.UpdateWindowEndTime.Equal(state.UpdateWindowEndTime) ||
		!plan.UpdateWindowTimeZone.Equal(state.UpdateWindowTimeZone) ||
		!plan.AppInsightsKey.Equal(state.AppInsightsKey) ||
		!plan.SecurityGroupID.Equal(state.SecurityGroupID) ||
		!plan.AccessWithM365Licenses.Equal(state.AccessWithM365Licenses) ||
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

	// Apply M365 license access if provided.
	if !settings.AccessWithM365Licenses.IsNull() {
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

	// Update Application Insights key if changed.
	if !plan.AppInsightsKey.Equal(state.AppInsightsKey) && !plan.AppInsightsKey.IsNull() {
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

	// Update M365 license access if changed.
	if !plan.AccessWithM365Licenses.Equal(state.AccessWithM365Licenses) && !plan.AccessWithM365Licenses.IsNull() {
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
		tflog.Warn(ctx, "Could not read M365 license access for inline settings", map[string]interface{}{
			"error": err.Error(),
		})
		settings.AccessWithM365Licenses = types.BoolNull()
	} else if m365Access != nil {
		settings.AccessWithM365Licenses = types.BoolValue(m365Access.Enabled)
	} else {
		settings.AccessWithM365Licenses = types.BoolNull()
	}

	// AppInsightsKey, AppUpdateCadence, and PartnerAccess are write-only / require elevated permissions.
	// They are not read back from the API; the current state values are preserved by the caller.

	return nil
}
