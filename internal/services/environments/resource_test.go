// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestEnvironmentResource_Metadata(t *testing.T) {
	r := NewEnvironmentResource()
	req := resource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_environment"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestEnvironmentResource_Schema(t *testing.T) {
	r := NewEnvironmentResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist
	requiredAttrs := []string{"name", "type", "country_code"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing required attribute: %s", attr)
		}
	}

	// Verify optional attributes exist
	optionalAttrs := []string{"application_family", "ring_name", "application_version", "azure_region"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing optional attribute: %s", attr)
		}
	}

	// Verify computed attributes exist
	computedAttrs := []string{"id", "status", "web_client_login_url", "web_service_url", "app_insights_key", "platform_version", "aad_tenant_id"}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing computed attribute: %s", attr)
		}
	}
}

func TestEnvironmentResource_Configure(t *testing.T) {
	r := &EnvironmentResource{}

	// Test with nil provider data (should not error, just skip)
	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil data returned errors: %v", resp.Diagnostics)
	}

	if r.client != nil {
		t.Error("Configure() with nil data should not set client")
	}
}

func TestEnvironmentResourceModel(t *testing.T) {
	// Test that the model struct can be created and populated
	model := EnvironmentResourceModel{}

	// Verify the struct has all expected fields
	_ = model.ID
	_ = model.Name
	_ = model.ApplicationFamily
	_ = model.Type
	_ = model.CountryCode
	_ = model.RingName
	_ = model.ApplicationVersion
	_ = model.AzureRegion
	_ = model.Status
	_ = model.WebClientLoginURL
	_ = model.WebServiceURL
	_ = model.AppInsightsKey
	_ = model.PlatformVersion
	_ = model.AADTenantID
	_ = model.Timeouts
}

func TestEnvironmentResource_ImportState(t *testing.T) {
	tests := []struct {
		name      string
		importID  string
		wantError bool
	}{
		{
			name:      "invalid import ID - too few parts",
			importID:  "BusinessCentral/test-sandbox",
			wantError: true,
		},
		{
			name:      "invalid import ID - too many parts",
			importID:  "tenant/BusinessCentral/test/extra",
			wantError: true,
		},
		{
			name:      "invalid import ID - empty",
			importID:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &EnvironmentResource{}
			req := resource.ImportStateRequest{
				ID: tt.importID,
			}
			resp := &resource.ImportStateResponse{}

			r.ImportState(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("ImportState() hasError = %v, wantError %v, diagnostics: %v", hasError, tt.wantError, resp.Diagnostics)
			}
		})
	}
}

func TestNewEnvironmentResource(t *testing.T) {
	r := NewEnvironmentResource()

	if r == nil {
		t.Error("NewEnvironmentResource() returned nil")
	}

	// Verify it returns a resource
	_, ok := r.(resource.Resource)
	if !ok {
		t.Error("NewEnvironmentResource() does not implement resource.Resource")
	}
}
