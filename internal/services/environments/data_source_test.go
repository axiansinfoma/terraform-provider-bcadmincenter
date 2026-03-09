// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEnvironmentDataSource_Metadata(t *testing.T) {
	d := NewEnvironmentDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_environment"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestEnvironmentDataSource_Schema(t *testing.T) {
	d := NewEnvironmentDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	if _, ok := resp.Schema.Attributes["application_family"]; !ok {
		t.Error("Schema missing application_family attribute")
	}
	if _, ok := resp.Schema.Attributes["name"]; !ok {
		t.Error("Schema missing name attribute")
	}
	if _, ok := resp.Schema.Attributes["type"]; !ok {
		t.Error("Schema missing type attribute")
	}
	if _, ok := resp.Schema.Attributes["status"]; !ok {
		t.Error("Schema missing status attribute")
	}
}

func TestEnvironmentDataSource_Configure(t *testing.T) {
	d := &environmentDataSource{}

	// Test with nil provider data (should not error)
	req := datasource.ConfigureRequest{ProviderData: nil}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil ProviderData should not error: %v", resp.Diagnostics)
	}
}

func TestEnvironmentDataSource_Configure_WithInvalidType(t *testing.T) {
	d := &environmentDataSource{}

	// Test with invalid provider data type.
	req := datasource.ConfigureRequest{
		ProviderData: "invalid-type",
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure() with invalid type should return error")
	}
}

func TestEnvironmentDataSourceModel(t *testing.T) {
	// Test that the model struct can be created and populated.
	model := environmentDataSourceModel{
		ApplicationFamily:  types.StringValue("BusinessCentral"),
		Name:               types.StringValue("test-env"),
		Type:               types.StringValue("Sandbox"),
		CountryCode:        types.StringValue("US"),
		RingName:           types.StringValue("Production"),
		ApplicationVersion: types.StringValue("25.0"),
		Status:             types.StringValue("Active"),
		WebClientLoginURL:  types.StringValue("https://example.com"),
		AadTenantID:        types.StringValue("tenant-id"),
	}

	if model.Name.ValueString() != "test-env" {
		t.Errorf("Name = %v, want test-env", model.Name.ValueString())
	}
	if model.Type.ValueString() != "Sandbox" {
		t.Errorf("Type = %v, want Sandbox", model.Type.ValueString())
	}
}

func TestEnvironmentsDataSource_Metadata(t *testing.T) {
	d := NewEnvironmentsDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_environments"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestEnvironmentsDataSource_Schema(t *testing.T) {
	d := NewEnvironmentsDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	if _, ok := resp.Schema.Attributes["application_family"]; !ok {
		t.Error("Schema missing application_family attribute")
	}
	if _, ok := resp.Schema.Attributes["environments"]; !ok {
		t.Error("Schema missing environments attribute")
	}
}

func TestEnvironmentsDataSource_Configure(t *testing.T) {
	d := &environmentsDataSource{}

	// Test with nil provider data (should not error)
	req := datasource.ConfigureRequest{ProviderData: nil}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil ProviderData should not error: %v", resp.Diagnostics)
	}
}

func TestEnvironmentsDataSource_Configure_WithInvalidType(t *testing.T) {
	d := &environmentsDataSource{}

	// Test with invalid provider data type.
	req := datasource.ConfigureRequest{
		ProviderData: "invalid-type",
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure() with invalid type should return error")
	}
}

func TestEnvironmentsDataSourceModel(t *testing.T) {
	// Test that the model struct can be created and populated.
	model := environmentsDataSourceModel{
		ApplicationFamily: types.StringValue("BusinessCentral"),
		Environments: []environmentListItemModel{
			{
				Name:               types.StringValue("env1"),
				Type:               types.StringValue("Production"),
				CountryCode:        types.StringValue("US"),
				RingName:           types.StringValue("Production"),
				ApplicationVersion: types.StringValue("25.0"),
				Status:             types.StringValue("Active"),
				WebClientLoginURL:  types.StringValue("https://example1.com"),
				AadTenantID:        types.StringValue("tenant-id-1"),
			},
			{
				Name:               types.StringValue("env2"),
				Type:               types.StringValue("Sandbox"),
				CountryCode:        types.StringValue("CA"),
				RingName:           types.StringValue("Production"),
				ApplicationVersion: types.StringValue("25.0"),
				Status:             types.StringValue("Active"),
				WebClientLoginURL:  types.StringValue("https://example2.com"),
				AadTenantID:        types.StringValue("tenant-id-2"),
			},
		},
	}

	if model.ApplicationFamily.ValueString() != "BusinessCentral" {
		t.Errorf("ApplicationFamily = %v, want BusinessCentral", model.ApplicationFamily.ValueString())
	}
	if len(model.Environments) != 2 {
		t.Errorf("len(Environments) = %v, want 2", len(model.Environments))
	}
	if model.Environments[0].Name.ValueString() != "env1" {
		t.Errorf("Environments[0].Name = %v, want env1", model.Environments[0].Name.ValueString())
	}
	if model.Environments[1].Name.ValueString() != "env2" {
		t.Errorf("Environments[1].Name = %v, want env2", model.Environments[1].Name.ValueString())
	}
}

func TestEnvironmentUpdatesDataSource_Metadata(t *testing.T) {
	d := NewEnvironmentUpdatesDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_environment_updates"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestEnvironmentUpdatesDataSource_Schema(t *testing.T) {
	d := NewEnvironmentUpdatesDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	requiredAttrs := []string{"application_family", "environment_name", "updates"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing %s attribute", attr)
		}
	}
}

func TestEnvironmentUpdatesDataSource_Configure(t *testing.T) {
	d := &environmentUpdatesDataSource{}

	// Test with nil provider data (should not error).
	req := datasource.ConfigureRequest{ProviderData: nil}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil ProviderData should not error: %v", resp.Diagnostics)
	}
}

func TestEnvironmentUpdatesDataSource_Configure_WithInvalidType(t *testing.T) {
	d := &environmentUpdatesDataSource{}

	// Test with invalid provider data type.
	req := datasource.ConfigureRequest{
		ProviderData: "invalid-type",
	}
	resp := &datasource.ConfigureResponse{}

	d.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure() with invalid type should return error")
	}
}

func TestEnvironmentUpdatesDataSourceModel(t *testing.T) {
	// Test that the model struct can be created and populated.
	model := environmentUpdatesDataSourceModel{
		ApplicationFamily: types.StringValue("BusinessCentral"),
		EnvironmentName:   types.StringValue("production"),
		AadTenantID:       types.StringValue("tenant-id"),
		Updates:           types.ListNull(types.ObjectType{AttrTypes: updateItemAttrTypes}),
	}

	if model.ApplicationFamily.ValueString() != "BusinessCentral" {
		t.Errorf("ApplicationFamily = %v, want BusinessCentral", model.ApplicationFamily.ValueString())
	}
	if model.EnvironmentName.ValueString() != "production" {
		t.Errorf("EnvironmentName = %v, want production", model.EnvironmentName.ValueString())
	}
}
