// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ resource.Resource                = &EnvironmentResource{}
	_ resource.ResourceWithConfigure   = &EnvironmentResource{}
	_ resource.ResourceWithImportState = &EnvironmentResource{}
)

// NewEnvironmentResource is a helper function to simplify the provider implementation
func NewEnvironmentResource() resource.Resource {
	return &EnvironmentResource{}
}

// EnvironmentResource is the resource implementation
type EnvironmentResource struct {
	client *client.Client
}

// EnvironmentResourceModel describes the resource data model
type EnvironmentResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	ApplicationFamily  types.String `tfsdk:"application_family"`
	Type               types.String `tfsdk:"type"`
	CountryCode        types.String `tfsdk:"country_code"`
	RingName           types.String `tfsdk:"ring_name"`
	ApplicationVersion types.String `tfsdk:"application_version"`
	AzureRegion        types.String `tfsdk:"azure_region"`
	Status             types.String `tfsdk:"status"`
	WebClientLoginURL  types.String `tfsdk:"web_client_login_url"`
	WebServiceURL      types.String `tfsdk:"web_service_url"`
	AppInsightsKey     types.String `tfsdk:"app_insights_key"`
	PlatformVersion    types.String `tfsdk:"platform_version"`
	AADTenantID        types.String `tfsdk:"aad_tenant_id"`
	Timeouts           types.Object `tfsdk:"timeouts"`
}

// Metadata returns the resource type name
func (r *EnvironmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

// Schema defines the schema for the resource
func (r *EnvironmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Business Central environment in the Admin Center.\n\n" +
			"This resource creates and manages Business Central environments (Production or Sandbox). " +
			"Environment creation is an asynchronous operation that can take several minutes to complete.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier for the environment (same as name).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the environment. Must be between 1 and 30 characters. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 30),
				},
			},
			"application_family": schema.StringAttribute{
				MarkdownDescription: "The application family for the environment. Defaults to 'BusinessCentral'. Cannot be changed after creation.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("BusinessCentral"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("BusinessCentral"),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of environment. Must be either 'Production' or 'Sandbox'. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("Production", "Sandbox"),
				},
			},
			"country_code": schema.StringAttribute{
				MarkdownDescription: "The country/region code for the environment (e.g., 'US', 'GB', 'DK'). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ring_name": schema.StringAttribute{
				MarkdownDescription: "The release ring for the environment. Must be one of 'PROD', 'PREVIEW', or 'FAST'. Defaults to 'PROD'. Cannot be changed after creation.",
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
				MarkdownDescription: "The specific application version to use. If not specified, the latest version for the ring will be used. Cannot be changed after creation.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"azure_region": schema.StringAttribute{
				MarkdownDescription: "The Azure region where the environment should be created. If not specified, a default region will be used. Cannot be changed after creation.",
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
				MarkdownDescription: "The Azure AD tenant ID for the environment.",
				Computed:            true,
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

// Configure adds the provider configured client to the resource
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

// Create creates the resource and sets the initial Terraform state
func (r *EnvironmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating BC Admin Center environment", map[string]interface{}{
		"name":               plan.Name.ValueString(),
		"application_family": plan.ApplicationFamily.ValueString(),
		"type":               plan.Type.ValueString(),
	})

	// Create environment service
	svc := NewService(r.client)

	// Prepare create request
	createReq := &CreateEnvironmentRequest{
		EnvironmentType:    plan.Type.ValueString(),
		Name:               plan.Name.ValueString(),
		CountryCode:        plan.CountryCode.ValueString(),
		RingName:           plan.RingName.ValueString(), // API expects "PROD", "PREVIEW", "FAST"
		ApplicationVersion: plan.ApplicationVersion.ValueString(),
		AzureRegion:        plan.AzureRegion.ValueString(),
	}

	// Create the environment
	operation, err := svc.Create(ctx, plan.ApplicationFamily.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating environment",
			fmt.Sprintf("Could not create environment: %s", err),
		)
		return
	}

	// Log the operation response for debugging
	tflog.Debug(ctx, "Create operation response", map[string]interface{}{
		"operation_id":       operation.ID,
		"operation_type":     operation.Type,
		"product_family":     operation.ProductFamily,
		"application_family": operation.ApplicationFamily,
		"environment_name":   operation.EnvironmentName,
		"destination_env":    operation.DestinationEnvironment,
		"source_env":         operation.SourceEnvironment,
	})

	// Determine timeout
	timeout := 60 * time.Minute // default
	if !plan.Timeouts.IsNull() {
		// TODO: Parse timeout from plan.Timeouts
	}

	// Wait for the operation to complete
	// Use ProductFamily from operation response if available, otherwise use the plan value
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

	// Log what we're about to use for the Get call
	tflog.Debug(ctx, "Reading created environment", map[string]interface{}{
		"application_family": appFamily,
		"environment_name":   envName,
	})

	// Wait for the environment to become Active
	// The operation succeeds when the create request is accepted, but the environment
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
			// Environment is ready, update state and return
			r.updateModelFromEnvironment(&plan, env)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}

		// Check for failed states
		if env.Status == "Failed" || env.Status == "Suspended" {
			resp.Diagnostics.AddError(
				"Environment creation failed",
				fmt.Sprintf("Environment entered %s state during creation", env.Status),
			)
			return
		}

		// Wait for next tick or timeout
		select {
		case <-envTimeout.Done():
			resp.Diagnostics.AddError(
				"Timeout waiting for environment",
				fmt.Sprintf("Environment did not become Active within %v (current status: %s)", timeout, env.Status),
			)
			return
		case <-ticker.C:
			// Continue polling
			continue
		}
	}
}

// Read refreshes the Terraform state with the latest data
func (r *EnvironmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading BC Admin Center environment", map[string]interface{}{
		"name":               state.Name.ValueString(),
		"application_family": state.ApplicationFamily.ValueString(),
	})

	// Create environment service
	svc := NewService(r.client)

	// Get the environment
	env, err := svc.Get(ctx, state.ApplicationFamily.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading environment",
			fmt.Sprintf("Could not read environment: %s", err),
		)
		return
	}

	// Update state with current environment data
	r.updateModelFromEnvironment(&state, env)

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *EnvironmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Most environment attributes require replacement (ForceNew)
	// This method is here for completeness but may not be used in practice
	resp.Diagnostics.AddError(
		"Update not supported",
		"Environment resources do not support in-place updates. Most changes require replacement.",
	)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *EnvironmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentResourceModel

	// Read current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting BC Admin Center environment", map[string]interface{}{
		"name":               state.Name.ValueString(),
		"application_family": state.ApplicationFamily.ValueString(),
	})

	// Create environment service
	svc := NewService(r.client)

	// Delete the environment
	operation, err := svc.Delete(ctx, state.ApplicationFamily.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting environment",
			fmt.Sprintf("Could not delete environment: %s", err),
		)
		return
	}

	// If operation is nil, the environment was already deleted
	if operation == nil {
		return
	}

	// Determine timeout
	timeout := 60 * time.Minute // default
	if !state.Timeouts.IsNull() {
		// TODO: Parse timeout from state.Timeouts
	}

	// Wait for the operation to complete
	// Use ProductFamily from operation response if available, otherwise fall back to state
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

// ImportState imports an existing resource into Terraform state
func (r *EnvironmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by ID in format: tenantId/applicationFamily/environmentName
	// Example: 9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/BusinessCentral/test-sandbox

	// Split the import ID
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'tenantId/applicationFamily/environmentName', got: %s\n"+
				"Example: 9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/BusinessCentral/test-sandbox", req.ID),
		)
		return
	}

	// We don't store tenant ID in the resource state, but we validate it matches the provider configuration
	// tenantId := parts[0]
	applicationFamily := parts[1]
	environmentName := parts[2]

	// Set the attributes
	resp.State.SetAttribute(ctx, path.Root("application_family"), applicationFamily)
	resp.State.SetAttribute(ctx, path.Root("name"), environmentName)
	resp.State.SetAttribute(ctx, path.Root("id"), environmentName)
}

// updateModelFromEnvironment updates the Terraform model with data from the API
func (r *EnvironmentResource) updateModelFromEnvironment(model *EnvironmentResourceModel, env *Environment) {
	model.ID = types.StringValue(env.Name)
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

	// Azure region is not returned by the API, so always set to null
	model.AzureRegion = types.StringNull()

	// Normalize ring name from API response format to Terraform format
	// API accepts "PROD", "PREVIEW", "FAST" on input but returns "Production", "Preview", "Fast" on output
	if env.RingName != "" {
		normalizedRing := normalizeRingName(env.RingName)
		model.RingName = types.StringValue(normalizedRing)
	} else {
		model.RingName = types.StringNull()
	}

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

// normalizeRingName converts API ring name format to Terraform format
// API returns "Production", "Preview", "Fast" but Terraform expects "PROD", "PREVIEW", "FAST"
func normalizeRingName(apiRingName string) string {
	switch apiRingName {
	case "Production":
		return "PROD"
	case "Preview":
		return "PREVIEW"
	case "Fast":
		return "FAST"
	default:
		// Return as-is if unknown
		return apiRingName
	}
}
