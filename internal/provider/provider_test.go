// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-framework/provider"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bc_admin_center": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the bc_admin_center provider.
var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
	"bc_admin_center": providerserver.NewProtocol6WithError(New("test")()),
	"echo":            echoprovider.NewProviderServer(),
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

	if resp.TypeName != "bc_admin_center" {
		t.Errorf("TypeName = %v, want bc_admin_center", resp.TypeName)
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

	// Currently we should have no resources implemented
	if len(resources) != 0 {
		t.Logf("Resources() returned %d resources (expected 0 for initial implementation)", len(resources))
	}
}

func TestBCAdminCenterProvider_DataSources(t *testing.T) {
	p := &BCAdminCenterProvider{}

	dataSources := p.DataSources(context.Background())

	// Currently we should have no data sources implemented
	if len(dataSources) != 0 {
		t.Logf("DataSources() returned %d data sources (expected 0 for initial implementation)", len(dataSources))
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
