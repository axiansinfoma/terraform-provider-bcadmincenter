// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package notificationrecipients

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

func TestNotificationRecipientIDRoundTrip(t *testing.T) {
	tenantID := "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d"
	recipID := "550e8400-e29b-41d4-a716-446655440000"

	id := BuildNotificationRecipientID(tenantID, recipID)
	parsedTenantID, parsedRecipID, err := ParseNotificationRecipientID(id)

	if err != nil {
		t.Fatalf("ParseNotificationRecipientID() unexpected error: %v", err)
	}

	if parsedTenantID != tenantID {
		t.Errorf("Round trip tenantID = %v, want %v", parsedTenantID, tenantID)
	}

	if parsedRecipID != recipID {
		t.Errorf("Round trip recipientID = %v, want %v", parsedRecipID, recipID)
	}
}
