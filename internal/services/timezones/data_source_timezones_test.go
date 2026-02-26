// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package timezones

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestTimeZonesDataSource_Metadata(t *testing.T) {
	d := NewTimeZonesDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_timezones"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestTimeZonesDataSource_Schema(t *testing.T) {
	d := NewTimeZonesDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing id attribute")
	}

	if _, ok := resp.Schema.Attributes["timezones"]; !ok {
		t.Error("Schema missing timezones attribute")
	}
}
