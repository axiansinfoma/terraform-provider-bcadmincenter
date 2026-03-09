// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &UpdateScheduleResource{}
	_ resource.ResourceWithConfigure   = &UpdateScheduleResource{}
	_ resource.ResourceWithImportState = &UpdateScheduleResource{}
)

// NewUpdateScheduleResource is a helper function to simplify the provider implementation.
func NewUpdateScheduleResource() resource.Resource {
	return &UpdateScheduleResource{}
}

// UpdateScheduleResource manages an explicitly scheduled environment upgrade.
type UpdateScheduleResource struct {
	client *client.Client
}

// UpdateScheduleResourceModel describes the resource data model.
type UpdateScheduleResourceModel struct {
	ID                       types.String `tfsdk:"id"`
	AADTenantID              types.String `tfsdk:"aad_tenant_id"`
	ApplicationFamily        types.String `tfsdk:"application_family"`
	EnvironmentName          types.String `tfsdk:"environment_name"`
	TargetVersion            types.String `tfsdk:"target_version"`
	ScheduledDatetime        types.String `tfsdk:"scheduled_datetime"`
	IgnoreUpdateWindow       types.Bool   `tfsdk:"ignore_update_window"`
	UpdateStatus             types.String `tfsdk:"update_status"`
	RolloutStatus            types.String `tfsdk:"rollout_status"`
	LatestSelectableDatetime types.String `tfsdk:"latest_selectable_datetime"`
}

// Metadata returns the resource type name.
func (r *UpdateScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_update_schedule"
}

// Schema defines the schema for the resource.
func (r *UpdateScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an explicitly scheduled upgrade for a Business Central environment.\n\n" +
			"Use this resource instead of `application_version` on `bcadmincenter_environment` when you need " +
			"explicit scheduling control over which version to target and when the upgrade should run.\n\n" +
			"~> **Note:** Do **not** use `application_version` on `bcadmincenter_environment` and " +
			"`bcadmincenter_environment_update_schedule` for the same environment simultaneously.\n\n" +
			"~> **Warning:** Destroying this resource removes it from Terraform state only. " +
			"No API call is made and the scheduled upgrade on the Business Central side is **not** cancelled.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ARM-like resource ID for the update schedule.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aad_tenant_id": schema.StringAttribute{
				MarkdownDescription: "The Azure AD tenant ID. If not specified, the provider's configured tenant ID is used.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_family": schema.StringAttribute{
				MarkdownDescription: "The application family for the environment (e.g. `\"BusinessCentral\"`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_name": schema.StringAttribute{
				MarkdownDescription: "The name of the target environment.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_version": schema.StringAttribute{
				MarkdownDescription: "The target application version to upgrade to (e.g. `\"26.1\"`). Must match a value from the environment's available updates.",
				Required:            true,
			},
			"scheduled_datetime": schema.StringAttribute{
				MarkdownDescription: "The RFC3339 datetime at which the upgrade should run. If omitted, the upgrade runs in the next update window.",
				Optional:            true,
			},
			"ignore_update_window": schema.BoolAttribute{
				MarkdownDescription: "When `true`, the upgrade may start at `scheduled_datetime` even if outside the environment's configured update window. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"update_status": schema.StringAttribute{
				MarkdownDescription: "The current update status (e.g. `\"scheduled\"`, `\"running\"`, `\"failed\"`).",
				Computed:            true,
			},
			"rollout_status": schema.StringAttribute{
				MarkdownDescription: "The rollout status (e.g. `\"Active\"`, `\"UnderMaintenance\"`, `\"Postponed\"`).",
				Computed:            true,
			},
			"latest_selectable_datetime": schema.StringAttribute{
				MarkdownDescription: "The latest datetime the update can be scheduled to.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *UpdateScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// Create schedules a version upgrade.
func (r *UpdateScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UpdateScheduleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := r.client.GetTenantID()
	if !plan.AADTenantID.IsNull() && !plan.AADTenantID.IsUnknown() {
		tenantID = plan.AADTenantID.ValueString()
	}

	svc := NewService(r.client.ForTenant(tenantID))

	scheduledDatetime := ""
	if !plan.ScheduledDatetime.IsNull() && !plan.ScheduledDatetime.IsUnknown() {
		scheduledDatetime = plan.ScheduledDatetime.ValueString()
	}

	tflog.Debug(ctx, "Scheduling environment update", map[string]interface{}{
		"application_family":   plan.ApplicationFamily.ValueString(),
		"environment_name":     plan.EnvironmentName.ValueString(),
		"target_version":       plan.TargetVersion.ValueString(),
		"scheduled_datetime":   scheduledDatetime,
		"ignore_update_window": plan.IgnoreUpdateWindow.ValueBool(),
	})

	if err := svc.ScheduleUpdateVersion(
		ctx,
		plan.ApplicationFamily.ValueString(),
		plan.EnvironmentName.ValueString(),
		plan.TargetVersion.ValueString(),
		scheduledDatetime,
		plan.IgnoreUpdateWindow.ValueBool(),
	); err != nil {
		resp.Diagnostics.AddError(
			"Error scheduling environment update",
			fmt.Sprintf("Could not schedule update to version %s: %s", plan.TargetVersion.ValueString(), err),
		)
		return
	}

	plan.AADTenantID = types.StringValue(tenantID)
	plan.ID = types.StringValue(BuildUpdateScheduleID(tenantID, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString()))

	// Read back the scheduled update to populate computed fields.
	r.readAndPopulate(ctx, svc, &plan)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *UpdateScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UpdateScheduleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc := NewService(r.client.ForTenant(state.AADTenantID.ValueString()))

	updates, err := svc.GetUpdates(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading environment updates",
			fmt.Sprintf("Could not read updates for environment %s: %s", state.EnvironmentName.ValueString(), err),
		)
		return
	}

	selectedUpdate := findSelectedUpdate(updates)
	if selectedUpdate == nil {
		// No selected update: remove resource from state.
		tflog.Debug(ctx, "No selected update found; removing update schedule resource from state")
		resp.State.RemoveResource(ctx)
		return
	}

	// Apply drift detection.
	switch selectedUpdate.UpdateStatus {
	case UpdateStatusScheduled, UpdateStatusRunning:
		// No drift: populate state from selected update.
		populateModelFromUpdate(&state, selectedUpdate)
	case UpdateStatusFailed:
		// Drift: force re-create on next apply by keeping the target version in state
		// but marking update_status as failed so the user can see it.
		populateModelFromUpdate(&state, selectedUpdate)
	default:
		populateModelFromUpdate(&state, selectedUpdate)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update handles changes to the scheduled upgrade.
func (r *UpdateScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state UpdateScheduleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc := NewService(r.client.ForTenant(state.AADTenantID.ValueString()))

	scheduledDatetime := ""
	if !plan.ScheduledDatetime.IsNull() && !plan.ScheduledDatetime.IsUnknown() {
		scheduledDatetime = plan.ScheduledDatetime.ValueString()
	}

	versionChanged := !plan.TargetVersion.Equal(state.TargetVersion)

	if versionChanged {
		// Re-select with new target version (API automatically deselects previous).
		tflog.Debug(ctx, "Rescheduling update with new target version", map[string]interface{}{
			"old_version": state.TargetVersion.ValueString(),
			"new_version": plan.TargetVersion.ValueString(),
		})

		if err := svc.ScheduleUpdateVersion(
			ctx,
			state.ApplicationFamily.ValueString(),
			state.EnvironmentName.ValueString(),
			plan.TargetVersion.ValueString(),
			scheduledDatetime,
			plan.IgnoreUpdateWindow.ValueBool(),
		); err != nil {
			resp.Diagnostics.AddError(
				"Error rescheduling environment update",
				fmt.Sprintf("Could not reschedule update to version %s: %s", plan.TargetVersion.ValueString(), err),
			)
			return
		}
	} else {
		// Only datetime or ignore_update_window changed: update schedule details without reselecting.
		tflog.Debug(ctx, "Updating schedule details without reselecting version")

		if err := svc.UpdateScheduleDetails(
			ctx,
			state.ApplicationFamily.ValueString(),
			state.EnvironmentName.ValueString(),
			state.TargetVersion.ValueString(),
			scheduledDatetime,
			plan.IgnoreUpdateWindow.ValueBool(),
		); err != nil {
			resp.Diagnostics.AddError(
				"Error updating schedule details",
				fmt.Sprintf("Could not update schedule details: %s", err),
			)
			return
		}
	}

	plan.AADTenantID = state.AADTenantID
	plan.ID = state.ID

	// Read back the updated schedule.
	r.readAndPopulate(ctx, svc, &plan)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the resource from Terraform state only — no API call is made.
// The Business Central Admin Center API does not support deselecting an update
// without selecting a new one, so destroying this resource does NOT cancel or
// remove the scheduled upgrade on the Business Central side.
func (r *UpdateScheduleResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Intentionally empty: state-only removal.
}

// ImportState imports an existing resource into Terraform state.
func (r *UpdateScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tenantID, applicationFamily, environmentName, err := ParseUpdateScheduleID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected ARM-like resource ID in format '/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/updateSchedule', got: %s\nError: %s",
				req.ID, err.Error()),
		)
		return
	}

	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
	resp.State.SetAttribute(ctx, path.Root("aad_tenant_id"), tenantID)
	resp.State.SetAttribute(ctx, path.Root("application_family"), applicationFamily)
	resp.State.SetAttribute(ctx, path.Root("environment_name"), environmentName)
}

// readAndPopulate calls GetUpdates and populates computed fields from the selected update.
func (r *UpdateScheduleResource) readAndPopulate(ctx context.Context, svc *Service, model *UpdateScheduleResourceModel) {
	updates, err := svc.GetUpdates(ctx, model.ApplicationFamily.ValueString(), model.EnvironmentName.ValueString())
	if err != nil {
		// Non-fatal for create/update: just leave computed fields as unknown.
		tflog.Warn(ctx, "Could not read back update schedule", map[string]interface{}{"error": err.Error()})
		return
	}

	selectedUpdate := findSelectedUpdate(updates)
	if selectedUpdate != nil {
		populateModelFromUpdate(model, selectedUpdate)
	}
}

// findSelectedUpdate returns the first selected update from the list, or nil.
func findSelectedUpdate(updates []EnvironmentUpdate) *EnvironmentUpdate {
	for i := range updates {
		if updates[i].Selected {
			return &updates[i]
		}
	}
	return nil
}

// populateModelFromUpdate fills the model's computed fields from an EnvironmentUpdate entry.
func populateModelFromUpdate(model *UpdateScheduleResourceModel, update *EnvironmentUpdate) {
	model.TargetVersion = types.StringValue(update.TargetVersion)

	if update.UpdateStatus != "" {
		model.UpdateStatus = types.StringValue(update.UpdateStatus)
	} else {
		model.UpdateStatus = types.StringNull()
	}

	if update.ScheduleDetails != nil {
		if update.ScheduleDetails.RolloutStatus != "" {
			model.RolloutStatus = types.StringValue(update.ScheduleDetails.RolloutStatus)
		} else {
			model.RolloutStatus = types.StringNull()
		}

		if update.ScheduleDetails.LatestSelectableDateTime != "" {
			model.LatestSelectableDatetime = types.StringValue(update.ScheduleDetails.LatestSelectableDateTime)
		} else {
			model.LatestSelectableDatetime = types.StringNull()
		}

		if update.ScheduleDetails.SelectedDateTime != "" {
			model.ScheduledDatetime = types.StringValue(update.ScheduleDetails.SelectedDateTime)
		}
		// If SelectedDateTime is empty, the user-configured scheduled_datetime is preserved as-is.
	} else {
		model.RolloutStatus = types.StringNull()
		model.LatestSelectableDatetime = types.StringNull()
	}
}
