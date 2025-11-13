// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package timezones

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
	"github.com/vllni/terraform-provider-bcadmincenter/internal/constants"
)

// mockTokenCredential implements azcore.TokenCredential for testing
type mockTokenCredential struct {
	token string
}

func (m *mockTokenCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token: m.token,
	}, nil
}

func TestService_GetTimeZones(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		validateResult func(*testing.T, *TimeZoneResponse)
	}{
		{
			name: "successful response",
			responseBody: TimeZoneResponse{
				Value: []TimeZone{
					{
						ID:                      "Pacific Standard Time",
						DisplayName:             "(UTC-08:00) Pacific Time (US & Canada)",
						SupportsDaylightSavings: true,
						OffsetFromUTC:           "-08:00",
					},
					{
						ID:                      "Central European Standard Time",
						DisplayName:             "(UTC+01:00) Belgrade, Bratislava, Budapest, Ljubljana, Prague",
						SupportsDaylightSavings: true,
						OffsetFromUTC:           "+01:00",
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			validateResult: func(t *testing.T, result *TimeZoneResponse) {
				if len(result.Value) != 2 {
					t.Errorf("expected 2 timezones, got %d", len(result.Value))
				}
				if result.Value[0].ID != "Pacific Standard Time" {
					t.Errorf("unexpected timezone ID: %s", result.Value[0].ID)
				}
			},
		},
		{
			name:           "not found error",
			responseBody:   map[string]string{"error": "not found"},
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			// Create client with mock server
			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			// Test the method
			svc := NewService(c)
			result, err := svc.GetTimeZones(context.Background())

			// Assert results
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}
