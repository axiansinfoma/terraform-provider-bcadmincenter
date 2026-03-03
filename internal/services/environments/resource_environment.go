// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"fmt"
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
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	ApplicationFamily  types.String `tfsdk:"application_family"`
	Type               types.String `tfsdk:"type"`
	CountryCode        types.String `tfsdk:"country_code"`
	RingName           types.String `tfsdk:"ring_name"`
	ApplicationVersion types.String `tfsdk:"application_version"`
	IgnoreUpdateWindow types.Bool   `tfsdk:"ignore_update_window"`
	AzureRegion        types.String `tfsdk:"azure_region"`
	Status             types.String `tfsdk:"status"`
	WebClientLoginURL  types.String `tfsdk:"web_client_login_url"`
	WebServiceURL      types.String `tfsdk:"web_service_url"`
	AppInsightsKey     types.String `tfsdk:"app_insights_key"`
	PlatformVersion    types.String `tfsdk:"platform_version"`
	AADTenantID        types.String `tfsdk:"aad_tenant_id"`
	Timeouts           types.Object `tfsdk:"timeouts"`
}

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
			},
			"web_client_login_url": schema.StringAttribute{
				MarkdownDescription: "The URL for accessing the web client.",
				Computed:            true,
			},
			"web_service_url": schema.StringAttribute{
				MarkdownDescription: "The URL for web service access.",
				Computed:            true,
			},
			"app_insights_key": schema.StringAttribute{
				MarkdownDescription: "The Application Insights instrumentation key for the environment.",
				Computed:            true,
				Sensitive:           true,
			},
			"platform_version": schema.StringAttribute{
				MarkdownDescription: "The platform version of the environment.",
				Computed:            true,
			},
			"aad_tenant_id": schema.StringAttribute{
				MarkdownDescription: "The Azure AD tenant ID for the environment. If not specified, the value is read from the API response.",
				Optional:            true,
				Computed:            true,
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

	if !versionChanged && !windowChanged {
		// Nothing to do; copy plan to state.
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
}

// applyUpdatesDriftDetection applies drift detection logic based on the environment updates list.
// It updates model.ApplicationVersion according to the three-condition table in the issue.
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
		return
	}

	switch selectedUpdate.UpdateStatus {
	case UpdateStatusScheduled, UpdateStatusRunning:
		// Suppress drift: store the target version so Terraform sees no change.
		model.ApplicationVersion = types.StringValue(selectedUpdate.TargetVersion)
	case UpdateStatusFailed:
		// Drift: store the currently running version so Terraform detects a change and retries.
		if env.ApplicationVersion != "" {
			model.ApplicationVersion = types.StringValue(env.ApplicationVersion)
		} else {
			model.ApplicationVersion = types.StringNull()
		}
	default:
		// For other statuses (e.g., "succeeded"), fall through to the environment version (already set).
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
