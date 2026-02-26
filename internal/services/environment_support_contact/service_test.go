// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentsupportcontact

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
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

func TestService_Get(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantNil        bool
		expectedName   string
		expectedEmail  string
	}{
		{
			name: "successful response",
			responseBody: SupportContact{
				Name:  "Support Team",
				Email: "support@example.com",
				URL:   "https://support.example.com",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        false,
			expectedName:   "Support Team",
			expectedEmail:  "support@example.com",
		},
		{
			name:           "not found - no support contact configured",
			responseBody:   nil,
			responseStatus: http.StatusNotFound,
			wantErr:        false,
			wantNil:        true,
		},
		{
			name:           "environment not found error",
			responseBody:   map[string]string{"error": "environment not found"},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
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
			result, err := svc.Get(context.Background(), "BusinessCentral", "production")

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if (result == nil) != tt.wantNil {
				t.Errorf("result nil = %v, wantNil %v", result == nil, tt.wantNil)
			}

			if !tt.wantNil && result != nil {
				if result.Name != tt.expectedName {
					t.Errorf("Name = %v, want %v", result.Name, tt.expectedName)
				}
				if result.Email != tt.expectedEmail {
					t.Errorf("Email = %v, want %v", result.Email, tt.expectedEmail)
				}
			}
		})
	}
}

func TestService_Set(t *testing.T) {
	tests := []struct {
		name           string
		inputContact   *SupportContact
		responseBody   interface{}
		responseStatus int
		wantErr        bool
	}{
		{
			name: "successful set",
			inputContact: &SupportContact{
				Name:  "Support Team",
				Email: "support@example.com",
				URL:   "https://support.example.com",
			},
			responseBody: SupportContact{
				Name:  "Support Team",
				Email: "support@example.com",
				URL:   "https://support.example.com",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name: "environment not found error",
			inputContact: &SupportContact{
				Name:  "Support Team",
				Email: "support@example.com",
				URL:   "https://support.example.com",
			},
			responseBody:   map[string]string{"error": "environment not found"},
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
		{
			name: "request body required error",
			inputContact: &SupportContact{
				Name:  "",
				Email: "",
				URL:   "",
			},
			responseBody:   map[string]string{"error": "request body required"},
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
			_, err := svc.Set(context.Background(), "BusinessCentral", "production", tt.inputContact)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "resource not found code",
			err: &client.AdminCenterError{
				Code:    "ResourceNotFound",
				Message: "not found",
			},
			want: true,
		},
		{
			name: "wrapped environment not found code",
			err: fmt.Errorf("wrapped: %w", &client.AdminCenterError{
				Code:    "EnvironmentNotFound",
				Message: "not found",
			}),
			want: true,
		},
		{
			name: "fallback 404 message",
			err:  fmt.Errorf("API returned status 404: Not Found"),
			want: true,
		},
		{
			name: "different error",
			err: &client.AdminCenterError{
				Code:    "ValidationError",
				Message: "bad request",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNotFoundError(tt.err); got != tt.want {
				t.Errorf("isNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}
