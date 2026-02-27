// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentsettings

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestEnvironmentSettingsResource_Metadata(t *testing.T) {
	r := NewEnvironmentSettingsResource()
	req := resource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_environment_settings"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestEnvironmentSettingsResource_Schema(t *testing.T) {
	r := NewEnvironmentSettingsResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	requiredAttrs := []string{"application_family", "environment_name"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing required attribute: %s", attr)
		}
	}

	// Verify optional settings exist.
	optionalAttrs := []string{
		"aad_tenant_id",
		"update_window_start_time",
		"update_window_end_time",
		"update_window_timezone",
		"app_insights_key",
		"security_group_id",
		"access_with_m365_licenses",
		"app_update_cadence",
		"partner_access_status",
		"allowed_partner_tenant_ids",
	}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing optional attribute: %s", attr)
		}
	}

	// Verify computed id exists.
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing computed 'id' attribute")
	}
}

func TestEnvironmentSettingsResource_Configure(t *testing.T) {
	r, ok := NewEnvironmentSettingsResource().(*EnvironmentSettingsResource)
	if !ok {
		t.Fatal("Failed to cast to EnvironmentSettingsResource")
	}

	// Test with nil provider data (should not error)
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil data should not error, got: %v", resp.Diagnostics)
	}
}
