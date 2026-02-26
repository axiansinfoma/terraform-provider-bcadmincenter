// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package quotas

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
	_ datasource.DataSource              = &QuotasDataSource{}
	_ datasource.DataSourceWithConfigure = &QuotasDataSource{}
)

// NewQuotasDataSource is a helper function to simplify the provider implementation.
func NewQuotasDataSource() datasource.DataSource {
	return &QuotasDataSource{}
}

// QuotasDataSource is the data source implementation.
type QuotasDataSource struct {
	client *client.Client
}

// QuotasDataSourceModel describes the data source data model.
type QuotasDataSourceModel struct {
	ID                              types.String `tfsdk:"id"`
	ProductionEnvironmentsQuota     types.Int64  `tfsdk:"production_environments_quota"`
	ProductionEnvironmentsAllocated types.Int64  `tfsdk:"production_environments_allocated"`
	ProductionEnvironmentsAvailable types.Int64  `tfsdk:"production_environments_available"`
	SandboxEnvironmentsQuota        types.Int64  `tfsdk:"sandbox_environments_quota"`
	SandboxEnvironmentsAllocated    types.Int64  `tfsdk:"sandbox_environments_allocated"`
	SandboxEnvironmentsAvailable    types.Int64  `tfsdk:"sandbox_environments_available"`
	StorageQuotaGB                  types.Int64  `tfsdk:"storage_quota_gb"`
	StorageAllocatedGB              types.Int64  `tfsdk:"storage_allocated_gb"`
	StorageAvailableGB              types.Int64  `tfsdk:"storage_available_gb"`
}

// Metadata returns the data source type name.
func (d *QuotasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_quotas"
}

// Schema defines the schema for the data source.
func (d *QuotasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves environment quotas and capacity information for the tenant, including production and sandbox environment limits and storage capacity.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for the data source",
				Computed:    true,
			},
			"production_environments_quota": schema.Int64Attribute{
				Description: "Total number of production environments allowed for the tenant",
				Computed:    true,
			},
			"production_environments_allocated": schema.Int64Attribute{
				Description: "Number of production environments currently allocated",
				Computed:    true,
			},
			"production_environments_available": schema.Int64Attribute{
				Description: "Number of production environments available to create (quota - allocated)",
				Computed:    true,
			},
			"sandbox_environments_quota": schema.Int64Attribute{
				Description: "Total number of sandbox environments allowed for the tenant",
				Computed:    true,
			},
			"sandbox_environments_allocated": schema.Int64Attribute{
				Description: "Number of sandbox environments currently allocated",
				Computed:    true,
			},
			"sandbox_environments_available": schema.Int64Attribute{
				Description: "Number of sandbox environments available to create (quota - allocated)",
				Computed:    true,
			},
			"storage_quota_gb": schema.Int64Attribute{
				Description: "Total storage quota in gigabytes",
				Computed:    true,
			},
			"storage_allocated_gb": schema.Int64Attribute{
				Description: "Storage currently allocated in gigabytes",
				Computed:    true,
			},
			"storage_available_gb": schema.Int64Attribute{
				Description: "Storage available in gigabytes (quota - allocated)",
				Computed:    true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *QuotasDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *QuotasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state QuotasDataSourceModel

	// Create service.
	svc := NewService(d.client)

	// Get quotas.
	result, err := svc.GetQuotas(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Quotas",
			"Could not read environment quotas: "+err.Error(),
		)
		return
	}

	// Set static ID.
	state.ID = types.StringValue("quotas")

	// Map response to state.
	state.ProductionEnvironmentsQuota = types.Int64Value(int64(result.ProductionEnvironmentsQuota))
	state.ProductionEnvironmentsAllocated = types.Int64Value(int64(result.ProductionEnvironmentsAllocated))
	state.ProductionEnvironmentsAvailable = types.Int64Value(int64(result.ProductionEnvironmentsQuota - result.ProductionEnvironmentsAllocated))

	state.SandboxEnvironmentsQuota = types.Int64Value(int64(result.SandboxEnvironmentsQuota))
	state.SandboxEnvironmentsAllocated = types.Int64Value(int64(result.SandboxEnvironmentsAllocated))
	state.SandboxEnvironmentsAvailable = types.Int64Value(int64(result.SandboxEnvironmentsQuota - result.SandboxEnvironmentsAllocated))

	state.StorageQuotaGB = types.Int64Value(int64(result.StorageQuotaGB))
	state.StorageAllocatedGB = types.Int64Value(int64(result.StorageAllocatedGB))
	state.StorageAvailableGB = types.Int64Value(int64(result.StorageQuotaGB - result.StorageAllocatedGB))

	// Set state.
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
