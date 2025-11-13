// Copyright (c) Michael Villani
// SPDX-License-Identifier: MPL-2.0

package resourceid

import (
	"testing"
)

func TestBuildNotificationRecipientID(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		recipID  string
		want     string
	}{
		{
			name:     "valid IDs",
			tenantID: "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			recipID:  "550e8400-e29b-41d4-a716-446655440000",
			want:     "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:     "different tenant",
			tenantID: "12345678-1234-1234-1234-123456789012",
			recipID:  "87654321-4321-4321-4321-210987654321",
			want:     "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/87654321-4321-4321-4321-210987654321",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildNotificationRecipientID(tt.tenantID, tt.recipID)
			if got != tt.want {
				t.Errorf("BuildNotificationRecipientID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseNotificationRecipientID(t *testing.T) {
	tests := []struct {
		name         string
		id           string
		wantTenantID string
		wantRecipID  string
		wantErr      bool
	}{
		{
			name:         "valid ID",
			id:           "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/550e8400-e29b-41d4-a716-446655440000",
			wantTenantID: "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			wantRecipID:  "550e8400-e29b-41d4-a716-446655440000",
			wantErr:      false,
		},
		{
			name:         "different valid ID",
			id:           "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/notificationRecipients/87654321-4321-4321-4321-210987654321",
			wantTenantID: "12345678-1234-1234-1234-123456789012",
			wantRecipID:  "87654321-4321-4321-4321-210987654321",
			wantErr:      false,
		},
		{
			name:    "invalid format - missing parts",
			id:      "/tenants/tenant123/providers/Microsoft.Dynamics365.BusinessCentral",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong provider",
			id:      "/tenants/tenant123/providers/WrongProvider/notificationRecipients/recip123",
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
			gotTenantID, gotRecipID, err := ParseNotificationRecipientID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseNotificationRecipientID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotTenantID != tt.wantTenantID {
					t.Errorf("ParseNotificationRecipientID() tenantID = %v, want %v", gotTenantID, tt.wantTenantID)
				}
				if gotRecipID != tt.wantRecipID {
					t.Errorf("ParseNotificationRecipientID() recipientID = %v, want %v", gotRecipID, tt.wantRecipID)
				}
			}
		})
	}
}

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

func TestBuildEnvironmentSettingsID(t *testing.T) {
	tests := []struct {
		name              string
		tenantID          string
		applicationFamily string
		envName           string
		want              string
	}{
		{
			name:              "production settings",
			tenantID:          "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			applicationFamily: "BusinessCentral",
			envName:           "production",
			want:              "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production/settings",
		},
		{
			name:              "sandbox settings",
			tenantID:          "12345678-1234-1234-1234-123456789012",
			applicationFamily: "BusinessCentral",
			envName:           "sandbox-dev",
			want:              "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/sandbox-dev/settings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildEnvironmentSettingsID(tt.tenantID, tt.applicationFamily, tt.envName)
			if got != tt.want {
				t.Errorf("BuildEnvironmentSettingsID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEnvironmentSettingsID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		wantTenantID  string
		wantAppFamily string
		wantEnvName   string
		wantErr       bool
	}{
		{
			name:          "valid production settings",
			id:            "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production/settings",
			wantTenantID:  "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			wantAppFamily: "BusinessCentral",
			wantEnvName:   "production",
			wantErr:       false,
		},
		{
			name:          "valid sandbox settings",
			id:            "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/sandbox-dev/settings",
			wantTenantID:  "12345678-1234-1234-1234-123456789012",
			wantAppFamily: "BusinessCentral",
			wantEnvName:   "sandbox-dev",
			wantErr:       false,
		},
		{
			name:    "invalid format - missing settings",
			id:      "/tenants/tenant123/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/prod",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong provider",
			id:      "/tenants/tenant123/providers/WrongProvider/applications/BusinessCentral/environments/prod/settings",
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
			gotTenantID, gotAppFamily, gotEnvName, err := ParseEnvironmentSettingsID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnvironmentSettingsID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotTenantID != tt.wantTenantID {
					t.Errorf("ParseEnvironmentSettingsID() tenantID = %v, want %v", gotTenantID, tt.wantTenantID)
				}
				if gotAppFamily != tt.wantAppFamily {
					t.Errorf("ParseEnvironmentSettingsID() applicationFamily = %v, want %v", gotAppFamily, tt.wantAppFamily)
				}
				if gotEnvName != tt.wantEnvName {
					t.Errorf("ParseEnvironmentSettingsID() environmentName = %v, want %v", gotEnvName, tt.wantEnvName)
				}
			}
		})
	}
}

func TestBuildEnvironmentSupportContactID(t *testing.T) {
	tests := []struct {
		name              string
		tenantID          string
		applicationFamily string
		envName           string
		want              string
	}{
		{
			name:              "production support contact",
			tenantID:          "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			applicationFamily: "BusinessCentral",
			envName:           "Production",
			want:              "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Production/supportContact",
		},
		{
			name:              "sandbox support contact",
			tenantID:          "12345678-1234-1234-1234-123456789012",
			applicationFamily: "BusinessCentral",
			envName:           "Sandbox-Dev",
			want:              "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Sandbox-Dev/supportContact",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildEnvironmentSupportContactID(tt.tenantID, tt.applicationFamily, tt.envName)
			if got != tt.want {
				t.Errorf("BuildEnvironmentSupportContactID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseEnvironmentSupportContactID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		wantTenantID  string
		wantAppFamily string
		wantEnvName   string
		wantErr       bool
	}{
		{
			name:          "valid production support contact",
			id:            "/tenants/9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Production/supportContact",
			wantTenantID:  "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d",
			wantAppFamily: "BusinessCentral",
			wantEnvName:   "Production",
			wantErr:       false,
		},
		{
			name:          "valid sandbox support contact",
			id:            "/tenants/12345678-1234-1234-1234-123456789012/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Sandbox-Dev/supportContact",
			wantTenantID:  "12345678-1234-1234-1234-123456789012",
			wantAppFamily: "BusinessCentral",
			wantEnvName:   "Sandbox-Dev",
			wantErr:       false,
		},
		{
			name:    "invalid format - missing supportContact",
			id:      "/tenants/tenant123/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Production",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong provider",
			id:      "/tenants/tenant123/providers/WrongProvider/applications/BusinessCentral/environments/Production/supportContact",
			wantErr: true,
		},
		{
			name:    "invalid format - empty",
			id:      "",
			wantErr: true,
		},
		{
			name:    "invalid format - wrong suffix",
			id:      "/tenants/tenant123/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/Production/wrongSuffix",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotTenantID, gotAppFamily, gotEnvName, err := ParseEnvironmentSupportContactID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEnvironmentSupportContactID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotTenantID != tt.wantTenantID {
					t.Errorf("ParseEnvironmentSupportContactID() tenantID = %v, want %v", gotTenantID, tt.wantTenantID)
				}
				if gotAppFamily != tt.wantAppFamily {
					t.Errorf("ParseEnvironmentSupportContactID() applicationFamily = %v, want %v", gotAppFamily, tt.wantAppFamily)
				}
				if gotEnvName != tt.wantEnvName {
					t.Errorf("ParseEnvironmentSupportContactID() environmentName = %v, want %v", gotEnvName, tt.wantEnvName)
				}
			}
		})
	}
}

// Test round-trip conversions (build -> parse -> build should yield same result)
func TestRoundTrip_NotificationRecipient(t *testing.T) {
	tenantID := "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d"
	recipID := "550e8400-e29b-41d4-a716-446655440000"

	// Build ID
	id := BuildNotificationRecipientID(tenantID, recipID)

	// Parse ID
	parsedTenantID, parsedRecipID, err := ParseNotificationRecipientID(id)
	if err != nil {
		t.Fatalf("ParseNotificationRecipientID() error = %v", err)
	}

	// Verify parsed values match originals
	if parsedTenantID != tenantID {
		t.Errorf("Round trip tenantID: got %v, want %v", parsedTenantID, tenantID)
	}
	if parsedRecipID != recipID {
		t.Errorf("Round trip recipientID: got %v, want %v", parsedRecipID, recipID)
	}

	// Build ID again and verify it matches original
	id2 := BuildNotificationRecipientID(parsedTenantID, parsedRecipID)
	if id != id2 {
		t.Errorf("Round trip ID: got %v, want %v", id2, id)
	}
}

func TestRoundTrip_Environment(t *testing.T) {
	tenantID := "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d"
	appFamily := "BusinessCentral"
	envName := "production"

	// Build ID
	id := BuildEnvironmentID(tenantID, appFamily, envName)

	// Parse ID
	parsedTenantID, parsedAppFamily, parsedEnvName, err := ParseEnvironmentID(id)
	if err != nil {
		t.Fatalf("ParseEnvironmentID() error = %v", err)
	}

	// Verify parsed values match originals
	if parsedTenantID != tenantID {
		t.Errorf("Round trip tenantID: got %v, want %v", parsedTenantID, tenantID)
	}
	if parsedAppFamily != appFamily {
		t.Errorf("Round trip applicationFamily: got %v, want %v", parsedAppFamily, appFamily)
	}
	if parsedEnvName != envName {
		t.Errorf("Round trip environmentName: got %v, want %v", parsedEnvName, envName)
	}

	// Build ID again and verify it matches original
	id2 := BuildEnvironmentID(parsedTenantID, parsedAppFamily, parsedEnvName)
	if id != id2 {
		t.Errorf("Round trip ID: got %v, want %v", id2, id)
	}
}

func TestRoundTrip_EnvironmentSettings(t *testing.T) {
	tenantID := "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d"
	appFamily := "BusinessCentral"
	envName := "production"

	// Build ID
	id := BuildEnvironmentSettingsID(tenantID, appFamily, envName)

	// Parse ID
	parsedTenantID, parsedAppFamily, parsedEnvName, err := ParseEnvironmentSettingsID(id)
	if err != nil {
		t.Fatalf("ParseEnvironmentSettingsID() error = %v", err)
	}

	// Verify parsed values match originals
	if parsedTenantID != tenantID {
		t.Errorf("Round trip tenantID: got %v, want %v", parsedTenantID, tenantID)
	}
	if parsedAppFamily != appFamily {
		t.Errorf("Round trip applicationFamily: got %v, want %v", parsedAppFamily, appFamily)
	}
	if parsedEnvName != envName {
		t.Errorf("Round trip environmentName: got %v, want %v", parsedEnvName, envName)
	}

	// Build ID again and verify it matches original
	id2 := BuildEnvironmentSettingsID(parsedTenantID, parsedAppFamily, parsedEnvName)
	if id != id2 {
		t.Errorf("Round trip ID: got %v, want %v", id2, id)
	}
}

func TestRoundTrip_EnvironmentSupportContact(t *testing.T) {
	tenantID := "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d"
	appFamily := "BusinessCentral"
	envName := "Production"

	// Build ID
	id := BuildEnvironmentSupportContactID(tenantID, appFamily, envName)

	// Parse ID
	parsedTenantID, parsedAppFamily, parsedEnvName, err := ParseEnvironmentSupportContactID(id)
	if err != nil {
		t.Fatalf("ParseEnvironmentSupportContactID() error = %v", err)
	}

	// Verify parsed values match originals
	if parsedTenantID != tenantID {
		t.Errorf("Round trip tenantID: got %v, want %v", parsedTenantID, tenantID)
	}
	if parsedAppFamily != appFamily {
		t.Errorf("Round trip applicationFamily: got %v, want %v", parsedAppFamily, appFamily)
	}
	if parsedEnvName != envName {
		t.Errorf("Round trip environmentName: got %v, want %v", parsedEnvName, envName)
	}

	// Build ID again and verify it matches original
	id2 := BuildEnvironmentSupportContactID(parsedTenantID, parsedAppFamily, parsedEnvName)
	if id != id2 {
		t.Errorf("Round trip ID: got %v, want %v", id2, id)
	}
}
