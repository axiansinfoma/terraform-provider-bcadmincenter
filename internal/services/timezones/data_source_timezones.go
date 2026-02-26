// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package timezones

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
	_ datasource.DataSource              = &TimeZonesDataSource{}
	_ datasource.DataSourceWithConfigure = &TimeZonesDataSource{}
)

// NewTimeZonesDataSource is a helper function to simplify the provider implementation.
func NewTimeZonesDataSource() datasource.DataSource {
	return &TimeZonesDataSource{}
}

// TimeZonesDataSource is the data source implementation.
type TimeZonesDataSource struct {
	client *client.Client
}

// TimeZonesDataSourceModel describes the data source data model.
type TimeZonesDataSourceModel struct {
	ID        types.String    `tfsdk:"id"`
	TimeZones []TimeZoneModel `tfsdk:"timezones"`
}

// TimeZoneModel represents a single timezone.
type TimeZoneModel struct {
	ID                      types.String `tfsdk:"id"`
	DisplayName             types.String `tfsdk:"display_name"`
	SupportsDaylightSavings types.Bool   `tfsdk:"supports_daylight_savings"`
	OffsetFromUTC           types.String `tfsdk:"offset_from_utc"`
}

// Metadata returns the data source type name.
func (d *TimeZonesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_timezones"
}

// Schema defines the schema for the data source.
func (d *TimeZonesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the list of valid time zone identifiers that can be used for environment update window settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Static identifier for the data source",
				Computed:    true,
			},
			"timezones": schema.ListNestedAttribute{
				Description: "List of available time zones",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Time zone identifier (e.g., 'Pacific Standard Time', 'Central European Standard Time')",
							Computed:    true,
						},
						"display_name": schema.StringAttribute{
							Description: "Human-readable display name for the time zone",
							Computed:    true,
						},
						"supports_daylight_savings": schema.BoolAttribute{
							Description: "Whether this time zone observes daylight saving time",
							Computed:    true,
						},
						"offset_from_utc": schema.StringAttribute{
							Description: "Current offset from UTC (e.g., '-08:00', '+01:00')",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *TimeZonesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *TimeZonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TimeZonesDataSourceModel

	// Create service.
	svc := NewService(d.client)

	// Get timezones.
	result, err := svc.GetTimeZones(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Time Zones",
			"Could not read time zones: "+err.Error(),
		)
		return
	}

	// Set static ID.
	state.ID = types.StringValue("timezones")

	// Map response to state.
	state.TimeZones = make([]TimeZoneModel, 0, len(result.Value))
	for _, tz := range result.Value {
		state.TimeZones = append(state.TimeZones, TimeZoneModel{
			ID:                      types.StringValue(tz.ID),
			DisplayName:             types.StringValue(tz.DisplayName),
			SupportsDaylightSavings: types.BoolValue(tz.SupportsDaylightSavings),
			OffsetFromUTC:           types.StringValue(tz.OffsetFromUTC),
		})
	}

	// Set state.
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
