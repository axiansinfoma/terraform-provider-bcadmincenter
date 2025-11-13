// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package quotas

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestQuotasDataSource_Metadata(t *testing.T) {
	d := NewQuotasDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_quotas"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestQuotasDataSource_Schema(t *testing.T) {
	d := NewQuotasDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing id attribute")
	}

	if _, ok := resp.Schema.Attributes["production_environments_quota"]; !ok {
		t.Error("Schema missing production_environments_quota attribute")
	}

	if _, ok := resp.Schema.Attributes["sandbox_environments_quota"]; !ok {
		t.Error("Schema missing sandbox_environments_quota attribute")
	}

	if _, ok := resp.Schema.Attributes["storage_quota_gb"]; !ok {
		t.Error("Schema missing storage_quota_gb attribute")
	}
}
