// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package environmentsettings

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &EnvironmentSettingsResource{}
	_ resource.ResourceWithConfigure   = &EnvironmentSettingsResource{}
	_ resource.ResourceWithImportState = &EnvironmentSettingsResource{}
)

// NewEnvironmentSettingsResource is a helper function to simplify the provider implementation.
func NewEnvironmentSettingsResource() resource.Resource {
	return &EnvironmentSettingsResource{}
}

// EnvironmentSettingsResource is the resource implementation.
type EnvironmentSettingsResource struct {
	client *client.Client
}

// EnvironmentSettingsResourceModel maps the resource schema data.
type EnvironmentSettingsResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	ApplicationFamily       types.String `tfsdk:"application_family"`
	EnvironmentName         types.String `tfsdk:"environment_name"`
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

// Metadata returns the resource type name.
func (r *EnvironmentSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_settings"
}

// Schema defines the schema for the resource.
func (r *EnvironmentSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Business Central environment settings including update windows, telemetry, security groups, and access controls.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ARM-like resource ID (format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/settings)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_family": schema.StringAttribute{
				Description: "Family of the environment's application (e.g., 'BusinessCentral')",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_name": schema.StringAttribute{
				Description: "Name of the environment",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"update_window_start_time": schema.StringAttribute{
				Description: "Start time for the update window in HH:mm format (24-hour). Requires update_window_timezone to be set.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						timeFormatRegex,
						"must be in HH:mm format (e.g., '22:00')",
					),
				},
			},
			"update_window_end_time": schema.StringAttribute{
				Description: "End time for the update window in HH:mm format (24-hour). Requires update_window_timezone to be set. Must be at least 6 hours after start time.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						timeFormatRegex,
						"must be in HH:mm format (e.g., '06:00')",
					),
				},
			},
			"update_window_timezone": schema.StringAttribute{
				Description: "Windows time zone identifier for the update window (e.g., 'Pacific Standard Time', 'Eastern Standard Time'). Required if update_window_start_time or update_window_end_time are set.",
				Optional:    true,
			},
			"app_insights_key": schema.StringAttribute{
				Description: "Application Insights connection string or instrumentation key for environment telemetry. Warning: Setting this triggers an automatic environment restart.",
				Optional:    true,
				Sensitive:   true,
			},
			"security_group_id": schema.StringAttribute{
				Description: "Microsoft Entra (Azure AD) security group object ID to restrict environment access",
				Optional:    true,
			},
			"access_with_m365_licenses": schema.BoolAttribute{
				Description: "Whether users can access the environment with Microsoft 365 licenses (requires environment version 21.1+). Note: This setting may not be available on all environments.",
				Optional:    true,
				Computed:    true,
			},
			"app_update_cadence": schema.StringAttribute{
				Description: "How frequently AppSource apps should be updated. Valid values: 'Default', 'DuringMajorUpgrade', 'DuringMajorMinorUpgrade'",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Default", "DuringMajorUpgrade", "DuringMajorMinorUpgrade"),
				},
			},
			"partner_access_status": schema.StringAttribute{
				Description: "Partner access configuration. Valid values: 'Disabled', 'AllowAllPartnerTenants', 'AllowSelectedPartnerTenants'. Note: Only internal global administrators can modify this setting.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("Disabled", "AllowAllPartnerTenants", "AllowSelectedPartnerTenants"),
				},
			},
			"allowed_partner_tenant_ids": schema.ListAttribute{
				Description: "List of partner tenant IDs allowed to access the environment. Only used when partner_access_status is 'AllowSelectedPartnerTenants'",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *EnvironmentSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *EnvironmentSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentSettingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the ID to the ARM-like format.
	tenantID := r.client.GetTenantID()
	plan.ID = types.StringValue(BuildEnvironmentSettingsID(
		tenantID,
		plan.ApplicationFamily.ValueString(),
		plan.EnvironmentName.ValueString(),
	))

	// Create service.
	svc := NewService(r.client)

	// Apply update window settings if provided.
	if !plan.UpdateWindowStartTime.IsNull() || !plan.UpdateWindowEndTime.IsNull() || !plan.UpdateWindowTimeZone.IsNull() {
		if err := r.applyUpdateSettings(ctx, svc, &plan); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Update Window",
				"Could not set update window settings: "+err.Error(),
			)
			return
		}
	}

	// Apply Application Insights key if provided.
	if !plan.AppInsightsKey.IsNull() {
		if err := svc.SetAppInsightsKey(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.AppInsightsKey.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Application Insights Key",
				"Could not set Application Insights key (this triggers environment restart): "+err.Error(),
			)
			return
		}
	}

	// Apply security group if provided.
	if !plan.SecurityGroupID.IsNull() {
		if err := svc.SetSecurityGroup(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.SecurityGroupID.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Security Group",
				"Could not set security group: "+err.Error(),
			)
			return
		}
	}

	// Apply M365 license access if provided.
	if !plan.AccessWithM365Licenses.IsNull() {
		if err := svc.SetAccessWithM365Licenses(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.AccessWithM365Licenses.ValueBool()); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting M365 License Access",
				"Could not set M365 license access: "+err.Error(),
			)
			return
		}
	}

	// Apply app update cadence if provided.
	if !plan.AppUpdateCadence.IsNull() {
		if err := svc.SetAppUpdateCadence(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.AppUpdateCadence.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting App Update Cadence",
				"Could not set app update cadence: "+err.Error(),
			)
			return
		}
	}

	// Apply partner access if provided.
	if !plan.PartnerAccessStatus.IsNull() {
		if err := r.applyPartnerAccessSettings(ctx, svc, &plan); err != nil {
			resp.Diagnostics.AddError(
				"Error Setting Partner Access",
				"Could not set partner access settings: "+err.Error(),
			)
			return
		}
	}

	// Save data into Terraform state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *EnvironmentSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentSettingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc := NewService(r.client)

	// Read update settings.
	updateSettings, err := svc.GetUpdateSettings(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Update Settings",
			"Could not read update settings: "+err.Error(),
		)
		return
	}

	if updateSettings != nil {
		if updateSettings.PreferredStartTime != nil {
			state.UpdateWindowStartTime = types.StringValue(*updateSettings.PreferredStartTime)
		}
		if updateSettings.PreferredEndTime != nil {
			state.UpdateWindowEndTime = types.StringValue(*updateSettings.PreferredEndTime)
		}
		if updateSettings.TimeZoneID != nil {
			state.UpdateWindowTimeZone = types.StringValue(*updateSettings.TimeZoneID)
		}
	}

	// Read security group (404/NoContent is expected if not set)
	securityGroup, err := svc.GetSecurityGroup(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString())
	if err != nil {
		// Log but don't fail - security group may not be set.
		resp.Diagnostics.AddWarning(
			"Could not read security group",
			err.Error(),
		)
	}
	if securityGroup != nil {
		state.SecurityGroupID = types.StringValue(securityGroup.ID)
	} else {
		state.SecurityGroupID = types.StringNull()
	}

	// Read M365 license access (may not be available on older environments)
	m365Access, err := svc.GetAccessWithM365Licenses(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString())
	if err != nil {
		// Only warn on actual errors, not when feature is unavailable.
		resp.Diagnostics.AddWarning(
			"Could not read M365 license access setting",
			err.Error(),
		)
		state.AccessWithM365Licenses = types.BoolNull()
	} else if m365Access != nil {
		state.AccessWithM365Licenses = types.BoolValue(m365Access.Enabled)
	} else {
		// Feature not available or not configured - set to null.
		state.AccessWithM365Licenses = types.BoolNull()
	}

	// Note: AppInsightsKey cannot be read back (write-only)
	// Note: AppUpdateCadence has no GET endpoint.
	// Note: PartnerAccess requires global admin permissions.

	// Save updated data into Terraform state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *EnvironmentSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentSettingsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state EnvironmentSettingsResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc := NewService(r.client)

	// Update window settings if changed.
	if !plan.UpdateWindowStartTime.Equal(state.UpdateWindowStartTime) ||
		!plan.UpdateWindowEndTime.Equal(state.UpdateWindowEndTime) ||
		!plan.UpdateWindowTimeZone.Equal(state.UpdateWindowTimeZone) {
		if err := r.applyUpdateSettings(ctx, svc, &plan); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Update Window",
				"Could not update update window settings: "+err.Error(),
			)
			return
		}
	}

	// Update Application Insights key if changed.
	if !plan.AppInsightsKey.Equal(state.AppInsightsKey) && !plan.AppInsightsKey.IsNull() {
		if err := svc.SetAppInsightsKey(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.AppInsightsKey.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Application Insights Key",
				"Could not update Application Insights key: "+err.Error(),
			)
			return
		}
	}

	// Update security group if changed.
	if !plan.SecurityGroupID.Equal(state.SecurityGroupID) {
		if plan.SecurityGroupID.IsNull() {
			// Clear the security group.
			if err := svc.ClearSecurityGroup(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString()); err != nil {
				resp.Diagnostics.AddError(
					"Error Clearing Security Group",
					"Could not clear security group: "+err.Error(),
				)
				return
			}
		} else {
			// Set new security group.
			if err := svc.SetSecurityGroup(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.SecurityGroupID.ValueString()); err != nil {
				resp.Diagnostics.AddError(
					"Error Updating Security Group",
					"Could not update security group: "+err.Error(),
				)
				return
			}
		}
	}

	// Update M365 license access if changed.
	if !plan.AccessWithM365Licenses.Equal(state.AccessWithM365Licenses) && !plan.AccessWithM365Licenses.IsNull() {
		if err := svc.SetAccessWithM365Licenses(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.AccessWithM365Licenses.ValueBool()); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating M365 License Access",
				"Could not update M365 license access: "+err.Error(),
			)
			return
		}
	}

	// Update app update cadence if changed.
	if !plan.AppUpdateCadence.Equal(state.AppUpdateCadence) && !plan.AppUpdateCadence.IsNull() {
		if err := svc.SetAppUpdateCadence(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.AppUpdateCadence.ValueString()); err != nil {
			resp.Diagnostics.AddError(
				"Error Updating App Update Cadence",
				"Could not update app update cadence: "+err.Error(),
			)
			return
		}
	}

	// Update partner access if changed.
	if !plan.PartnerAccessStatus.Equal(state.PartnerAccessStatus) || !plan.AllowedPartnerTenantIDs.Equal(state.AllowedPartnerTenantIDs) {
		if !plan.PartnerAccessStatus.IsNull() {
			if err := r.applyPartnerAccessSettings(ctx, svc, &plan); err != nil {
				resp.Diagnostics.AddError(
					"Error Updating Partner Access",
					"Could not update partner access settings: "+err.Error(),
				)
				return
			}
		}
	}

	// Save updated data into Terraform state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *EnvironmentSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentSettingsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Environment settings are tied to the environment lifecycle.
	// Deleting the resource doesn't delete the settings, just removes from Terraform state.
	// Settings will revert to defaults or remain as configured.

	resp.Diagnostics.AddWarning(
		"Environment Settings Not Reset",
		"Deleting this resource removes it from Terraform state but does not reset the environment settings to defaults. "+
			"The settings will remain as configured on the environment.",
	)
}

// ImportState imports the resource state.
func (r *EnvironmentSettingsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using format: applicationFamily/environmentName.
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	// Parse the ID to set application_family and environment_name.
	// This will be handled in the Read operation.
}

// Helper functions.

func (r *EnvironmentSettingsResource) applyUpdateSettings(ctx context.Context, svc *Service, plan *EnvironmentSettingsResourceModel) error {
	settings := &UpdateSettings{}

	if !plan.UpdateWindowStartTime.IsNull() {
		startTime := plan.UpdateWindowStartTime.ValueString()
		settings.PreferredStartTime = &startTime
	}

	if !plan.UpdateWindowEndTime.IsNull() {
		endTime := plan.UpdateWindowEndTime.ValueString()
		settings.PreferredEndTime = &endTime
	}

	if !plan.UpdateWindowTimeZone.IsNull() {
		timezone := plan.UpdateWindowTimeZone.ValueString()
		settings.TimeZoneID = &timezone
	}

	_, err := svc.SetUpdateSettings(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), settings)
	return err
}

func (r *EnvironmentSettingsResource) applyPartnerAccessSettings(ctx context.Context, svc *Service, plan *EnvironmentSettingsResourceModel) error {
	settings := &PartnerAccessRequest{
		Status: plan.PartnerAccessStatus.ValueString(),
	}

	if plan.PartnerAccessStatus.ValueString() == "AllowSelectedPartnerTenants" && !plan.AllowedPartnerTenantIDs.IsNull() {
		var tenantIDs []string
		plan.AllowedPartnerTenantIDs.ElementsAs(ctx, &tenantIDs, false)
		settings.AllowedPartnerTenantIDs = tenantIDs
	}

	return svc.SetPartnerAccess(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), settings)
}

// Regex for time format validation (HH:mm).
var timeFormatRegex = regexp.MustCompile(`^([01]\d|2[0-3]):([0-5]\d)$`)
