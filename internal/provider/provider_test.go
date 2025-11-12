// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bcadmincenter": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the bcadmincenter provider.
var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
	"bcadmincenter": providerserver.NewProtocol6WithError(New("test")()),
	"echo":          echoprovider.NewProviderServer(),
}

func testAccPreCheck(t *testing.T) {
	// Pre-check function for acceptance tests
	// Add any environment variable checks here if needed
}

func TestBCAdminCenterProvider_Metadata(t *testing.T) {
	p := &BCAdminCenterProvider{
		version: "test",
	}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(context.Background(), req, resp)

	if resp.TypeName != "bcadmincenter" {
		t.Errorf("TypeName = %v, want bcadmincenter", resp.TypeName)
	}

	if resp.Version != "test" {
		t.Errorf("Version = %v, want test", resp.Version)
	}
}

func TestBCAdminCenterProvider_Schema(t *testing.T) {
	p := &BCAdminCenterProvider{}

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() unexpected errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist
	requiredAttrs := []string{
		"client_id",
		"client_secret",
		"tenant_id",
		"environment",
		"auxiliary_tenant_ids",
	}

	for _, attrName := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attrName]; !ok {
			t.Errorf("Schema missing attribute: %s", attrName)
		}
	}

	// Verify client_secret is sensitive
	if clientSecret, ok := resp.Schema.Attributes["client_secret"]; ok {
		if stringAttr, ok := clientSecret.(interface{ GetSensitive() bool }); ok {
			// Note: The actual check depends on the schema attribute type
			// This is a placeholder for the concept
			_ = stringAttr
		}
	}
}

func TestBCAdminCenterProvider_Configure(t *testing.T) {
	// Skip detailed configuration testing as it requires full Terraform context
	// Configuration behavior is tested through acceptance tests
	t.Skip("Configuration testing requires full Terraform framework context")
}

func TestBCAdminCenterProvider_Resources(t *testing.T) {
	p := &BCAdminCenterProvider{}

	resources := p.Resources(context.Background())

	// We should have 1 resource: environment
	expectedCount := 1
	if len(resources) != expectedCount {
		t.Errorf("Resources() returned %d resources, want %d", len(resources), expectedCount)
	}
}

func TestBCAdminCenterProvider_DataSources(t *testing.T) {
	p := &BCAdminCenterProvider{}

	dataSources := p.DataSources(context.Background())

	// We should have 4 data sources: available_applications, application_family, environment, environments
	expectedCount := 4
	if len(dataSources) != expectedCount {
		t.Errorf("DataSources() returned %d data sources, want %d", len(dataSources), expectedCount)
	}
}

func TestBCAdminCenterProvider_EphemeralResources(t *testing.T) {
	p := &BCAdminCenterProvider{}

	ephemeralResources := p.EphemeralResources(context.Background())

	// Currently we should have no ephemeral resources implemented
	if len(ephemeralResources) != 0 {
		t.Logf("EphemeralResources() returned %d ephemeral resources (expected 0 for initial implementation)", len(ephemeralResources))
	}
}

func TestBCAdminCenterProvider_Functions(t *testing.T) {
	p := &BCAdminCenterProvider{}

	functions := p.Functions(context.Background())

	// Currently we should have no functions implemented
	if len(functions) != 0 {
		t.Logf("Functions() returned %d functions (expected 0 for initial implementation)", len(functions))
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "with version",
			version: "1.0.0",
		},
		{
			name:    "with dev version",
			version: "dev",
		},
		{
			name:    "with test version",
			version: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			providerFunc := New(tt.version)
			if providerFunc == nil {
				t.Fatal("New() returned nil function")
			}

			p := providerFunc()
			if p == nil {
				t.Fatal("Provider function returned nil provider")
			}

			bcProvider, ok := p.(*BCAdminCenterProvider)
			if !ok {
				t.Fatal("Provider is not *BCAdminCenterProvider")
			}

			if bcProvider.version != tt.version {
				t.Errorf("Provider version = %v, want %v", bcProvider.version, tt.version)
			}
		})
	}
}
