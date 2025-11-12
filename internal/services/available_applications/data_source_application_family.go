// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package available_applications

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vllni/terraform-provider-bc-admin-center/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ApplicationFamilyDataSource{}

func NewApplicationFamilyDataSource() datasource.DataSource {
	return &ApplicationFamilyDataSource{}
}

// ApplicationFamilyDataSource defines the data source implementation.
type ApplicationFamilyDataSource struct {
	client *client.Client
}

// ApplicationFamilyDataSourceModel describes the data source data model.
type ApplicationFamilyDataSourceModel struct {
	Name                 types.String              `tfsdk:"name"`
	CountriesRingDetails []CountryRingDetailsModel `tfsdk:"countries_ring_details"`
	ID                   types.String              `tfsdk:"id"`
}

func (d *ApplicationFamilyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_family"
}

func (d *ApplicationFamilyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves information about a specific application family including available countries/regions and rings. Use this data source to get detailed information for a single application family.",

		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The name of the application family to retrieve (e.g., 'BusinessCentral')",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Data source identifier",
			},
			"countries_ring_details": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of countries/regions with their available rings",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"country_code": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Code for the country/region (e.g., 'US', 'GB', 'DK')",
						},
						"rings": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "List of available rings for this country/region",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The API name of the ring (e.g., 'PROD', 'PREVIEW')",
									},
									"production_ring": schema.BoolAttribute{
										Computed:            true,
										MarkdownDescription: "Indicates whether this is a production ring",
									},
									"friendly_name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "The display-friendly name of the ring",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *ApplicationFamilyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

func (d *ApplicationFamilyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ApplicationFamilyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create service
	svc := NewService(d.client)

	// Get the specific application family
	appFamily, err := svc.GetApplicationFamily(ctx, data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving application family",
			fmt.Sprintf("Could not retrieve application family '%s': %s", data.Name.ValueString(), err.Error()),
		)
		return
	}

	// Map API response to Terraform state
	data.CountriesRingDetails = make([]CountryRingDetailsModel, 0, len(appFamily.CountriesRingDetails))

	for _, countryRingDetails := range appFamily.CountriesRingDetails {
		countryModel := CountryRingDetailsModel{
			CountryCode: types.StringValue(countryRingDetails.CountryCode),
			Rings:       make([]RingModel, 0, len(countryRingDetails.Rings)),
		}

		for _, ring := range countryRingDetails.Rings {
			ringModel := RingModel{
				Name:           types.StringValue(ring.Name),
				ProductionRing: types.BoolValue(ring.ProductionRing),
				FriendlyName:   types.StringValue(ring.FriendlyName),
			}
			countryModel.Rings = append(countryModel.Rings, ringModel)
		}

		data.CountriesRingDetails = append(data.CountriesRingDetails, countryModel)
	}

	// Set ID to the application family name
	data.ID = types.StringValue(data.Name.ValueString())

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
