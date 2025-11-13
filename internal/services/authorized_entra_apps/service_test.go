// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package authorized_entra_apps

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/constants"

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

func TestService_ListAuthorizedApps(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful response with apps",
			responseBody: []AuthorizedApp{
				{AppID: "app-1", IsAdminConsentGranted: true},
				{AppID: "app-2", IsAdminConsentGranted: false},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantCount:      2,
		},
		{
			name:           "successful response with empty list",
			responseBody:   []AuthorizedApp{},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantCount:      0,
		},
		{
			name:           "not found error",
			responseBody:   map[string]string{"error": "not found"},
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
		{
			name:           "unauthorized error",
			responseBody:   map[string]string{"error": "unauthorized"},
			responseStatus: http.StatusUnauthorized,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/admin/" + constants.DefaultAPIVersion + "/authorizedAadApps"
				if r.URL.Path != expectedPath {
					t.Errorf("unexpected path: %s, expected: %s", r.URL.Path, expectedPath)
				}
				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			apps, err := svc.ListAuthorizedApps(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListAuthorizedApps() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(apps) != tt.wantCount {
				t.Errorf("ListAuthorizedApps() returned %d apps, want %d", len(apps), tt.wantCount)
			}
		})
	}
}

func TestService_AuthorizeApp(t *testing.T) {
	tests := []struct {
		name           string
		appID          string
		responseBody   AuthorizedApp
		responseStatus int
		wantErr        bool
	}{
		{
			name:  "successful authorization",
			appID: "app-1",
			responseBody: AuthorizedApp{
				AppID:                 "app-1",
				IsAdminConsentGranted: false,
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:  "successful authorization with created status",
			appID: "app-2",
			responseBody: AuthorizedApp{
				AppID:                 "app-2",
				IsAdminConsentGranted: false,
			},
			responseStatus: http.StatusCreated,
			wantErr:        false,
		},
		{
			name:           "bad request error",
			appID:          "invalid",
			responseBody:   AuthorizedApp{},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("unexpected method: %s", r.Method)
				}
				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			app, err := svc.AuthorizeApp(context.Background(), tt.appID)

			if (err != nil) != tt.wantErr {
				t.Errorf("AuthorizeApp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && app.AppID != tt.appID {
				t.Errorf("AuthorizeApp() returned app with ID %s, want %s", app.AppID, tt.appID)
			}
		})
	}
}

func TestService_RemoveAuthorizedApp(t *testing.T) {
	tests := []struct {
		name           string
		appID          string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful removal with OK status",
			appID:          "app-1",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "successful removal with no content status",
			appID:          "app-2",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "not found error",
			appID:          "app-3",
			responseStatus: http.StatusNotFound,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("unexpected method: %s", r.Method)
				}
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
			err := svc.RemoveAuthorizedApp(context.Background(), tt.appID)

			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveAuthorizedApp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetManageableTenants(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   ManageableTenantsResponse
		responseStatus int
		wantErr        bool
		wantCount      int
	}{
		{
			name: "successful response with tenants",
			responseBody: ManageableTenantsResponse{
				Value: []ManageableTenant{
					{EntraTenantID: "tenant-1"},
					{EntraTenantID: "tenant-2"},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantCount:      2,
		},
		{
			name: "successful response with no tenants",
			responseBody: ManageableTenantsResponse{
				Value: []ManageableTenant{},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantCount:      0,
		},
		{
			name:           "unauthorized error",
			responseBody:   ManageableTenantsResponse{},
			responseStatus: http.StatusUnauthorized,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/admin/" + constants.DefaultAPIVersion + "/authorizedAadApps/manageableTenants"
				if r.URL.Path != expectedPath {
					t.Errorf("unexpected path: %s, expected: %s", r.URL.Path, expectedPath)
				}
				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			tenants, err := svc.GetManageableTenants(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetManageableTenants() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(tenants) != tt.wantCount {
				t.Errorf("GetManageableTenants() returned %d tenants, want %d", len(tenants), tt.wantCount)
			}
		})
	}
}
