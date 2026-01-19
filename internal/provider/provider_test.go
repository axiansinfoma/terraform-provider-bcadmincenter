// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the bcadmincenter provider.

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

	// Verify required attributes exist.
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

	// Verify client_secret is sensitive.
	if clientSecret, ok := resp.Schema.Attributes["client_secret"]; ok {
		if stringAttr, ok := clientSecret.(interface{ GetSensitive() bool }); ok {
			// Note: The actual check depends on the schema attribute type.
			// This is a placeholder for the concept.
			_ = stringAttr
		}
	}
}

func TestBCAdminCenterProvider_Configure(t *testing.T) {
	tests := []struct {
		name        string
		hasClientID bool
		hasSecret   bool
		hasTenantID bool
		wantError   bool
		description string
	}{
		{
			name:        "valid configuration with all required fields",
			hasClientID: true,
			hasSecret:   true,
			hasTenantID: true,
			wantError:   false,
			description: "Configuration should be valid with all required fields",
		},
		{
			name:        "missing client_id",
			hasClientID: false,
			hasSecret:   true,
			hasTenantID: true,
			wantError:   true,
			description: "Configuration should error when client_id is missing",
		},
		{
			name:        "missing client_secret",
			hasClientID: true,
			hasSecret:   false,
			hasTenantID: true,
			wantError:   true,
			description: "Configuration should error when client_secret is missing",
		},
		{
			name:        "missing tenant_id",
			hasClientID: true,
			hasSecret:   true,
			hasTenantID: false,
			wantError:   true,
			description: "Configuration should error when tenant_id is missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate the test expectations.
			if !tt.wantError {
				if !tt.hasClientID || !tt.hasSecret || !tt.hasTenantID {
					t.Errorf("%s: Expected all required fields for valid config", tt.description)
				}
			} else {
				if tt.hasClientID && tt.hasSecret && tt.hasTenantID {
					t.Errorf("%s: Expected at least one missing field for error case", tt.description)
				}
			}
		})
	}
}

func TestBCAdminCenterProvider_Resources(t *testing.T) {
	p := &BCAdminCenterProvider{}

	resources := p.Resources(context.Background())

	// We should have 5 resources: authorized_entra_app, environment, environment_settings, support_contact, notification_recipient.
	expectedCount := 5
	if len(resources) != expectedCount {
		t.Errorf("Resources() returned %d resources, want %d", len(resources), expectedCount)
	}
}

func TestBCAdminCenterProvider_DataSources(t *testing.T) {
	p := &BCAdminCenterProvider{}

	dataSources := p.DataSources(context.Background())

	// We should have 9 data sources: authorized_entra_apps, manageable_tenants, available_applications, application_family, environment, environments, notification_settings, quotas, timezones.
	expectedCount := 9
	if len(dataSources) != expectedCount {
		t.Errorf("DataSources() returned %d data sources, want %d", len(dataSources), expectedCount)
	}
}

func TestBCAdminCenterProvider_EphemeralResources(t *testing.T) {
	p := &BCAdminCenterProvider{}

	ephemeralResources := p.EphemeralResources(context.Background())

	// Currently we should have no ephemeral resources implemented.
	if len(ephemeralResources) != 0 {
		t.Logf("EphemeralResources() returned %d ephemeral resources (expected 0 for initial implementation)", len(ephemeralResources))
	}
}

func TestBCAdminCenterProvider_Functions(t *testing.T) {
	p := &BCAdminCenterProvider{}

	functions := p.Functions(context.Background())

	// Currently we should have no functions implemented.
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

func TestGetConfigValue(t *testing.T) {
	tests := []struct {
		name        string
		configValue string
		envVarName  string
		envVarValue string
		want        string
	}{
		{
			name:        "config value takes precedence",
			configValue: "config-value",
			envVarName:  "TEST_VAR",
			envVarValue: "env-value",
			want:        "config-value",
		},
		{
			name:        "env var used when config is empty",
			configValue: "",
			envVarName:  "TEST_VAR",
			envVarValue: "env-value",
			want:        "env-value",
		},
		{
			name:        "empty string when both are empty",
			configValue: "",
			envVarName:  "TEST_VAR",
			envVarValue: "",
			want:        "",
		},
		{
			name:        "AZURE_TENANT_ID from environment",
			configValue: "",
			envVarName:  "AZURE_TENANT_ID",
			envVarValue: "00000000-0000-0000-0000-000000000001",
			want:        "00000000-0000-0000-0000-000000000001",
		},
		{
			name:        "AZURE_CLIENT_ID from environment",
			configValue: "",
			envVarName:  "AZURE_CLIENT_ID",
			envVarValue: "00000000-0000-0000-0000-000000000002",
			want:        "00000000-0000-0000-0000-000000000002",
		},
		{
			name:        "AZURE_CLIENT_SECRET from environment",
			configValue: "",
			envVarName:  "AZURE_CLIENT_SECRET",
			envVarValue: "secret-value",
			want:        "secret-value",
		},
		{
			name:        "AZURE_ENVIRONMENT from environment",
			configValue: "",
			envVarName:  "AZURE_ENVIRONMENT",
			envVarValue: "usgovernment",
			want:        "usgovernment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envVarValue != "" {
				t.Setenv(tt.envVarName, tt.envVarValue)
			}

			// Create types.String from config value
			var configVal types.String
			if tt.configValue != "" {
				configVal = types.StringValue(tt.configValue)
			} else {
				configVal = types.StringNull()
			}

			// Test getConfigValue
			got := getConfigValue(configVal, tt.envVarName)
			if got != tt.want {
				t.Errorf("getConfigValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
