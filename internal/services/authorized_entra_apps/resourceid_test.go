// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package authorized_entra_apps

import (
	"testing"
)

func TestBuildAuthorizedEntraAppID(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		appID    string
		want     string
	}{
		{
			name:     "valid app registration",
			tenantID: "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			appID:    "550e8400-e29b-41d4-a716-446655440000",
			want:     "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/authorizedEntraApps/550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "different tenant",
			tenantID: "12345678-1234-1234-1234-123456789012",
			appID:    "87654321-4321-4321-4321-210987654321",
			want:     "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/authorizedEntraApps/87654321-4321-4321-4321-210987654321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildAuthorizedEntraAppID(tt.tenantID, tt.appID)
			if got != tt.want {
				t.Errorf("BuildAuthorizedEntraAppID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAuthorizedEntraAppID(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		wantTenantID string
		wantAppID    string
		wantErr      bool
	}{
		{
			name:         "valid app ID",
			id:           "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/authorizedEntraApps/550e8400-e29b-41d4-a716-446655440000",
			wantTenantID: "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			wantAppID:    "550e8400-e29b-41d4-a716-446655440000",
			wantErr:      false,
		},
		{
			name:         "different valid ID",
			id:           "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/authorizedEntraApps/87654321-4321-4321-4321-210987654321",
			wantTenantID: "12345678-1234-1234-1234-123456789012",
			wantAppID:    "87654321-4321-4321-4321-210987654321",
			wantErr:      false,
		},
		{
			name:    "invalid format - missing parts",
			id:      "/tenants/tenant123/providers/Microsoft.Dynamics365.BusinessCentral",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong provider",
			id:      "/tenants/tenant123/providers/WrongProvider/authorizedEntraApps/app123",
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
			gotTenantID, gotAppID, err := ParseAuthorizedEntraAppID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAuthorizedEntraAppID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotTenantID != tt.wantTenantID {
					t.Errorf("ParseAuthorizedEntraAppID() tenantID = %v, want %v", gotTenantID, tt.wantTenantID)
				}
				if gotAppID != tt.wantAppID {
					t.Errorf("ParseAuthorizedEntraAppID() appID = %v, want %v", gotAppID, tt.wantAppID)
				}
			}
		})
	}
}

func TestAuthorizedEntraAppIDRoundTrip(t *testing.T) {
	tenantID := "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d"
	appID := "550e8400-e29b-41d4-a716-446655440000"

	id := BuildAuthorizedEntraAppID(tenantID, appID)
	parsedTenantID, parsedAppID, err := ParseAuthorizedEntraAppID(id)

	if err != nil {
		t.Fatalf("ParseAuthorizedEntraAppID() unexpected error: %v", err)
	}

	if parsedTenantID != tenantID {
		t.Errorf("Round trip tenantID = %v, want %v", parsedTenantID, tenantID)
	}

	if parsedAppID != appID {
		t.Errorf("Round trip appID = %v, want %v", parsedAppID, appID)
	}
}
