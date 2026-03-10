// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentapps

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &EnvironmentAppResource{}
	_ resource.ResourceWithConfigure   = &EnvironmentAppResource{}
	_ resource.ResourceWithImportState = &EnvironmentAppResource{}
)

// NewEnvironmentAppResource is a helper function to simplify the provider implementation.
func NewEnvironmentAppResource() resource.Resource {
	return &EnvironmentAppResource{}
}

// EnvironmentAppResource is the resource implementation.
type EnvironmentAppResource struct {
	client *client.Client
}

// EnvironmentAppResourceModel describes the resource data model.
type EnvironmentAppResourceModel struct {
	ID                                types.String `tfsdk:"id"`
	AADTenantID                       types.String `tfsdk:"aad_tenant_id"`
	ApplicationFamily                 types.String `tfsdk:"application_family"`
	EnvironmentName                   types.String `tfsdk:"environment_name"`
	AppID                             types.String `tfsdk:"app_id"`
	TargetVersion                     types.String `tfsdk:"target_version"`
	AllowPreviewVersion               types.Bool   `tfsdk:"allow_preview_version"`
	InstallOrUpdateNeededDependencies types.Bool   `tfsdk:"install_or_update_needed_dependencies"`
	AcceptIsvEula                     types.Bool   `tfsdk:"accept_isv_eula"`
	LanguageID                        types.String `tfsdk:"language_id"`
	UseEnvironmentUpdateWindow        types.Bool   `tfsdk:"use_environment_update_window"`
	PendingTargetVersion              types.String `tfsdk:"pending_target_version"`
	PendingOperationID                types.String `tfsdk:"pending_operation_id"`
	Name                              types.String `tfsdk:"name"`
	Publisher                         types.String `tfsdk:"publisher"`
	PublishedAs                       types.String `tfsdk:"published_as"`
	Status                            types.String `tfsdk:"status"`
	Timeouts                          types.Object `tfsdk:"timeouts"`
}

// Metadata returns the resource type name.
func (r *EnvironmentAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_app"
}

// Schema defines the schema for the resource.
func (r *EnvironmentAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the install/update/uninstall lifecycle for an app in a Business Central environment.\n\n" +
			"This resource installs a Business Central app and manages its version. " +
			"Install, update and uninstall are asynchronous operations that can take several minutes to complete.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ARM-like resource ID (format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/apps/{appId})",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aad_tenant_id": schema.StringAttribute{
				MarkdownDescription: "The Azure AD tenant ID. If not specified, defaults to the provider's configured tenant ID.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"application_family": schema.StringAttribute{
				MarkdownDescription: "The application family for the environment (e.g. `\"BusinessCentral\"`). Changing this forces a new resource to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_name": schema.StringAttribute{
				MarkdownDescription: "The name of the target environment. Changing this forces a new resource to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "The app GUID to install. Changing this forces a new resource to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_version": schema.StringAttribute{
				MarkdownDescription: "The target app version to install or update to (e.g. `\"1.2.3.4\"`). " +
					"Omit or leave null to install the latest available version. " +
					"Changing this to a higher version schedules an in-place update. " +
					"Downgrading is blocked at plan time.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					utils.NoDowngradeAppVersion(),
				},
			},
			"allow_preview_version": schema.BoolAttribute{
				MarkdownDescription: "When `true`, allows installing preview versions of the app. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"install_or_update_needed_dependencies": schema.BoolAttribute{
				MarkdownDescription: "When `true`, automatically installs or updates app dependencies. Defaults to `true`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"accept_isv_eula": schema.BoolAttribute{
				MarkdownDescription: "When `true`, accepts the ISV End User License Agreement (EULA) for the app. Required for some ISV apps. Defaults to `false`. Changing this forces a new resource to be created.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"language_id": schema.StringAttribute{
				MarkdownDescription: "The language identifier for the app installation (e.g. `\"en-US\"`). If not specified, the default language is used. Changing this forces a new resource to be created.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"use_environment_update_window": schema.BoolAttribute{
				MarkdownDescription: "When `true` (default), update and uninstall operations respect the environment's configured update window. Set to `false` to bypass the update window and apply the operation immediately.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"pending_target_version": schema.StringAttribute{
				MarkdownDescription: "The target version of a currently scheduled or running update. " +
					"Non-empty when an update has been deferred to the environment's update window. " +
					"While non-empty, `target_version` is suppressed to this value so no drift is reported.",
				Computed: true,
			},
			"pending_operation_id": schema.StringAttribute{
				MarkdownDescription: "The BC operation ID of a currently scheduled (deferred) update. " +
					"Non-empty when an update has been deferred to the environment's update window. " +
					"Used internally to cancel and reschedule the operation when `use_environment_update_window` changes.",
				Computed: true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The display name of the app (read from the API).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"publisher": schema.StringAttribute{
				MarkdownDescription: "The publisher of the app (read from the API).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"published_as": schema.StringAttribute{
				MarkdownDescription: "How the app is published (e.g. `\"Global\"`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current install status of the app (e.g. `\"installed\"`, `\"installFailed\"`, `\"updateFailed\"`). " +
					"When the status is `\"installFailed\"` or `\"updateFailed\"`, the resource will be replaced on the next apply.",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplaceIf(
						func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
							switch req.StateValue.ValueString() {
							case AppStatusInstallFailed, AppStatusUpdateFailed:
								resp.RequiresReplace = true
							}
						},
						"Forces replacement when the app is in a terminal failure state (installFailed or updateFailed).",
						"Forces replacement when the app is in a terminal failure state (`installFailed` or `updateFailed`).",
					),
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
func (r *EnvironmentAppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *EnvironmentAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentAppResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Installing BC Admin Center environment app", map[string]interface{}{
		"application_family": plan.ApplicationFamily.ValueString(),
		"environment_name":   plan.EnvironmentName.ValueString(),
		"app_id":             plan.AppID.ValueString(),
	})

	tenantID := r.client.GetTenantID()
	if !plan.AADTenantID.IsNull() && !plan.AADTenantID.IsUnknown() {
		tenantID = plan.AADTenantID.ValueString()
	}
	plan.AADTenantID = types.StringValue(tenantID)

	svc := NewService(r.client.ForTenant(tenantID))

	installReq := &InstallAppRequest{
		AllowPreviewVersion:               plan.AllowPreviewVersion.ValueBool(),
		InstallOrUpdateNeededDependencies: plan.InstallOrUpdateNeededDependencies.ValueBool(),
		AcceptIsvEula:                     plan.AcceptIsvEula.ValueBool(),
	}
	if !plan.TargetVersion.IsNull() && !plan.TargetVersion.IsUnknown() && plan.TargetVersion.ValueString() != "" {
		installReq.TargetVersion = plan.TargetVersion.ValueString()
	}
	if !plan.LanguageID.IsNull() && !plan.LanguageID.IsUnknown() && plan.LanguageID.ValueString() != "" {
		installReq.LanguageID = plan.LanguageID.ValueString()
	}

	operation, err := svc.Install(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.AppID.ValueString(), installReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error installing app",
			fmt.Sprintf("Could not install app %s: %s", plan.AppID.ValueString(), err),
		)
		return
	}

	timeout := 60 * time.Minute

	if _, err := svc.WaitForOperation(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), operation.ID, timeout, false); err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for app installation",
			fmt.Sprintf("App installation failed: %s", err),
		)
		return
	}

	// Set the ARM resource ID.
	plan.ID = types.StringValue(BuildEnvironmentAppID(tenantID, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.AppID.ValueString()))

	// Populate computed fields from API.
	app, err := svc.GetByID(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), plan.AppID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading installed app",
			fmt.Sprintf("Could not read app after installation: %s", err),
		)
		return
	}
	if app != nil {
		updateModelFromApp(&plan, app)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *EnvironmentAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentAppResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading BC Admin Center environment app", map[string]interface{}{
		"application_family": state.ApplicationFamily.ValueString(),
		"environment_name":   state.EnvironmentName.ValueString(),
		"app_id":             state.AppID.ValueString(),
	})

	svc := NewService(r.client.ForTenant(state.AADTenantID.ValueString()))

	app, err := svc.GetByID(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString(), state.AppID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading app",
			fmt.Sprintf("Could not read app %s: %s", state.AppID.ValueString(), err),
		)
		return
	}

	if app == nil {
		// App is no longer installed — remove from state.
		tflog.Debug(ctx, "App no longer installed, removing from state", map[string]interface{}{
			"app_id": state.AppID.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	// Capture the pending target version and operation ID before updateModelFromApp resets them.
	priorTargetVersion := state.TargetVersion
	priorPending := state.PendingTargetVersion
	priorPendingOpID := state.PendingOperationID

	updateModelFromApp(&state, app)

	// Drift suppression — mirrors the environment resource's applyUpdatesDriftDetection.
	// Use pending_target_version from state as the "selected update" signal, since
	// the app API has no equivalent of the /updates endpoint.
	pendingVersion := priorPending.ValueString()
	switch {
	case pendingVersion != "" && (app.Status == AppStatusInstallFailed || app.Status == AppStatusUpdateFailed):
		// Update failed — clear pending, let Terraform see the actual version and retry.
		state.PendingTargetVersion = types.StringValue("")
		state.PendingOperationID = types.StringValue("")
	case pendingVersion != "" && app.Version == pendingVersion:
		// Update completed successfully — clear pending, use actual version.
		state.PendingTargetVersion = types.StringValue("")
		state.PendingOperationID = types.StringValue("")
	case pendingVersion != "":
		// Update is still in flight (in window queue or running) — suppress drift.
		state.TargetVersion = priorTargetVersion
		state.PendingTargetVersion = priorPending
		state.PendingOperationID = priorPendingOpID
	default:
		// No pending update tracked in state — clear (already cleared by updateModelFromApp).
		state.PendingTargetVersion = types.StringValue("")
		state.PendingOperationID = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *EnvironmentAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state EnvironmentAppResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Updating BC Admin Center environment app", map[string]interface{}{
		"application_family": state.ApplicationFamily.ValueString(),
		"environment_name":   state.EnvironmentName.ValueString(),
		"app_id":             state.AppID.ValueString(),
		"target_version":     plan.TargetVersion.ValueString(),
	})

	svc := NewService(r.client.ForTenant(state.AADTenantID.ValueString()))

	// Option C: if use_environment_update_window changed while an update is pending,
	// cancel the existing scheduled operation first so we can re-submit with the new
	// flag value.  The BC cancel endpoint requires the scheduled operation ID in the
	// request body; use the stored pending_operation_id if available, otherwise look
	// it up from the app operations list.
	cancelFailed := false
	pendingInState := state.PendingOperationID.ValueString() != "" || state.PendingTargetVersion.ValueString() != ""
	if pendingInState && !plan.UseEnvironmentUpdateWindow.Equal(state.UseEnvironmentUpdateWindow) {
		// Obtain the scheduled operation ID.
		scheduledOpID := state.PendingOperationID.ValueString()
		if scheduledOpID == "" {
			var lookupErr error
			scheduledOpID, lookupErr = svc.GetScheduledUpdateOperationID(ctx,
				state.ApplicationFamily.ValueString(),
				state.EnvironmentName.ValueString(),
				state.AppID.ValueString(),
			)
			if lookupErr != nil {
				cancelFailed = true
				tflog.Debug(ctx, "Could not look up scheduled operation ID, skipping cancel", map[string]interface{}{
					"app_id":       state.AppID.ValueString(),
					"lookup_error": lookupErr.Error(),
				})
			}
		}

		if !cancelFailed {
			tflog.Debug(ctx, "Cancelling scheduled app update to reschedule with new window setting", map[string]interface{}{
				"app_id":         state.AppID.ValueString(),
				"operation_id":   scheduledOpID,
				"new_use_window": plan.UseEnvironmentUpdateWindow.ValueBool(),
			})
			cancelErr := svc.CancelUpdate(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString(), state.AppID.ValueString(), scheduledOpID)
			if cancelErr != nil {
				// Log and proceed: the re-submit below may still succeed if BC accepts it.
				// If cancel failed AND the re-submit returns "already scheduled" we know the
				// original operation is still alive with the old window setting — tracked via cancelFailed.
				cancelFailed = true
				tflog.Debug(ctx, "Could not cancel scheduled app update, proceeding with re-submit", map[string]interface{}{
					"app_id":       state.AppID.ValueString(),
					"operation_id": scheduledOpID,
					"cancel_error": cancelErr.Error(),
				})
			}
		}
		// Clear the pending signals regardless — we're about to re-submit.
		state.PendingOperationID = types.StringValue("")
		state.PendingTargetVersion = types.StringValue("")
	}

	updateReq := &UpdateAppRequest{
		AllowPreviewVersion:               plan.AllowPreviewVersion.ValueBool(),
		InstallOrUpdateNeededDependencies: plan.InstallOrUpdateNeededDependencies.ValueBool(),
		UseEnvironmentUpdateWindow:        plan.UseEnvironmentUpdateWindow.ValueBool(),
	}
	if !plan.TargetVersion.IsNull() && !plan.TargetVersion.IsUnknown() && plan.TargetVersion.ValueString() != "" {
		updateReq.TargetVersion = plan.TargetVersion.ValueString()
	}

	operation, err := svc.Update(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString(), state.AppID.ValueString(), updateReq)
	if err != nil {
		if IsAlreadyScheduledError(err) {
			if cancelFailed {
				// The cancel didn't take effect and BC rejected our re-submit because the
				// original operation (with the old use_environment_update_window value) is
				// still scheduled.  Surface a clear error so the user knows the setting
				// was not changed; they can retry once the scheduled update completes.
				resp.Diagnostics.AddError(
					"Could not change update window setting: existing scheduled update could not be cancelled",
					fmt.Sprintf(
						"App %s has a scheduled update that BC rejected the cancellation of, "+
							"and the re-submit was also rejected because the update is already queued. "+
							"The scheduled update retains its original `use_environment_update_window` setting. "+
							"Wait for the scheduled update to complete, then apply again.",
						state.AppID.ValueString()),
				)
				return
			}
			// An update to the same target version is already queued in BC's update
			// window — nothing to do. Treat this as a deferred success so the state
			// is written consistently and the next refresh resolves it.
			tflog.Debug(ctx, "App update already scheduled, treating as deferred success", map[string]interface{}{
				"app_id":         state.AppID.ValueString(),
				"target_version": plan.TargetVersion.ValueString(),
			})
			plan.ID = state.ID
			plan.AADTenantID = state.AADTenantID
			intendedVersion := plan.TargetVersion
			app, readErr := svc.GetByID(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString(), state.AppID.ValueString())
			if readErr != nil {
				resp.Diagnostics.AddError(
					"Error reading app after already-scheduled conflict",
					fmt.Sprintf("Could not read app %s: %s", state.AppID.ValueString(), readErr),
				)
				return
			}
			if app != nil {
				updateModelFromApp(&plan, app)
			} else {
				// App not visible yet (update in flight) — carry over computed fields from state.
				plan.Name = state.Name
				plan.Publisher = state.Publisher
				plan.PublishedAs = state.PublishedAs
				plan.Status = state.Status
			}
			// Preserve the intended target version and mark the update as pending.
			// No operation ID is available for the already-scheduled case.
			plan.TargetVersion = intendedVersion
			plan.PendingTargetVersion = intendedVersion
			plan.PendingOperationID = types.StringValue("")
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}
		resp.Diagnostics.AddError(
			"Error updating app",
			fmt.Sprintf("Could not update app %s: %s", state.AppID.ValueString(), err),
		)
		return
	}

	timeout := 60 * time.Minute

	deferred, err := svc.WaitForOperation(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString(), operation.ID, timeout, plan.UseEnvironmentUpdateWindow.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error waiting for app update",
			fmt.Sprintf("App update failed: %s", err),
		)
		return
	}

	// Preserve the ID and tenant from state.
	plan.ID = state.ID
	plan.AADTenantID = state.AADTenantID

	// Capture the intended target version before updateModelFromApp can overwrite it.
	intendedTargetVersion := plan.TargetVersion

	// Refresh state from API.
	app, err := svc.GetByID(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString(), state.AppID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated app",
			fmt.Sprintf("Could not read app after update: %s", err),
		)
		return
	}
	if app != nil {
		updateModelFromApp(&plan, app)
		if deferred {
			// The update was deferred to the environment's update window. Set
			// pending_target_version as the persistent "selected update" signal so
			// Read can suppress drift on every subsequent refresh until the window runs.
			plan.TargetVersion = intendedTargetVersion
			plan.PendingTargetVersion = intendedTargetVersion
			plan.PendingOperationID = types.StringValue(operation.ID)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *EnvironmentAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentAppResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Uninstalling BC Admin Center environment app", map[string]interface{}{
		"application_family": state.ApplicationFamily.ValueString(),
		"environment_name":   state.EnvironmentName.ValueString(),
		"app_id":             state.AppID.ValueString(),
	})

	svc := NewService(r.client.ForTenant(state.AADTenantID.ValueString()))

	uninstallReq := &UninstallAppRequest{
		DoNotSaveData:              false,
		UninstallDependents:        false,
		UseEnvironmentUpdateWindow: state.UseEnvironmentUpdateWindow.ValueBool(),
	}

	operation, err := svc.Uninstall(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString(), state.AppID.ValueString(), uninstallReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error uninstalling app",
			fmt.Sprintf("Could not uninstall app %s: %s", state.AppID.ValueString(), err),
		)
		return
	}

	// If the app was already gone, the uninstall may have returned an error above,
	// but if the API returns 202 with an operation we still wait for it.
	if operation != nil {
		timeout := 60 * time.Minute
		if _, err := svc.WaitForOperation(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString(), operation.ID, timeout, state.UseEnvironmentUpdateWindow.ValueBool()); err != nil {
			resp.Diagnostics.AddError(
				"Error waiting for app uninstall",
				fmt.Sprintf("App uninstall failed: %s", err),
			)
			return
		}
	}
}

// ImportState handles importing an existing resource.
func (r *EnvironmentAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tenantID, applicationFamily, environmentName, appID, err := ParseEnvironmentAppID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Could not parse environment app resource ID %q: %s\n\n"+
				"Expected format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/apps/{appId}",
				req.ID, err),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aad_tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_family"), applicationFamily)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_name"), environmentName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), appID)...)
}

// updateModelFromApp populates the resource model from an App API response.
// It always updates status, name, publisher, published_as and target_version.
// Callers that need drift suppression (Read, deferred Update) must adjust
// target_version and pending_target_version after this call.
func updateModelFromApp(model *EnvironmentAppResourceModel, app *App) {
	model.Name = types.StringValue(app.Name)
	model.Publisher = types.StringValue(app.Publisher)
	model.PublishedAs = types.StringValue(app.PublishedAs)
	model.Status = types.StringValue(app.Status)
	model.PendingTargetVersion = types.StringValue("")
	model.PendingOperationID = types.StringValue("")
	if app.Version != "" {
		model.TargetVersion = types.StringValue(app.Version)
	}
}
