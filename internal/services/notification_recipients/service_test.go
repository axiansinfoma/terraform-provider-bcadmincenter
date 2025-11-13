// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package notificationrecipients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
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

func TestService_List(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		expectedCount  int
	}{
		{
			name: "successful response with recipients",
			responseBody: NotificationRecipientsResponse{
				Value: []NotificationRecipient{
					{
						ID:    "00000000-0000-0000-0000-000000000001",
						Email: "admin1@example.com",
						Name:  "Admin One",
					},
					{
						ID:    "00000000-0000-0000-0000-000000000002",
						Email: "admin2@example.com",
						Name:  "Admin Two",
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedCount:  2,
		},
		{
			name: "successful response with no recipients",
			responseBody: NotificationRecipientsResponse{
				Value: []NotificationRecipient{},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedCount:  0,
		},
		{
			name:           "tenant not found error",
			responseBody:   map[string]string{"error": "tenant not found"},
			responseStatus: http.StatusNotFound,
			wantErr:        true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion("v2.24")
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			result, err := svc.List(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(result) != tt.expectedCount {
				t.Errorf("result count = %d, want %d", len(result), tt.expectedCount)
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	tests := []struct {
		name           string
		recipientID    string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantNil        bool
		expectedEmail  string
	}{
		{
			name:        "successful get",
			recipientID: "00000000-0000-0000-0000-000000000001",
			responseBody: NotificationRecipientsResponse{
				Value: []NotificationRecipient{
					{
						ID:    "00000000-0000-0000-0000-000000000001",
						Email: "admin1@example.com",
						Name:  "Admin One",
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        false,
			expectedEmail:  "admin1@example.com",
		},
		{
			name:        "recipient not found",
			recipientID: "00000000-0000-0000-0000-000000000999",
			responseBody: NotificationRecipientsResponse{
				Value: []NotificationRecipient{
					{
						ID:    "00000000-0000-0000-0000-000000000001",
						Email: "admin1@example.com",
						Name:  "Admin One",
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        true,
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion("v2.24")
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			result, err := svc.Get(context.Background(), tt.recipientID)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if (result == nil) != tt.wantNil {
				t.Errorf("result nil = %v, wantNil %v", result == nil, tt.wantNil)
			}

			if !tt.wantNil && result != nil {
				if result.Email != tt.expectedEmail {
					t.Errorf("Email = %v, want %v", result.Email, tt.expectedEmail)
				}
			}
		})
	}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		recipientName  string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
	}{
		{
			name:          "successful create",
			email:         "newadmin@example.com",
			recipientName: "New Admin",
			responseBody: NotificationRecipient{
				ID:    "00000000-0000-0000-0000-000000000003",
				Email: "newadmin@example.com",
				Name:  "New Admin",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "invalid input - empty email",
			email:          "",
			recipientName:  "New Admin",
			responseBody:   map[string]string{"error": "email can't be null or whitespace"},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "invalid input - empty name",
			email:          "newadmin@example.com",
			recipientName:  "",
			responseBody:   map[string]string{"error": "name can't be null or whitespace"},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "tenant not found",
			email:          "newadmin@example.com",
			recipientName:  "New Admin",
			responseBody:   map[string]string{"error": "tenant not found"},
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion("v2.24")
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			result, err := svc.Create(context.Background(), tt.email, tt.recipientName)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result == nil {
				t.Error("expected result, got nil")
			}

			if !tt.wantErr && result != nil {
				if result.Email != tt.email {
					t.Errorf("Email = %v, want %v", result.Email, tt.email)
				}
				if result.Name != tt.recipientName {
					t.Errorf("Name = %v, want %v", result.Name, tt.recipientName)
				}
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name           string
		recipientID    string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful delete",
			recipientID:    "00000000-0000-0000-0000-000000000001",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "successful delete with no content",
			recipientID:    "00000000-0000-0000-0000-000000000001",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "invalid input - empty guid",
			recipientID:    "00000000-0000-0000-0000-000000000000",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "tenant not found",
			recipientID:    "00000000-0000-0000-0000-000000000001",
			responseStatus: http.StatusNotFound,
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
			c.SetAPIVersion("v2.24")
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			err := svc.Delete(context.Background(), tt.recipientID)

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetNotificationSettings(t *testing.T) {
	tests := []struct {
		name             string
		responseBody     interface{}
		responseStatus   int
		wantErr          bool
		expectedTenantID string
		expectedCount    int
	}{
		{
			name: "successful response with recipients",
			responseBody: NotificationSettings{
				AADTenantID: "00000000-0000-0000-0000-000000000099",
				Recipients: []NotificationRecipient{
					{
						ID:    "00000000-0000-0000-0000-000000000001",
						Email: "admin1@example.com",
						Name:  "Admin One",
					},
					{
						ID:    "00000000-0000-0000-0000-000000000002",
						Email: "admin2@example.com",
						Name:  "Admin Two",
					},
				},
			},
			responseStatus:   http.StatusOK,
			wantErr:          false,
			expectedTenantID: "00000000-0000-0000-0000-000000000099",
			expectedCount:    2,
		},
		{
			name: "successful response with no recipients",
			responseBody: NotificationSettings{
				AADTenantID: "00000000-0000-0000-0000-000000000099",
				Recipients:  []NotificationRecipient{},
			},
			responseStatus:   http.StatusOK,
			wantErr:          false,
			expectedTenantID: "00000000-0000-0000-0000-000000000099",
			expectedCount:    0,
		},
		{
			name:           "tenant not found error",
			responseBody:   map[string]string{"error": "tenant not found"},
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if tt.responseBody != nil {
					json.NewEncoder(w).Encode(tt.responseBody)
				}
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion("v2.24")
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			result, err := svc.GetNotificationSettings(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && result != nil {
				if result.AADTenantID != tt.expectedTenantID {
					t.Errorf("AADTenantID = %v, want %v", result.AADTenantID, tt.expectedTenantID)
				}
				if len(result.Recipients) != tt.expectedCount {
					t.Errorf("Recipients count = %d, want %d", len(result.Recipients), tt.expectedCount)
				}
			}
		})
	}
}
