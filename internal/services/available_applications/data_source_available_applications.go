// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package available_applications

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AvailableApplicationsDataSource{}

func NewAvailableApplicationsDataSource() datasource.DataSource {
	return &AvailableApplicationsDataSource{}
}

// AvailableApplicationsDataSource defines the data source implementation.
type AvailableApplicationsDataSource struct {
	client *client.Client
}

// AvailableApplicationsDataSourceModel describes the data source data model.
type AvailableApplicationsDataSourceModel struct {
	ApplicationFamilies []ApplicationFamilyModel `tfsdk:"application_families"`
	ID                  types.String             `tfsdk:"id"`
}

type ApplicationFamilyModel struct {
	Name                 types.String              `tfsdk:"name"`
	CountriesRingDetails []CountryRingDetailsModel `tfsdk:"countries_ring_details"`
}

type CountryRingDetailsModel struct {
	CountryCode types.String `tfsdk:"country_code"`
	Rings       []RingModel  `tfsdk:"rings"`
}

type RingModel struct {
	Name           types.String `tfsdk:"name"`
	ProductionRing types.Bool   `tfsdk:"production_ring"`
	FriendlyName   types.String `tfsdk:"friendly_name"`
}

func (d *AvailableApplicationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_available_applications"
}

func (d *AvailableApplicationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves the list of available application families with their countries/regions and rings. Use this data source to discover what values can be used for environment creation.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Data source identifier",
			},
			"application_families": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of available application families",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the application family (typically 'BusinessCentral')",
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
				},
			},
		},
	}
}

func (d *AvailableApplicationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *AvailableApplicationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AvailableApplicationsDataSourceModel

	// Read Terraform configuration data into the model.
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Create service.
	svc := NewService(d.client)

	// Get available applications.
	availableApps, err := svc.GetAvailableApplications(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving available applications",
			fmt.Sprintf("Could not retrieve available applications: %s", err.Error()),
		)
		return
	}

	// Map API response to Terraform state.
	data.ApplicationFamilies = make([]ApplicationFamilyModel, 0, len(availableApps.Value))

	for _, appFamily := range availableApps.Value {
		appFamilyModel := ApplicationFamilyModel{
			Name:                 types.StringValue(appFamily.ApplicationFamily),
			CountriesRingDetails: make([]CountryRingDetailsModel, 0, len(appFamily.CountriesRingDetails)),
		}

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

			appFamilyModel.CountriesRingDetails = append(appFamilyModel.CountriesRingDetails, countryModel)
		}

		data.ApplicationFamilies = append(data.ApplicationFamilies, appFamilyModel)
	}

	// Set a static ID since this is a singleton data source.
	data.ID = types.StringValue("available_applications")

	// Save data into Terraform state.
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
