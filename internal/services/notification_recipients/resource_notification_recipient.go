// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package notificationrecipients

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NotificationRecipientResource{}
var _ resource.ResourceWithConfigure = &NotificationRecipientResource{}
var _ resource.ResourceWithImportState = &NotificationRecipientResource{}

// NewNotificationRecipientResource is a helper function to simplify the provider implementation.
func NewNotificationRecipientResource() resource.Resource {
	return &NotificationRecipientResource{}
}

// NotificationRecipientResource is the resource implementation.
type NotificationRecipientResource struct {
	client *client.Client
}

// Metadata returns the resource type name.
func (r *NotificationRecipientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_recipient"
}

// Schema defines the schema for the resource.
func (r *NotificationRecipientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a notification recipient for the Business Central tenant. " +
			"Notification recipients receive emails about environment lifecycle events such as " +
			"update availability, successful updates, update failures, and extension validations. " +
			"Up to 100 notification recipients can be configured per tenant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ARM-like resource ID (format: /tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/{recipientId})",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"email": schema.StringAttribute{
				Description: "The email address of the notification recipient. Must be a valid email address.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The full name of the notification recipient",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"aad_tenant_id": schema.StringAttribute{
				Description: "The Azure AD tenant ID. If not specified, defaults to the provider's configured tenant ID. This allows managing notification recipients in different tenants.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *NotificationRecipientResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *NotificationRecipientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NotificationRecipientResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the notification recipient.
	svc := NewService(r.client)

	// Use aad_tenant_id from plan if provided, otherwise use provider's tenant ID.
	tenantID := r.client.GetTenantID()
	if !plan.AADTenantID.IsNull() && !plan.AADTenantID.IsUnknown() {
		tenantID = plan.AADTenantID.ValueString()
	}

	recipient, err := svc.Create(ctx, tenantID, plan.Email.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Notification Recipient",
			"Could not create notification recipient: "+err.Error(),
		)
		return
	}

	// Update the plan with the response.
	plan.ID = types.StringValue(BuildNotificationRecipientID(tenantID, recipient.ID))
	plan.Email = types.StringValue(recipient.Email)
	plan.Name = types.StringValue(recipient.Name)
	plan.AADTenantID = types.StringValue(tenantID) // Save data into Terraform state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *NotificationRecipientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NotificationRecipientResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc := NewService(r.client)

	// Parse the ARM-like ID to get tenant ID and recipient ID.
	tenantID, recipientID, err := ParseNotificationRecipientID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Could not parse notification recipient ID: %s", err.Error()),
		)
		return
	}

	recipient, err := svc.Get(ctx, tenantID, recipientID)
	if err != nil {
		// If the recipient is not found, remove it from state.
		resp.Diagnostics.AddWarning(
			"Notification Recipient Not Found",
			fmt.Sprintf("The notification recipient with ID %s was not found and will be removed from state.", state.ID.ValueString()),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state.
	state.Email = types.StringValue(recipient.Email)
	state.Name = types.StringValue(recipient.Name)
	state.AADTenantID = types.StringValue(tenantID)

	// Save updated data into Terraform state.
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *NotificationRecipientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// The API does not support updating notification recipients.
	// Since email and name both have RequiresReplace, this should never be called.
	// However, we implement it for completeness.
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"The Business Central Admin Center API does not support updating notification recipients. "+
			"To change a recipient's details, delete and recreate the resource.",
	)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *NotificationRecipientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NotificationRecipientResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the notification recipient.
	svc := NewService(r.client)

	// Parse the ARM-like ID to get tenant ID and recipient ID.
	tenantID, recipientID, err := ParseNotificationRecipientID(state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Could not parse notification recipient ID: %s", err.Error()),
		)
		return
	}
	err = svc.Delete(ctx, tenantID, recipientID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Notification Recipient",
			"Could not delete notification recipient: "+err.Error(),
		)
		return
	}

	// Resource is removed from state automatically.
}

// ImportState imports the resource state.
func (r *NotificationRecipientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Parse the ARM-like ID.
	tenantID, recipientID, err := ParseNotificationRecipientID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected ARM-like resource ID in format '/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/{recipientId}', got: %s\nError: %s",
				req.ID, err.Error()),
		)
		return
	}

	// Set the ID and tenant ID in state.
	resp.State.SetAttribute(ctx, path.Root("id"), req.ID)
	resp.State.SetAttribute(ctx, path.Root("aad_tenant_id"), tenantID)

	// Note: The Read method will populate email and name.
	_ = recipientID // Used by Read method via ID parsing
}
