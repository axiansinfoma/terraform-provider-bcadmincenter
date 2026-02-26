// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentsettings

import (
	"context"
	"encoding/json"
	"github.com/vllni/terraform-provider-bcadmincenter/internal/constants"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// mockTokenCredential implements azcore.TokenCredential for testing.
type mockTokenCredential struct {
	token string
}

func (m *mockTokenCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token: m.token,
	}, nil
}

func TestService_GetUpdateSettings(t *testing.T) {
	tests := []struct {
		name              string
		responseBody      interface{}
		responseStatus    int
		wantErr           bool
		wantNil           bool
		checkStartTime    bool
		expectedStartTime string
	}{
		{
			name: "successful response with settings",
			responseBody: UpdateSettings{
				PreferredStartTime: strPtr("22:00"),
				PreferredEndTime:   strPtr("06:00"),
				TimeZoneID:         strPtr("Pacific Standard Time"),
			},
			responseStatus:    http.StatusOK,
			wantErr:           false,
			wantNil:           false,
			checkStartTime:    true,
			expectedStartTime: "22:00",
		},
		{
			name:           "no content response",
			responseBody:   nil,
			responseStatus: http.StatusNoContent,
			wantErr:        false,
			wantNil:        true,
		},
		{
			name:           "not found error",
			responseBody:   map[string]string{"error": "environment not found"},
			responseStatus: http.StatusNotFound,
			wantErr:        true,
			wantNil:        true, // Changed from false - when there's an error, result should be nil
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {

						t.Fatalf("Failed to encode response: %v", err)

					}
				}
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			result, err := svc.GetUpdateSettings(context.Background(), "BusinessCentral", "test-env")

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if (result == nil) != tt.wantNil {
				t.Errorf("result nil = %v, wantNil %v", result == nil, tt.wantNil)
			}

			if tt.checkStartTime && result != nil && result.PreferredStartTime != nil {
				if *result.PreferredStartTime != tt.expectedStartTime {
					t.Errorf("PreferredStartTime = %v, want %v", *result.PreferredStartTime, tt.expectedStartTime)
				}
			}
		})
	}
}

func TestService_SetUpdateSettings(t *testing.T) {
	tests := []struct {
		name           string
		inputSettings  *UpdateSettings
		responseBody   interface{}
		responseStatus int
		wantErr        bool
	}{
		{
			name: "successful update",
			inputSettings: &UpdateSettings{
				PreferredStartTime: strPtr("22:00"),
				PreferredEndTime:   strPtr("06:00"),
				TimeZoneID:         strPtr("Pacific Standard Time"),
			},
			responseBody: UpdateSettings{
				PreferredStartTime: strPtr("22:00"),
				PreferredEndTime:   strPtr("06:00"),
				TimeZoneID:         strPtr("Pacific Standard Time"),
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "invalid range error",
			inputSettings: &UpdateSettings{
				PreferredStartTime: strPtr("23:00"),
				PreferredEndTime:   strPtr("01:00"), // Too small window
			},
			responseBody:   map[string]string{"error": "invalid range"},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {

					t.Fatalf("Failed to encode response: %v", err)

				}
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			_, err := svc.SetUpdateSettings(context.Background(), "BusinessCentral", "test-env", tt.inputSettings)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetSecurityGroup(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantNil        bool
		expectedID     string
	}{
		{
			name: "successful response",
			responseBody: SecurityGroupResponse{
				ID:          "12345-67890",
				DisplayName: "Test Security Group",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        false,
			expectedID:     "12345-67890",
		},
		{
			name:           "no group configured",
			responseBody:   nil,
			responseStatus: http.StatusNoContent,
			wantErr:        false,
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {

						t.Fatalf("Failed to encode response: %v", err)

					}
				}
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			result, err := svc.GetSecurityGroup(context.Background(), "BusinessCentral", "test-env")

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if (result == nil) != tt.wantNil {
				t.Errorf("result nil = %v, wantNil %v", result == nil, tt.wantNil)
			}

			if !tt.wantNil && result != nil && result.ID != tt.expectedID {
				t.Errorf("ID = %v, want %v", result.ID, tt.expectedID)
			}
		})
	}
}

func TestService_SetAppInsightsKey(t *testing.T) {
	tests := []struct {
		name           string
		key            string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful set",
			key:            "InstrumentationKey=test-key",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "environment not active error",
			key:            "test-key",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			err := svc.SetAppInsightsKey(context.Background(), "BusinessCentral", "test-env", tt.key)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetAccessWithM365Licenses(t *testing.T) {
	tests := []struct {
		name            string
		responseBody    interface{}
		responseStatus  int
		wantErr         bool
		expectedEnabled bool
	}{
		{
			name: "access enabled",
			responseBody: AccessWithM365LicensesResponse{
				Enabled: true,
			},
			responseStatus:  http.StatusOK,
			wantErr:         false,
			expectedEnabled: true,
		},
		{
			name: "access disabled",
			responseBody: AccessWithM365LicensesResponse{
				Enabled: false,
			},
			responseStatus:  http.StatusOK,
			wantErr:         false,
			expectedEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {

					t.Fatalf("Failed to encode response: %v", err)

				}
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			result, err := svc.GetAccessWithM365Licenses(context.Background(), "BusinessCentral", "test-env")

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if result != nil && result.Enabled != tt.expectedEnabled {
				t.Errorf("Enabled = %v, want %v", result.Enabled, tt.expectedEnabled)
			}
		})
	}
}

// Helper function to create string pointers.
func strPtr(s string) *string {
	return &s
}
