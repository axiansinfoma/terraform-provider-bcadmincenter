// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package authorized_entra_apps

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
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

// TestServiceForTenant_TargetsConfiguredTenant covers the multi-tenant routing that this
// resource previously skipped entirely: it accepted aad_tenant_id, wrote it into state
// and into the resource ID, but built its service once from the provider client, so every
// API call went to the provider's own tenant. Terraform then reported success for an
// authorization that did not exist in the named tenant.
func TestServiceForTenant_TargetsConfiguredTenant(t *testing.T) {
	const providerTenant = "11111111-1111-1111-1111-111111111111"
	const otherTenant = "22222222-2222-2222-2222-222222222222"

	base, err := client.NewClient(context.Background(), &client.Config{
		TenantID:     providerTenant,
		ClientID:     "client-id",
		ClientSecret: "client-secret",
	})
	if err != nil {
		t.Fatalf("failed to build client: %v", err)
	}

	tests := []struct {
		name       string
		configured types.String
		wantTenant string
	}{
		{
			name:       "aad_tenant_id names another tenant",
			configured: types.StringValue(otherTenant),
			wantTenant: otherTenant,
		},
		{
			name:       "unset falls back to the provider tenant",
			configured: types.StringNull(),
			wantTenant: providerTenant,
		},
		{
			name:       "empty falls back to the provider tenant",
			configured: types.StringValue(""),
			wantTenant: providerTenant,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &AuthorizedEntraAppResource{client: base}

			gotTenant, svc := r.serviceForTenant(tt.configured)
			if gotTenant != tt.wantTenant {
				t.Errorf("tenant = %q, want %q", gotTenant, tt.wantTenant)
			}
			if svc == nil {
				t.Fatal("expected a service")
			}
			// The service must be bound to that tenant, so its requests carry a token
			// issued for it rather than for the provider's tenant.
			if got := svc.client.GetTenantID(); got != tt.wantTenant {
				t.Errorf("service client tenant = %q, want %q", got, tt.wantTenant)
			}
		})
	}
}
