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
	_ datasource.DataSource              = &environmentDataSource{}
	_ datasource.DataSourceWithConfigure = &environmentDataSource{}
)

// NewEnvironmentDataSource is a helper function to simplify the provider implementation.
func NewEnvironmentDataSource() datasource.DataSource {
	return &environmentDataSource{}
}

// environmentDataSource is the data source implementation.
type environmentDataSource struct {
	client *client.Client
}

// environmentDataSourceModel maps the data source schema data.
type environmentDataSourceModel struct {
	ApplicationFamily  types.String `tfsdk:"application_family"`
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
func (d *environmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment"
}

// Schema defines the schema for the data source.
func (d *environmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a specific Business Central environment.",
		Attributes: map[string]schema.Attribute{
			"application_family": schema.StringAttribute{
				Description: "The application family (e.g., 'BusinessCentral').",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the environment.",
				Required:    true,
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *environmentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *environmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state environmentDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get environment from API
	service := NewService(d.client)
	env, err := service.Get(ctx, state.ApplicationFamily.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Environment",
			fmt.Sprintf("An error occurred while reading the environment: %s", err.Error()),
		)
		return
	}

	// Map response to model
	state.Type = types.StringValue(env.Type)
	state.CountryCode = types.StringValue(env.CountryCode)
	state.RingName = types.StringValue(env.RingName)
	state.ApplicationVersion = types.StringValue(env.ApplicationVersion)
	state.Status = types.StringValue(env.Status)
	state.WebClientLoginURL = types.StringValue(env.WebClientLoginURL)
	state.AadTenantID = types.StringValue(env.AADTenantID)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
