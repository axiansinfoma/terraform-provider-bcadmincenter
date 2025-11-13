// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &environmentsDataSource{}
	_ datasource.DataSourceWithConfigure = &environmentsDataSource{}
)

// NewEnvironmentsDataSource is a helper function to simplify the provider implementation.
func NewEnvironmentsDataSource() datasource.DataSource {
	return &environmentsDataSource{}
}

// environmentsDataSource is the data source implementation.
type environmentsDataSource struct {
	client *client.Client
}

// environmentsDataSourceModel maps the data source schema data.
type environmentsDataSourceModel struct {
	ApplicationFamily types.String              `tfsdk:"application_family"`
	Environments      []environmentListItemModel `tfsdk:"environments"`
}

// environmentListItemModel represents a single environment in the list.
type environmentListItemModel struct {
	Name               types.String `tfsdk:"name"`
	Type               types.String `tfsdk:"type"`
	CountryCode        types.String `tfsdk:"country_code"`
	RingName           types.String `tfsdk:"ring_name"`
	ApplicationVersion types.String `tfsdk:"application_version"`
	Status             types.String `tfsdk:"status"`
	WebClientLoginURL  types.String `tfsdk:"web_client_login_url"`
	AadTenantID        types.String `tfsdk:"aad_tenant_id"`
}

// Metadata returns the data source type name.
func (d *environmentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

// Schema defines the schema for the data source.
func (d *environmentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of all Business Central environments for a given application family.",
		Attributes: map[string]schema.Attribute{
			"application_family": schema.StringAttribute{
				Description: "The application family (e.g., 'BusinessCentral').",
				Required:    true,
			},
			"environments": schema.ListNestedAttribute{
				Description: "List of environments.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "The name of the environment.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The type of environment (Production or Sandbox).",
							Computed:    true,
						},
						"country_code": schema.StringAttribute{
							Description: "The country code for the environment.",
							Computed:    true,
						},
						"ring_name": schema.StringAttribute{
							Description: "The ring name for the environment.",
							Computed:    true,
						},
						"application_version": schema.StringAttribute{
							Description: "The application version of the environment.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The current status of the environment.",
							Computed:    true,
						},
						"web_client_login_url": schema.StringAttribute{
							Description: "The web client login URL for the environment.",
							Computed:    true,
						},
						"aad_tenant_id": schema.StringAttribute{
							Description: "The Azure AD tenant ID associated with the environment.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *environmentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *environmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state environmentsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get environments from API
	service := NewService(d.client)
	environments, err := service.List(ctx, state.ApplicationFamily.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Environments",
			fmt.Sprintf("An error occurred while reading the environments: %s", err.Error()),
		)
		return
	}

	// Map response to model
	state.Environments = make([]environmentListItemModel, len(environments))
	for i, env := range environments {
		state.Environments[i] = environmentListItemModel{
			Name:               types.StringValue(env.Name),
			Type:               types.StringValue(env.Type),
			CountryCode:        types.StringValue(env.CountryCode),
			RingName:           types.StringValue(env.RingName),
			ApplicationVersion: types.StringValue(env.ApplicationVersion),
			Status:             types.StringValue(env.Status),
			WebClientLoginURL:  types.StringValue(env.WebClientLoginURL),
			AadTenantID:        types.StringValue(env.AADTenantID),
		}
	}

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
