// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package environmentsupportcontact

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestEnvironmentSupportContactResource_Metadata(t *testing.T) {
	r := NewEnvironmentSupportContactResource()
	req := resource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_environment_support_contact"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestEnvironmentSupportContactResource_Schema(t *testing.T) {
	r := NewEnvironmentSupportContactResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	requiredAttrs := []string{"application_family", "environment_name", "name", "email"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing required attribute: %s", attr)
		}
	}

	// Verify optional attributes exist.
	optionalAttrs := []string{"url"}
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

func TestEnvironmentSupportContactResource_Configure(t *testing.T) {
	r, ok := NewEnvironmentSupportContactResource().(*EnvironmentSupportContactResource)
	if !ok {
		t.Fatal("Failed to cast to EnvironmentSupportContactResource")
	}

	// Test with nil provider data (should not error)
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil data should not error, got: %v", resp.Diagnostics)
	}
}
