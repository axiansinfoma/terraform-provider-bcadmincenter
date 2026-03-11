// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"testing"
)

func TestBuildPerTenantExtensionID(t *testing.T) {
	tests := []struct {
		name              string
		tenantID          string
		applicationFamily string
		environmentName   string
		appID             string
		want              string
	}{
		{
			name:              "basic PTE ID",
			tenantID:          "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			applicationFamily: "BusinessCentral",
			environmentName:   "Production",
			appID:             "d0e4c7e2-1234-5678-abcd-ef0123456789",
			want:              "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Production/perTenantExtensions/d0e4c7e2-1234-5678-abcd-ef0123456789",
		},
		{
			name:              "sandbox environment",
			tenantID:          "12345678-1234-1234-1234-123456789012",
			applicationFamily: "BusinessCentral",
			environmentName:   "Sandbox",
			appID:             "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
			want:              "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Sandbox/perTenantExtensions/aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildPerTenantExtensionID(tt.tenantID, tt.applicationFamily, tt.environmentName, tt.appID)
			if got != tt.want {
				t.Errorf("BuildPerTenantExtensionID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePerTenantExtensionID(t *testing.T) {
	tests := []struct {
		name              string
		id                string
		wantTenantID      string
		wantAppFamily     string
		wantEnvName       string
		wantAppID         string
		wantErr           bool
	}{
		{
			name:          "valid ID",
			id:            "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Production/perTenantExtensions/d0e4c7e2-1234-5678-abcd-ef0123456789",
			wantTenantID:  "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			wantAppFamily: "BusinessCentral",
			wantEnvName:   "Production",
			wantAppID:     "d0e4c7e2-1234-5678-abcd-ef0123456789",
			wantErr:       false,
		},
		{
			name:    "wrong segment count",
			id:      "/tenants/foo/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Production",
			wantErr: true,
		},
		{
			name:    "wrong provider namespace",
			id:      "/tenants/foo/providers/Microsoft.Other/applications/BusinessCentral/environments/Production/perTenantExtensions/bar",
			wantErr: true,
		},
		{
			name:    "missing perTenantExtensions segment",
			id:      "/tenants/foo/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Production/apps/bar",
			wantErr: true,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantID, appFamily, envName, appID, err := ParsePerTenantExtensionID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePerTenantExtensionID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tenantID != tt.wantTenantID {
					t.Errorf("tenantID = %v, want %v", tenantID, tt.wantTenantID)
				}
				if appFamily != tt.wantAppFamily {
					t.Errorf("appFamily = %v, want %v", appFamily, tt.wantAppFamily)
				}
				if envName != tt.wantEnvName {
					t.Errorf("envName = %v, want %v", envName, tt.wantEnvName)
				}
				if appID != tt.wantAppID {
					t.Errorf("appID = %v, want %v", appID, tt.wantAppID)
				}
			}
		})
	}
}

func TestBuildParseRoundTrip(t *testing.T) {
	tenantID := "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d"
	appFamily := "BusinessCentral"
	envName := "Production"
	appID := "d0e4c7e2-1234-5678-abcd-ef0123456789"

	id := BuildPerTenantExtensionID(tenantID, appFamily, envName, appID)

	gotTenantID, gotAppFamily, gotEnvName, gotAppID, err := ParsePerTenantExtensionID(id)
	if err != nil {
		t.Fatalf("ParsePerTenantExtensionID() error = %v", err)
	}

	if gotTenantID != tenantID {
		t.Errorf("round-trip tenantID = %v, want %v", gotTenantID, tenantID)
	}
	if gotAppFamily != appFamily {
		t.Errorf("round-trip appFamily = %v, want %v", gotAppFamily, appFamily)
	}
	if gotEnvName != envName {
		t.Errorf("round-trip envName = %v, want %v", gotEnvName, envName)
	}
	if gotAppID != appID {
		t.Errorf("round-trip appID = %v, want %v", gotAppID, appID)
	}
}
