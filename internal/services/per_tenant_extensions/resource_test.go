// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/utils"
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
		"environment_name",
		"application_family",
		"file_path",
		"file_content",
		"file_name",
		"file_sha256",
		"deployment_schedule",
		"sync_mode",
		"language_id",
		"accept_isv_eula",
		"install_or_update_needed_dependencies",
		"delete_data",
		"uninstall_dependents",
		"uninstall_in_update_window",
		"cancel_scheduled_on_destroy",
		"timeouts",
		"app_id",
		"display_name",
		"publisher",
		"version",
		"state",
		"app_type",
		"last_operation_id",
		"pending_target_version",
		"pending_schedule_kind",
	}

	for _, attrName := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attrName]; !ok {
			t.Errorf("Schema missing attribute: %s", attrName)
		}
	}

	// Attributes that only existed in the pre-2.29 Automation API implementation.
	removedAttrs := []string{"company_id", "schedule", "schema_sync_mode", "unpublish_on_delete", "package_id"}
	for _, attrName := range removedAttrs {
		if _, ok := resp.Schema.Attributes[attrName]; ok {
			t.Errorf("Schema still defines removed Automation API attribute: %s", attrName)
		}
	}

	if !resp.Schema.Attributes["accept_isv_eula"].IsRequired() {
		t.Error("accept_isv_eula should be required so the EULA is accepted explicitly")
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

func TestResolveFileName(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		filePath    string
		fileContent string
		want        string
	}{
		{
			name:     "explicit file_name wins",
			fileName: "Explicit.app",
			filePath: "/tmp/FromPath.app",
			want:     "Explicit.app",
		},
		{
			name:     "derived from file_path",
			filePath: "/tmp/extensions/MyExtension_1.0.0.0.app",
			want:     "MyExtension_1.0.0.0.app",
		},
		{
			name:        "fallback for base64 content",
			fileContent: "Zm9v",
			want:        "extension.app",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &PerTenantExtensionResourceModel{
				FileName:    stringOrNullValue(tt.fileName),
				FilePath:    stringOrNullValue(tt.filePath),
				FileContent: stringOrNullValue(tt.fileContent),
			}

			if got := resolveFileName(data); got != tt.want {
				t.Errorf("resolveFileName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveFileBytes(t *testing.T) {
	t.Run("from file_path", func(t *testing.T) {
		dir := t.TempDir()
		filePath := filepath.Join(dir, "MyExtension.app")
		if err := os.WriteFile(filePath, []byte("app-bytes"), 0o600); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		data := &PerTenantExtensionResourceModel{FilePath: types.StringValue(filePath)}

		got, err := resolveFileBytes(data)
		if err != nil {
			t.Fatalf("resolveFileBytes() unexpected error: %v", err)
		}
		if string(got) != "app-bytes" {
			t.Errorf("resolveFileBytes() = %q, want app-bytes", string(got))
		}
	})

	t.Run("from file_content", func(t *testing.T) {
		data := &PerTenantExtensionResourceModel{
			FileContent: types.StringValue(base64.StdEncoding.EncodeToString([]byte("app-bytes"))),
		}

		got, err := resolveFileBytes(data)
		if err != nil {
			t.Fatalf("resolveFileBytes() unexpected error: %v", err)
		}
		if string(got) != "app-bytes" {
			t.Errorf("resolveFileBytes() = %q, want app-bytes", string(got))
		}
	})

	t.Run("invalid base64", func(t *testing.T) {
		data := &PerTenantExtensionResourceModel{FileContent: types.StringValue("not-base64!!!")}

		if _, err := resolveFileBytes(data); err == nil {
			t.Error("resolveFileBytes() expected error for invalid base64, got nil")
		}
	})

	t.Run("neither set", func(t *testing.T) {
		if _, err := resolveFileBytes(&PerTenantExtensionResourceModel{}); err == nil {
			t.Error("resolveFileBytes() expected error when neither input is set, got nil")
		}
	})
}

func TestOperationTimeout(t *testing.T) {
	timeoutsType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"create": types.StringType,
		"update": types.StringType,
		"delete": types.StringType,
	}}

	withValues := func(create, update, del attr.Value) types.Object {
		obj, diags := types.ObjectValue(timeoutsType.AttrTypes, map[string]attr.Value{
			"create": create,
			"update": update,
			"delete": del,
		})
		if diags.HasError() {
			t.Fatalf("failed to build timeouts object: %v", diags)
		}
		return obj
	}

	tests := []struct {
		name     string
		timeouts types.Object
		key      string
		want     time.Duration
	}{
		{
			name:     "null object falls back to the default",
			timeouts: types.ObjectNull(timeoutsType.AttrTypes),
			key:      "create",
			want:     utils.DefaultOperationTimeout,
		},
		{
			name:     "configured value is used",
			timeouts: withValues(types.StringValue("90m"), types.StringNull(), types.StringNull()),
			key:      "create",
			want:     90 * time.Minute,
		},
		{
			name:     "unset key falls back to the default",
			timeouts: withValues(types.StringValue("90m"), types.StringNull(), types.StringNull()),
			key:      "delete",
			want:     utils.DefaultOperationTimeout,
		},
		{
			name:     "unparseable value falls back to the default",
			timeouts: withValues(types.StringValue("not-a-duration"), types.StringNull(), types.StringNull()),
			key:      "create",
			want:     utils.DefaultOperationTimeout,
		},
		{
			name:     "non-positive value falls back to the default",
			timeouts: withValues(types.StringValue("0s"), types.StringNull(), types.StringNull()),
			key:      "create",
			want:     utils.DefaultOperationTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := operationTimeout(context.Background(), tt.timeouts, tt.key); got != tt.want {
				t.Errorf("operationTimeout() = %v, want %v", got, tt.want)
			}
		})
	}
}

// stringOrNullValue converts a test fixture string into a Terraform value, mapping "" to null.
func stringOrNullValue(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}
