// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package authorized_entra_apps

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

var (
	_ resource.Resource                = &AuthorizedEntraAppResource{}
	_ resource.ResourceWithConfigure   = &AuthorizedEntraAppResource{}
	_ resource.ResourceWithImportState = &AuthorizedEntraAppResource{}
)

// NewAuthorizedEntraAppResource creates a new instance of the authorized Entra app resource
func NewAuthorizedEntraAppResource() resource.Resource {
	return &AuthorizedEntraAppResource{}
}

// AuthorizedEntraAppResource defines the resource implementation
type AuthorizedEntraAppResource struct {
	client  *client.Client
	service *Service
}

// authorizedEntraAppResourceModel describes the resource data model
type authorizedEntraAppResourceModel struct {
	ID                    types.String `tfsdk:"id"`
	AADTenantID           types.String `tfsdk:"aad_tenant_id"`
	AppID                 types.String `tfsdk:"app_id"`
	IsAdminConsentGranted types.Bool   `tfsdk:"is_admin_consent_granted"`
}

// Metadata returns the resource type name
func (r *AuthorizedEntraAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authorized_entra_app"
}

// Schema defines the schema for the resource
func (r *AuthorizedEntraAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages authorization of a Microsoft Entra app to call the Business Central Admin Center API. " +
			"Note: This does not grant admin consent or assign permission sets in environments.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ARM-like resource ID (format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/authorizedEntraApps/{appId})",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aad_tenant_id": schema.StringAttribute{
				Description: "The Azure AD tenant ID. Defaults to the provider's configured tenant ID.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"app_id": schema.StringAttribute{
				Description: "The application (client) ID of the Microsoft Entra app to authorize.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_admin_consent_granted": schema.BoolAttribute{
				Description: "Indicates whether admin consent has been granted for the app. This is read-only and managed outside Terraform.",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *AuthorizedEntraAppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = providerClient
	r.service = NewService(providerClient)
}

// Create creates the resource and sets the initial Terraform state
func (r *AuthorizedEntraAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan authorizedEntraAppResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use provider tenant ID if not specified
	tenantID := plan.AADTenantID.ValueString()
	if tenantID == "" {
		tenantID = r.client.GetTenantID()
		plan.AADTenantID = types.StringValue(tenantID)
	}

	// Authorize the app
	app, err := r.service.AuthorizeApp(ctx, plan.AppID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error authorizing Entra app",
			fmt.Sprintf("Could not authorize app %s: %s", plan.AppID.ValueString(), err.Error()),
		)
		return
	}

	// Set resource ID
	plan.ID = types.StringValue(BuildAuthorizedEntraAppID(tenantID, app.AppID))
	plan.IsAdminConsentGranted = types.BoolValue(app.IsAdminConsentGranted)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data
func (r *AuthorizedEntraAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state authorizedEntraAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state from API by listing all apps and filtering
	// Note: The API doesn't provide a GET endpoint for a single app
	apps, err := r.service.ListAuthorizedApps(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading authorized Entra app",
			fmt.Sprintf("Could not list authorized apps: %s", err.Error()),
		)
		return
	}

	// Find the specific app in the list
	var found bool
	for _, app := range apps {
		if app.AppID == state.AppID.ValueString() {
			// Update state
			state.IsAdminConsentGranted = types.BoolValue(app.IsAdminConsentGranted)
			found = true
			break
		}
	}

	// If app is not found, remove from state
	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *AuthorizedEntraAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// This resource doesn't support updates - all attributes are either computed or require replacement
	resp.Diagnostics.AddError(
		"Update not supported",
		"Authorized Entra App resource does not support updates. All changes require replacement.",
	)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *AuthorizedEntraAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state authorizedEntraAppResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Remove the authorized app
	err := r.service.RemoveAuthorizedApp(ctx, state.AppID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing authorized Entra app",
			fmt.Sprintf("Could not remove app %s: %s", state.AppID.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource into Terraform state
func (r *AuthorizedEntraAppResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the resource ID
	tenantID, appID, err := ParseAuthorizedEntraAppID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected ARM-like resource ID in format '/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/authorizedEntraApps/{appId}', got: %s\nError: %s",
				req.ID, err.Error()),
		)
		return
	}

	// Set the parsed values
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aad_tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), appID)...)
}
