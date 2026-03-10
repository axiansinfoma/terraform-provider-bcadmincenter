// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentapps

import (
	"strings"
	"testing"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"
)

func TestBuildEnvironmentAppID(t *testing.T) {
	tests := []struct {
		name              string
		tenantID          string
		applicationFamily string
		environmentName   string
		appID             string
		want              string
	}{
		{
			name:              "basic app ID",
			tenantID:          "tenant-1",
			applicationFamily: "BusinessCentral",
			environmentName:   "production",
			appID:             "00000000-0000-0000-0000-000000000001",
			want:              "/tenants/tenant-1/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production/apps/00000000-0000-0000-0000-000000000001",
		},
		{
			name:              "sandbox environment app ID",
			tenantID:          "tenant-abc",
			applicationFamily: "BusinessCentral",
			environmentName:   "sandbox",
			appID:             "app-guid-xyz",
			want:              "/tenants/tenant-abc/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/sandbox/apps/app-guid-xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildEnvironmentAppID(tt.tenantID, tt.applicationFamily, tt.environmentName, tt.appID)
			if got != tt.want {
				t.Errorf("BuildEnvironmentAppID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEnvironmentAppID(t *testing.T) {
	tests := []struct {
		name                  string
		id                    string
		wantTenantID          string
		wantApplicationFamily string
		wantEnvironmentName   string
		wantAppID             string
		wantErr               bool
	}{
		{
			name:                  "valid ID",
			id:                    "/tenants/tenant-1/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production/apps/00000000-0000-0000-0000-000000000001",
			wantTenantID:          "tenant-1",
			wantApplicationFamily: "BusinessCentral",
			wantEnvironmentName:   "production",
			wantAppID:             "00000000-0000-0000-0000-000000000001",
			wantErr:               false,
		},
		{
			name:    "empty string",
			id:      "",
			wantErr: true,
		},
		{
			name:    "missing segments (8 parts instead of 10)",
			id:      "/tenants/tenant-1/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production",
			wantErr: true,
		},
		{
			name:    "wrong provider namespace",
			id:      "/tenants/tenant-1/providers/Microsoft.WrongNamespace/applications/BusinessCentral/environments/production/apps/app-id",
			wantErr: true,
		},
		{
			name:    "wrong apps segment",
			id:      "/tenants/tenant-1/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production/notapps/app-id",
			wantErr: true,
		},
		{
			name:    "wrong environments segment",
			id:      "/tenants/tenant-1/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/notenvironments/production/apps/app-id",
			wantErr: true,
		},
		{
			name:    "wrong applications segment",
			id:      "/tenants/tenant-1/providers/Microsoft.Dynamics365.BusinessCentral/notapplications/BusinessCentral/environments/production/apps/app-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tenantID, appFamily, envName, appID, err := ParseEnvironmentAppID(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnvironmentAppID() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if tenantID != tt.wantTenantID {
					t.Errorf("tenantID = %v, want %v", tenantID, tt.wantTenantID)
				}
				if appFamily != tt.wantApplicationFamily {
					t.Errorf("applicationFamily = %v, want %v", appFamily, tt.wantApplicationFamily)
				}
				if envName != tt.wantEnvironmentName {
					t.Errorf("environmentName = %v, want %v", envName, tt.wantEnvironmentName)
				}
				if appID != tt.wantAppID {
					t.Errorf("appID = %v, want %v", appID, tt.wantAppID)
				}
			}
		})
	}
}

func TestEnvironmentAppID_RoundTrip(t *testing.T) {
	tenantID := "00000000-0000-0000-0000-000000000001"
	appFamily := "BusinessCentral"
	envName := "my-sandbox"
	appID := "00000000-0000-0000-0000-000000000002"

	built := BuildEnvironmentAppID(tenantID, appFamily, envName, appID)

	parsedTenantID, parsedAppFamily, parsedEnvName, parsedAppID, err := ParseEnvironmentAppID(built)
	if err != nil {
		t.Fatalf("ParseEnvironmentAppID() error = %v", err)
	}

	if parsedTenantID != tenantID {
		t.Errorf("Round-trip tenantID = %v, want %v", parsedTenantID, tenantID)
	}
	if parsedAppFamily != appFamily {
		t.Errorf("Round-trip applicationFamily = %v, want %v", parsedAppFamily, appFamily)
	}
	if parsedEnvName != envName {
		t.Errorf("Round-trip environmentName = %v, want %v", parsedEnvName, envName)
	}
	if parsedAppID != appID {
		t.Errorf("Round-trip appID = %v, want %v", parsedAppID, appID)
	}

	// Verify provider namespace is embedded correctly.
	if !strings.Contains(built, constants.ProviderNamespace) {
		t.Errorf("Built ID %q does not contain provider namespace %q", built, constants.ProviderNamespace)
	}
}
