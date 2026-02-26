// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package available_applications

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAvailableApplicationsDataSource_Metadata(t *testing.T) {
	d := NewAvailableApplicationsDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "bcadmincenter",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_available_applications"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestAvailableApplicationsDataSource_Schema(t *testing.T) {
	d := NewAvailableApplicationsDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() unexpected errors: %v", resp.Diagnostics)
	}

	// Verify schema has required attributes.
	if resp.Schema.Attributes == nil {
		t.Fatal("Schema.Attributes is nil")
	}

	// Check for id attribute.
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing 'id' attribute")
	}

	// Check for application_families attribute.
	if _, ok := resp.Schema.Attributes["application_families"]; !ok {
		t.Error("Schema missing 'application_families' attribute")
	}
}

func TestAvailableApplicationsDataSource_Configure(t *testing.T) {
	d := &AvailableApplicationsDataSource{}

	// Test with nil provider data.
	req := datasource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil provider data should not error, got: %v", resp.Diagnostics)
	}

	// Test with invalid provider data type.
	req = datasource.ConfigureRequest{
		ProviderData: "invalid",
	}
	resp = &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure() with invalid provider data should error")
	}
}

func TestApplicationFamilyDataSource_Metadata(t *testing.T) {
	d := NewApplicationFamilyDataSource()

	req := datasource.MetadataRequest{
		ProviderTypeName: "bcadmincenter",
	}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_application_family"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestApplicationFamilyDataSource_Schema(t *testing.T) {
	d := NewApplicationFamilyDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() unexpected errors: %v", resp.Diagnostics)
	}

	// Verify schema has required attributes.
	if resp.Schema.Attributes == nil {
		t.Fatal("Schema.Attributes is nil")
	}

	// Check for required 'name' attribute.
	if nameAttr, ok := resp.Schema.Attributes["name"]; !ok {
		t.Error("Schema missing 'name' attribute")
	} else {
		// Verify name is required.
		if !nameAttr.IsRequired() {
			t.Error("'name' attribute should be required")
		}
	}

	// Check for id attribute.
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing 'id' attribute")
	}

	// Check for countries_ring_details attribute.
	if _, ok := resp.Schema.Attributes["countries_ring_details"]; !ok {
		t.Error("Schema missing 'countries_ring_details' attribute")
	}
}

func TestApplicationFamilyDataSource_Configure(t *testing.T) {
	d := &ApplicationFamilyDataSource{}

	// Test with nil provider data.
	req := datasource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil provider data should not error, got: %v", resp.Diagnostics)
	}

	// Test with invalid provider data type.
	req = datasource.ConfigureRequest{
		ProviderData: "invalid",
	}
	resp = &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure() with invalid provider data should error")
	}
}

func TestAvailableApplicationsDataSourceModel(t *testing.T) {
	// Test model creation.
	model := AvailableApplicationsDataSourceModel{
		ID: types.StringValue("test-id"),
		ApplicationFamilies: []ApplicationFamilyModel{
			{
				Name: types.StringValue("BusinessCentral"),
				CountriesRingDetails: []CountryRingDetailsModel{
					{
						CountryCode: types.StringValue("US"),
						Rings: []RingModel{
							{
								Name:           types.StringValue("PROD"),
								ProductionRing: types.BoolValue(true),
								FriendlyName:   types.StringValue("Production"),
							},
						},
					},
				},
			},
		},
	}

	if model.ID.ValueString() != "test-id" {
		t.Errorf("ID = %v, want test-id", model.ID.ValueString())
	}

	if len(model.ApplicationFamilies) != 1 {
		t.Errorf("ApplicationFamilies length = %d, want 1", len(model.ApplicationFamilies))
	}

	if model.ApplicationFamilies[0].Name.ValueString() != "BusinessCentral" {
		t.Errorf("ApplicationFamily Name = %v, want BusinessCentral", model.ApplicationFamilies[0].Name.ValueString())
	}
}

func TestApplicationFamilyDataSourceModel(t *testing.T) {
	// Test model creation.
	model := ApplicationFamilyDataSourceModel{
		Name: types.StringValue("BusinessCentral"),
		ID:   types.StringValue("BusinessCentral"),
		CountriesRingDetails: []CountryRingDetailsModel{
			{
				CountryCode: types.StringValue("US"),
				Rings: []RingModel{
					{
						Name:           types.StringValue("PROD"),
						ProductionRing: types.BoolValue(true),
						FriendlyName:   types.StringValue("Production"),
					},
					{
						Name:           types.StringValue("PREVIEW"),
						ProductionRing: types.BoolValue(false),
						FriendlyName:   types.StringValue("Preview"),
					},
				},
			},
		},
	}

	if model.Name.ValueString() != "BusinessCentral" {
		t.Errorf("Name = %v, want BusinessCentral", model.Name.ValueString())
	}

	if model.ID.ValueString() != "BusinessCentral" {
		t.Errorf("ID = %v, want BusinessCentral", model.ID.ValueString())
	}

	if len(model.CountriesRingDetails) != 1 {
		t.Errorf("CountriesRingDetails length = %d, want 1", len(model.CountriesRingDetails))
	}

	if len(model.CountriesRingDetails[0].Rings) != 2 {
		t.Errorf("Rings length = %d, want 2", len(model.CountriesRingDetails[0].Rings))
	}

	// Verify production ring.
	prodRing := model.CountriesRingDetails[0].Rings[0]
	if !prodRing.ProductionRing.ValueBool() {
		t.Error("First ring should be a production ring")
	}

	// Verify preview ring.
	previewRing := model.CountriesRingDetails[0].Rings[1]
	if previewRing.ProductionRing.ValueBool() {
		t.Error("Second ring should not be a production ring")
	}
}
