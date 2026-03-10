// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentapps

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestEnvironmentAppResource_Metadata(t *testing.T) {
	r := NewEnvironmentAppResource()

	req := resource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_environment_app"
	if resp.TypeName != expected {
		t.Errorf("Metadata() TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestEnvironmentAppResource_Schema(t *testing.T) {
	r := NewEnvironmentAppResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() unexpected errors: %v", resp.Diagnostics)
	}

	expectedAttrs := []string{
		"id",
		"aad_tenant_id",
		"application_family",
		"environment_name",
		"app_id",
		"target_version",
		"allow_preview_version",
		"install_or_update_needed_dependencies",
		"accept_isv_eula",
		"language_id",
		"ignore_update_window",
		"name",
		"publisher",
		"published_as",
		"status",
		"timeouts",
	}

	for _, attrName := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attrName]; !ok {
			t.Errorf("Schema missing attribute: %s", attrName)
		}
	}
}

func TestEnvironmentAppResource_Configure(t *testing.T) {
	r := &EnvironmentAppResource{}

	// Nil provider data should not cause an error.
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil ProviderData should not error, got: %v", resp.Diagnostics)
	}

	if r.client != nil {
		t.Error("Configure() with nil ProviderData should not set client")
	}
}

func TestEnvironmentAppResource_Configure_InvalidType(t *testing.T) {
	r := &EnvironmentAppResource{}

	// Wrong type should produce an error.
	req := resource.ConfigureRequest{ProviderData: "not-a-client"}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure() with wrong ProviderData type should error")
	}
}
