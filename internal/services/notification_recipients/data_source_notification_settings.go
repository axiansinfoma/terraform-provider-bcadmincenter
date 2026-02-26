// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package notificationrecipients

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &NotificationSettingsDataSource{}
var _ datasource.DataSourceWithConfigure = &NotificationSettingsDataSource{}

// NewNotificationSettingsDataSource is a helper function to simplify the provider implementation.
func NewNotificationSettingsDataSource() datasource.DataSource {
	return &NotificationSettingsDataSource{}
}

// NotificationSettingsDataSource is the data source implementation.
type NotificationSettingsDataSource struct {
	client *client.Client
}

// Metadata returns the data source type name.
func (d *NotificationSettingsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_settings"
}

// Schema defines the schema for the data source.
func (d *NotificationSettingsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the complete notification settings for the Business Central tenant, " +
			"including the Azure AD tenant ID and all configured notification recipients. " +
			"This data source provides read-only access to the tenant-wide notification configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Terraform identifier (set to the AAD tenant ID)",
				Computed:    true,
			},
			"aad_tenant_id": schema.StringAttribute{
				Description: "The Azure Active Directory tenant ID. If not specified, defaults to the provider's configured tenant ID. This value is also returned by the API.",
				Optional:    true,
				Computed:    true,
			},
			"recipients": schema.ListNestedAttribute{
				Description: "List of notification recipients configured for the tenant",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the notification recipient",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The email address of the notification recipient",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The full name of the notification recipient",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NotificationSettingsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data.
func (d *NotificationSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NotificationSettingsDataSourceModel

	// Read configuration.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Determine tenant ID to use.
	tenantID := d.client.GetTenantID()
	if !config.AADTenantID.IsNull() && !config.AADTenantID.IsUnknown() {
		tenantID = config.AADTenantID.ValueString()
	}

	// Get notification settings from API.
	svc := NewService(d.client)
	settings, err := svc.GetNotificationSettings(ctx, tenantID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Notification Settings",
			"Could not read notification settings: "+err.Error(),
		)
		return
	}

	// Map response to state.
	config.ID = types.StringValue(settings.AADTenantID)
	config.AADTenantID = types.StringValue(settings.AADTenantID)

	// Map recipients.
	recipients := make([]NotificationRecipientDataSourceModel, len(settings.Recipients))
	for i, recipient := range settings.Recipients {
		recipients[i] = NotificationRecipientDataSourceModel{
			ID:    types.StringValue(recipient.ID),
			Email: types.StringValue(recipient.Email),
			Name:  types.StringValue(recipient.Name),
		}
	}
	config.Recipients = recipients

	// Save data into Terraform state.
	diags := resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}
