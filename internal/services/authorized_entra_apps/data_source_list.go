// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package authorized_entra_apps

import (
	"context"
	"fmt"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &authorizedEntraAppsDataSource{}
	_ datasource.DataSourceWithConfigure = &authorizedEntraAppsDataSource{}
)

// NewAuthorizedEntraAppsDataSource creates a new instance of the data source.
func NewAuthorizedEntraAppsDataSource() datasource.DataSource {
	return &authorizedEntraAppsDataSource{}
}

// authorizedEntraAppsDataSource is the data source implementation.
type authorizedEntraAppsDataSource struct {
	client  *client.Client
	service *Service
}

// authorizedEntraAppsDataSourceModel describes the data source data model.
type authorizedEntraAppsDataSourceModel struct {
	Apps []authorizedAppModel `tfsdk:"apps"`
}

type authorizedAppModel struct {
	AppID                 types.String `tfsdk:"app_id"`
	IsAdminConsentGranted types.Bool   `tfsdk:"is_admin_consent_granted"`
}

// Metadata returns the data source type name.
func (d *authorizedEntraAppsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authorized_entra_apps"
}

// Schema defines the schema for the data source.
func (d *authorizedEntraAppsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all Microsoft Entra apps authorized to call the Business Central Admin Center API.",
		Attributes: map[string]schema.Attribute{
			"apps": schema.ListNestedAttribute{
				Description: "List of authorized Microsoft Entra apps.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"app_id": schema.StringAttribute{
							Description: "The application (client) ID of the Microsoft Entra app.",
							Computed:    true,
						},
						"is_admin_consent_granted": schema.BoolAttribute{
							Description: "Indicates whether admin consent has been granted for the app.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *authorizedEntraAppsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	providerClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = providerClient
	d.service = NewService(providerClient)
}

// Read refreshes the Terraform state with the latest data.
func (d *authorizedEntraAppsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state authorizedEntraAppsDataSourceModel

	// Get all authorized apps.
	apps, err := d.service.ListAuthorizedApps(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading authorized Entra apps",
			fmt.Sprintf("Could not list authorized apps: %s", err.Error()),
		)
		return
	}

	// Convert to model.
	state.Apps = make([]authorizedAppModel, len(apps))
	for i, app := range apps {
		state.Apps[i] = authorizedAppModel{
			AppID:                 types.StringValue(app.AppID),
			IsAdminConsentGranted: types.BoolValue(app.IsAdminConsentGranted),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
