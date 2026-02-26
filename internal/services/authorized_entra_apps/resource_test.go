// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package authorized_entra_apps

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestDataSourceMetadata_AuthorizedEntraApps(t *testing.T) {
	d := NewAuthorizedEntraAppsDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_authorized_entra_apps"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestDataSourceMetadata_ManageableTenants(t *testing.T) {
	d := NewManageableTenantsDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_manageable_tenants"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestDataSourceSchema_AuthorizedEntraApps(t *testing.T) {
	d := NewAuthorizedEntraAppsDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	if _, ok := resp.Schema.Attributes["apps"]; !ok {
		t.Error("Schema missing apps attribute")
	}
}

func TestDataSourceSchema_ManageableTenants(t *testing.T) {
	d := NewManageableTenantsDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	if _, ok := resp.Schema.Attributes["tenants"]; !ok {
		t.Error("Schema missing tenants attribute")
	}
}

func TestResourceMetadata_AuthorizedEntraApp(t *testing.T) {
	r := NewAuthorizedEntraAppResource()
	req := resource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_authorized_entra_app"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestResourceSchema_AuthorizedEntraApp(t *testing.T) {
	r := NewAuthorizedEntraAppResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing id attribute")
	}
	if _, ok := resp.Schema.Attributes["aad_tenant_id"]; !ok {
		t.Error("Schema missing aad_tenant_id attribute")
	}
	if _, ok := resp.Schema.Attributes["app_id"]; !ok {
		t.Error("Schema missing app_id attribute")
	}
	if _, ok := resp.Schema.Attributes["is_admin_consent_granted"]; !ok {
		t.Error("Schema missing is_admin_consent_granted attribute")
	}
}
