// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentapps

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"
)

// mockTokenCredential is a mock implementation of azcore.TokenCredential for testing.
type mockTokenCredential struct {
	token string
}

func (m *mockTokenCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{
		Token: m.token,
	}, nil
}

func newTestClient(t *testing.T, serverURL string) *client.Client {
	t.Helper()
	mockCred := &mockTokenCredential{token: "test-token"}
	c := &client.Client{}
	c.SetCredential(mockCred)
	c.SetBaseURL(serverURL)
	c.SetAPIVersion(constants.DefaultAPIVersion)
	c.SetHTTPClient(&http.Client{})
	return c
}

func TestService_GetByID(t *testing.T) {
	tests := []struct {
		name           string
		appID          string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantNil        bool
	}{
		{
			name:  "app found",
			appID: "app-id-1",
			responseBody: AppListResponse{
				Value: []App{
					{ID: "app-id-1", Name: "My App", Publisher: "Contoso", Version: "1.0.0.0", Status: AppStatusInstalled},
					{ID: "app-id-2", Name: "Other App", Publisher: "Contoso", Version: "2.0.0.0", Status: AppStatusInstalled},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        false,
		},
		{
			name:  "app not found (empty list)",
			appID: "missing-app-id",
			responseBody: AppListResponse{
				Value: []App{},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        true,
		},
		{
			name:  "app not found (not in list)",
			appID: "missing-app-id",
			responseBody: AppListResponse{
				Value: []App{
					{ID: "other-app-id", Name: "Other App", Publisher: "Contoso", Version: "1.0.0.0", Status: AppStatusInstalled},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        true,
		},
		{
			name:           "HTTP error",
			appID:          "app-id-1",
			responseBody:   map[string]string{"error": "internal server error"},
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
			wantNil:        true,
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

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			app, err := svc.GetByID(context.Background(), "BusinessCentral", "my-env", tt.appID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (app == nil) != tt.wantNil {
				t.Errorf("GetByID() app == nil is %v, wantNil %v", app == nil, tt.wantNil)
			}
			if app != nil && app.ID != tt.appID {
				t.Errorf("GetByID() app.ID = %v, want %v", app.ID, tt.appID)
			}
		})
	}
}

func TestService_Install(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
	}{
		{
			name: "success 202",
			responseBody: Operation{
				ID:     "op-123",
				Type:   "AppInstall",
				Status: OperationStatusRunning,
			},
			responseStatus: http.StatusAccepted,
			wantErr:        false,
		},
		{
			name:           "unexpected status 400",
			responseBody:   map[string]string{"error": "bad request"},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "unexpected status 500",
			responseBody:   map[string]string{"error": "internal server error"},
			responseStatus: http.StatusInternalServerError,
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

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			req := &InstallAppRequest{
				AllowPreviewVersion:               false,
				InstallOrUpdateNeededDependencies: true,
			}
			op, err := svc.Install(context.Background(), "BusinessCentral", "my-env", "app-id-1", req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && op == nil {
				t.Error("Install() returned nil operation on success")
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
	}{
		{
			name: "success 202",
			responseBody: Operation{
				ID:     "op-456",
				Type:   "AppUpdate",
				Status: OperationStatusRunning,
			},
			responseStatus: http.StatusAccepted,
			wantErr:        false,
		},
		{
			name:           "unexpected status 400",
			responseBody:   map[string]string{"error": "bad request"},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "unexpected status 500",
			responseBody:   map[string]string{"error": "internal server error"},
			responseStatus: http.StatusInternalServerError,
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

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			req := &UpdateAppRequest{
				TargetVersion:                     "2.0.0.0",
				AllowPreviewVersion:               false,
				InstallOrUpdateNeededDependencies: true,
			}
			op, err := svc.Update(context.Background(), "BusinessCentral", "my-env", "app-id-1", req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && op == nil {
				t.Error("Update() returned nil operation on success")
			}
		})
	}
}

func TestService_Uninstall(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
	}{
		{
			name: "success 202",
			responseBody: Operation{
				ID:     "op-789",
				Type:   "AppUninstall",
				Status: OperationStatusRunning,
			},
			responseStatus: http.StatusAccepted,
			wantErr:        false,
		},
		{
			name:           "unexpected status 400",
			responseBody:   map[string]string{"error": "bad request"},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "unexpected status 500",
			responseBody:   map[string]string{"error": "internal server error"},
			responseStatus: http.StatusInternalServerError,
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

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			req := &UninstallAppRequest{
				DoNotSaveData:       false,
				UninstallDependents: false,
			}
			op, err := svc.Uninstall(context.Background(), "BusinessCentral", "my-env", "app-id-1", req)

			if (err != nil) != tt.wantErr {
				t.Errorf("Uninstall() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && op == nil {
				t.Error("Uninstall() returned nil operation on success")
			}
		})
	}
}
