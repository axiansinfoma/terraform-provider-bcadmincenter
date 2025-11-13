// Copyright (c) 2025 Michael Villani
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
	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &NotificationRecipientResource{}
var _ resource.ResourceWithConfigure = &NotificationRecipientResource{}
var _ resource.ResourceWithImportState = &NotificationRecipientResource{}

// NewNotificationRecipientResource is a helper function to simplify the provider implementation
func NewNotificationRecipientResource() resource.Resource {
	return &NotificationRecipientResource{}
}

// NotificationRecipientResource is the resource implementation
type NotificationRecipientResource struct {
	client *client.Client
}

// Metadata returns the resource type name
func (r *NotificationRecipientResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_recipient"
}

// Schema defines the schema for the resource
func (r *NotificationRecipientResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a notification recipient for the Business Central tenant. " +
			"Notification recipients receive emails about environment lifecycle events such as " +
			"update availability, successful updates, update failures, and extension validations. " +
			"Up to 100 notification recipients can be configured per tenant.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the notification recipient (assigned by the API)",
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
		},
	}
}

// Configure adds the provider configured client to the resource
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

// Create creates the resource and sets the initial Terraform state
func (r *NotificationRecipientResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NotificationRecipientResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the notification recipient
	svc := NewService(r.client)
	recipient, err := svc.Create(ctx, plan.Email.ValueString(), plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Notification Recipient",
			"Could not create notification recipient: "+err.Error(),
		)
		return
	}

	// Update the plan with the response
	plan.ID = types.StringValue(recipient.ID)
	plan.Email = types.StringValue(recipient.Email)
	plan.Name = types.StringValue(recipient.Name)

	// Save data into Terraform state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data
func (r *NotificationRecipientResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NotificationRecipientResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	svc := NewService(r.client)
	recipient, err := svc.Get(ctx, state.ID.ValueString())
	if err != nil {
		// If the recipient is not found, remove it from state
		resp.Diagnostics.AddWarning(
			"Notification Recipient Not Found",
			fmt.Sprintf("The notification recipient with ID %s was not found and will be removed from state.", state.ID.ValueString()),
		)
		resp.State.RemoveResource(ctx)
		return
	}

	// Update the state
	state.Email = types.StringValue(recipient.Email)
	state.Name = types.StringValue(recipient.Name)

	// Save updated data into Terraform state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success
func (r *NotificationRecipientResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// The API does not support updating notification recipients
	// Since email and name both have RequiresReplace, this should never be called
	// However, we implement it for completeness
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"The Business Central Admin Center API does not support updating notification recipients. "+
			"To change a recipient's details, delete and recreate the resource.",
	)
}

// Delete deletes the resource and removes the Terraform state on success
func (r *NotificationRecipientResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NotificationRecipientResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete the notification recipient
	svc := NewService(r.client)
	err := svc.Delete(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Notification Recipient",
			"Could not delete notification recipient: "+err.Error(),
		)
		return
	}

	// Resource is removed from state automatically
}

// ImportState imports the resource state
func (r *NotificationRecipientResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import using the recipient ID
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
