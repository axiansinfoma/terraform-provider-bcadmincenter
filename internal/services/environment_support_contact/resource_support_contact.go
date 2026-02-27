// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentsupportcontact

import (
	"context"
	"fmt"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &EnvironmentSupportContactResource{}
var _ resource.ResourceWithConfigure = &EnvironmentSupportContactResource{}
var _ resource.ResourceWithImportState = &EnvironmentSupportContactResource{}

// NewEnvironmentSupportContactResource is a helper function to simplify the provider implementation.
func NewEnvironmentSupportContactResource() resource.Resource {
	return &EnvironmentSupportContactResource{}
}

// EnvironmentSupportContactResource is the resource implementation.
type EnvironmentSupportContactResource struct {
	client *client.Client
}

// EnvironmentSupportContactResourceModel maps the resource schema data.
type EnvironmentSupportContactResourceModel struct {
	ID                types.String `tfsdk:"id"`
	AADTenantID       types.String `tfsdk:"aad_tenant_id"`
	ApplicationFamily types.String `tfsdk:"application_family"`
	EnvironmentName   types.String `tfsdk:"environment_name"`
	Name              types.String `tfsdk:"name"`
	Email             types.String `tfsdk:"email"`
	URL               types.String `tfsdk:"url"`
}

// Metadata returns the resource type name.
func (r *EnvironmentSupportContactResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_support_contact"
}

// Schema defines the schema for the resource.
func (r *EnvironmentSupportContactResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the support contact information for a Business Central environment. This is the contact information displayed to users in the Help and Support page.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ARM-like resource ID (format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/supportContact)",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aad_tenant_id": schema.StringAttribute{
				Description: "The Azure AD tenant ID. If not specified, defaults to the provider's configured tenant ID.",
				Optional:    true,
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
			"name": schema.StringAttribute{
				Description: "The name of the support contact (displayed to users)",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email address of the support contact",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"url": schema.StringAttribute{
				Description: "A URL for additional support information such as a support website or portal",
				Optional:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *EnvironmentSupportContactResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *EnvironmentSupportContactResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EnvironmentSupportContactResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set the ID to the ARM-like format.
	tenantID := r.client.GetTenantID()
	if !plan.AADTenantID.IsNull() && !plan.AADTenantID.IsUnknown() {
		tenantID = plan.AADTenantID.ValueString()
	}
	plan.AADTenantID = types.StringValue(tenantID)
	plan.ID = types.StringValue(BuildEnvironmentSupportContactID(
		tenantID,
		plan.ApplicationFamily.ValueString(),
		plan.EnvironmentName.ValueString(),
	))

	// Create the support contact using the tenant-specific client.
	svc := NewService(r.client.ForTenant(tenantID))
	contact := &SupportContact{
		Name:  plan.Name.ValueString(),
		Email: plan.Email.ValueString(),
		URL:   plan.URL.ValueString(),
	}

	updatedContact, err := svc.Set(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), contact)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Support Contact",
			"Could not create support contact: "+err.Error(),
		)
		return
	}

	// Update the plan with the response.
	plan.Name = types.StringValue(updatedContact.Name)
	plan.Email = types.StringValue(updatedContact.Email)
	plan.URL = types.StringValue(updatedContact.URL)

	// Save data into Terraform state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *EnvironmentSupportContactResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EnvironmentSupportContactResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc := NewService(r.client.ForTenant(state.AADTenantID.ValueString()))
	contact, err := svc.Get(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Support Contact",
			"Could not read support contact: "+err.Error(),
		)
		return
	}

	// If contact is nil, it means it was deleted outside Terraform.
	if contact == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state.
	state.Name = types.StringValue(contact.Name)
	state.Email = types.StringValue(contact.Email)
	state.URL = types.StringValue(contact.URL)

	// Save updated data into Terraform state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *EnvironmentSupportContactResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EnvironmentSupportContactResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state EnvironmentSupportContactResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the support contact using the tenant-specific client.
	svc := NewService(r.client.ForTenant(state.AADTenantID.ValueString()))
	contact := &SupportContact{
		Name:  plan.Name.ValueString(),
		Email: plan.Email.ValueString(),
		URL:   plan.URL.ValueString(),
	}

	updatedContact, err := svc.Set(ctx, plan.ApplicationFamily.ValueString(), plan.EnvironmentName.ValueString(), contact)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Support Contact",
			"Could not update support contact: "+err.Error(),
		)
		return
	}

	// Update the plan with the response.
	plan.Name = types.StringValue(updatedContact.Name)
	plan.Email = types.StringValue(updatedContact.Email)
	plan.URL = types.StringValue(updatedContact.URL)

	// Save updated data into Terraform state.
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *EnvironmentSupportContactResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EnvironmentSupportContactResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The API doesn't have a DELETE endpoint for support contact.
	// We can either:.
	// 1. Set empty values (which might fail validation)
	// 2. Just remove from state with a warning.

	resp.Diagnostics.AddWarning(
		"Support Contact Not Cleared",
		"The Business Central Admin Center API does not provide a DELETE endpoint for support contacts. "+
			"The support contact information will remain configured on the environment but will be removed from Terraform state. "+
			"To clear the support contact, manually update it in the Business Central Admin Center or set it to different values.",
	)

	// Resource is removed from state automatically.
}

// ImportState imports the resource state.
func (r *EnvironmentSupportContactResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the ARM-like ID.
	tenantID, applicationFamily, environmentName, err := ParseEnvironmentSupportContactID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected ARM-like resource ID in format '/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/supportContact', got: %s\nError: %s",
				req.ID, err.Error()),
		)
		return
	}

	// Set the attributes.
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
	resp.State.SetAttribute(ctx, path.Root("aad_tenant_id"), tenantID)
	resp.State.SetAttribute(ctx, path.Root("application_family"), applicationFamily)
	resp.State.SetAttribute(ctx, path.Root("environment_name"), environmentName)
}
