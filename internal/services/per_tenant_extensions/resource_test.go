// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPerTenantExtensionResource_Metadata(t *testing.T) {
	r := NewPerTenantExtensionResource()

	req := resource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_per_tenant_extension"
	if resp.TypeName != expected {
		t.Errorf("Metadata() TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestPerTenantExtensionResource_Schema(t *testing.T) {
	r := NewPerTenantExtensionResource()

	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() unexpected errors: %v", resp.Diagnostics)
	}

	expectedAttrs := []string{
		"id",
		"aad_tenant_id",
		"company_id",
		"environment_name",
		"application_family",
		"file_path",
		"file_content",
		"file_sha256",
		"schedule",
		"schema_sync_mode",
		"delete_data",
		"unpublish_on_delete",
		"package_id",
		"app_id",
		"display_name",
		"publisher",
		"version",
	}

	for _, attrName := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attrName]; !ok {
			t.Errorf("Schema missing attribute: %s", attrName)
		}
	}
}

func TestPerTenantExtensionResource_Configure(t *testing.T) {
	r := &PerTenantExtensionResource{}

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

func TestPerTenantExtensionResource_Configure_InvalidType(t *testing.T) {
	r := &PerTenantExtensionResource{}

	// Wrong type should produce an error.
	req := resource.ConfigureRequest{ProviderData: "not-a-client"}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure() with wrong ProviderData type should error")
	}
}

func TestValidateFileInputs(t *testing.T) {
	tests := []struct {
		name        string
		filePath    string
		fileContent string
		wantErr     bool
	}{
		{
			name:        "only file_path set",
			filePath:    "/path/to/extension.app",
			fileContent: "",
			wantErr:     false,
		},
		{
			name:        "only file_content set",
			filePath:    "",
			fileContent: "base64encodedcontent",
			wantErr:     false,
		},
		{
			name:        "both set",
			filePath:    "/path/to/extension.app",
			fileContent: "base64encodedcontent",
			wantErr:     true,
		},
		{
			name:        "neither set",
			filePath:    "",
			fileContent: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var data PerTenantExtensionResourceModel
			if tt.filePath != "" {
				data.FilePath = types.StringValue(tt.filePath)
			} else {
				data.FilePath = types.StringNull()
			}
			if tt.fileContent != "" {
				data.FileContent = types.StringValue(tt.fileContent)
			} else {
				data.FileContent = types.StringNull()
			}

			err := validateFileInputs(&data)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateFileInputs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
