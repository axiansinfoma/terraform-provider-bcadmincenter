// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"testing"
)

func TestBuildEnvironmentID(t *testing.T) {
	tests := []struct {
		name              string
		tenantID          string
		applicationFamily string
		envName           string
		want              string
	}{
		{
			name:              "production environment",
			tenantID:          "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			applicationFamily: "BusinessCentral",
			envName:           "production",
			want:              "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production",
		},
		{
			name:              "sandbox environment",
			tenantID:          "12345678-1234-1234-1234-123456789012",
			applicationFamily: "BusinessCentral",
			envName:           "sandbox-dev",
			want:              "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/sandbox-dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildEnvironmentID(tt.tenantID, tt.applicationFamily, tt.envName)
			if got != tt.want {
				t.Errorf("BuildEnvironmentID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEnvironmentID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		wantTenantID  string
		wantAppFamily string
		wantEnvName   string
		wantErr       bool
	}{
		{
			name:          "valid production environment",
			id:            "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production",
			wantTenantID:  "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			wantAppFamily: "BusinessCentral",
			wantEnvName:   "production",
			wantErr:       false,
		},
		{
			name:          "valid sandbox environment",
			id:            "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/sandbox-dev",
			wantTenantID:  "12345678-1234-1234-1234-123456789012",
			wantAppFamily: "BusinessCentral",
			wantEnvName:   "sandbox-dev",
			wantErr:       false,
		},
		{
			name:    "invalid format - missing parts",
			id:      "/tenants/tenant123/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong provider",
			id:      "/tenants/tenant123/providers/WrongProvider/applications/BusinessCentral/environments/prod",
			wantErr: true,
		},
		{
			name:    "invalid format - empty",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTenantID, gotAppFamily, gotEnvName, err := ParseEnvironmentID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnvironmentID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotTenantID != tt.wantTenantID {
					t.Errorf("ParseEnvironmentID() tenantID = %v, want %v", gotTenantID, tt.wantTenantID)
				}
				if gotAppFamily != tt.wantAppFamily {
					t.Errorf("ParseEnvironmentID() applicationFamily = %v, want %v", gotAppFamily, tt.wantAppFamily)
				}
				if gotEnvName != tt.wantEnvName {
					t.Errorf("ParseEnvironmentID() environmentName = %v, want %v", gotEnvName, tt.wantEnvName)
				}
			}
		})
	}
}

func TestEnvironmentIDRoundTrip(t *testing.T) {
	tenantID := "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d"
	appFamily := "BusinessCentral"
	envName := "production"

	id := BuildEnvironmentID(tenantID, appFamily, envName)
	parsedTenantID, parsedAppFamily, parsedEnvName, err := ParseEnvironmentID(id)

	if err != nil {
		t.Fatalf("ParseEnvironmentID() unexpected error: %v", err)
	}

	if parsedTenantID != tenantID {
		t.Errorf("Round trip tenantID = %v, want %v", parsedTenantID, tenantID)
	}

	if parsedAppFamily != appFamily {
		t.Errorf("Round trip applicationFamily = %v, want %v", parsedAppFamily, appFamily)
	}

	if parsedEnvName != envName {
		t.Errorf("Round trip environmentName = %v, want %v", parsedEnvName, envName)
	}
}
