// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package authorized_entra_apps

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

var (
	_ datasource.DataSource              = &manageableTenantsDataSource{}
	_ datasource.DataSourceWithConfigure = &manageableTenantsDataSource{}
)

// NewManageableTenantsDataSource creates a new instance of the data source
func NewManageableTenantsDataSource() datasource.DataSource {
	return &manageableTenantsDataSource{}
}

// manageableTenantsDataSource is the data source implementation
type manageableTenantsDataSource struct {
	client  *client.Client
	service *Service
}

// manageableTenantsDataSourceModel describes the data source data model
type manageableTenantsDataSourceModel struct {
	Tenants []manageableTenantModel `tfsdk:"tenants"`
}

// manageableTenantModel describes a single manageable tenant
type manageableTenantModel struct {
	EntraTenantID types.String `tfsdk:"entra_tenant_id"`
}

// Metadata returns the data source type name
func (d *manageableTenantsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manageable_tenants"
}

// Schema defines the schema for the data source
func (d *manageableTenantsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a list of Microsoft Entra tenants that the authenticating app can manage. Note: This data source can only be used when authenticated as an app (service principal), not with user authentication.",
		Attributes: map[string]schema.Attribute{
			"tenants": schema.ListNestedAttribute{
				Description: "List of manageable tenants for the authenticating app.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"entra_tenant_id": schema.StringAttribute{
							Description: "The Microsoft Entra tenant ID where the app is authorized.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *manageableTenantsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data
func (d *manageableTenantsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state manageableTenantsDataSourceModel

	// Get the manageable tenants
	tenants, err := d.service.GetManageableTenants(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading manageable tenants",
			fmt.Sprintf("Could not retrieve manageable tenants: %s", err.Error()),
		)
		return
	}

	// Convert to data source model
	state.Tenants = make([]manageableTenantModel, len(tenants))
	for i, tenant := range tenants {
		state.Tenants[i] = manageableTenantModel{
			EntraTenantID: types.StringValue(tenant.EntraTenantID),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
