// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentapps

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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
			name: "success 200",
			responseBody: Operation{
				ID:     "op-123",
				Type:   "install",
				Status: OperationStatusScheduled,
			},
			responseStatus: http.StatusOK,
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
			name: "success 200",
			responseBody: Operation{
				ID:     "op-456",
				Type:   "update",
				Status: OperationStatusScheduled,
			},
			responseStatus: http.StatusOK,
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

func TestService_CancelUpdate(t *testing.T) {
	const testOpID = "scheduled-op-id-123"
	tests := []struct {
		name           string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "success 200",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "success 202",
			responseStatus: http.StatusAccepted,
			wantErr:        false,
		},
		{
			name:           "success 204",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "not allowed 400",
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "not allowed 409",
			responseStatus: http.StatusConflict,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("CancelUpdate() method = %v, want POST", r.Method)
				}
				if !strings.HasSuffix(r.URL.Path, "/update/cancel") {
					t.Errorf("CancelUpdate() path = %v, want .../update/cancel", r.URL.Path)
				}
				// Verify the ScheduledOperationId is sent in the body.
				var body CancelUpdateRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("CancelUpdate() could not decode request body: %v", err)
				}
				if body.ScheduledOperationID != testOpID {
					t.Errorf("CancelUpdate() ScheduledOperationId = %q, want %q", body.ScheduledOperationID, testOpID)
				}
				w.WriteHeader(tt.responseStatus)
			}))
			defer server.Close()

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			err := svc.CancelUpdate(context.Background(), "BusinessCentral", "my-env", "app-id-1", testOpID)

			if (err != nil) != tt.wantErr {
				t.Errorf("CancelUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetScheduledUpdateOperationID(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantID         string
		wantErr        bool
	}{
		{
			name: "returns scheduled update operation ID",
			responseBody: AppOperationsResponse{
				Value: []AppOperation{
					{ID: "op-scheduled-1", Status: "Scheduled", Type: "update"},
					{ID: "op-succeeded-1", Status: "Succeeded", Type: "update"},
				},
			},
			responseStatus: http.StatusOK,
			wantID:         "op-scheduled-1",
			wantErr:        false,
		},
		{
			name: "case insensitive status match",
			responseBody: AppOperationsResponse{
				Value: []AppOperation{
					{ID: "op-scheduled-2", Status: "scheduled", Type: "Update"},
				},
			},
			responseStatus: http.StatusOK,
			wantID:         "op-scheduled-2",
			wantErr:        false,
		},
		{
			name: "no scheduled update operation",
			responseBody: AppOperationsResponse{
				Value: []AppOperation{
					{ID: "op-running-1", Status: "Running", Type: "update"},
				},
			},
			responseStatus: http.StatusOK,
			wantID:         "",
			wantErr:        true,
		},
		{
			name:           "empty operations list",
			responseBody:   AppOperationsResponse{Value: []AppOperation{}},
			responseStatus: http.StatusOK,
			wantID:         "",
			wantErr:        true,
		},
		{
			name:           "api error",
			responseBody:   map[string]string{"code": "NotFound", "message": "not found"},
			responseStatus: http.StatusNotFound,
			wantID:         "",
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("GetScheduledUpdateOperationID() method = %v, want GET", r.Method)
				}
				if !strings.Contains(r.URL.Path, "/operations") {
					t.Errorf("GetScheduledUpdateOperationID() path = %v, want .../operations", r.URL.Path)
				}
				w.WriteHeader(tt.responseStatus)
				_ = json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			gotID, err := svc.GetScheduledUpdateOperationID(context.Background(), "BusinessCentral", "my-env", "app-id-1")

			if (err != nil) != tt.wantErr {
				t.Errorf("GetScheduledUpdateOperationID() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotID != tt.wantID {
				t.Errorf("GetScheduledUpdateOperationID() = %q, want %q", gotID, tt.wantID)
			}
		})
	}
}

func TestIsCancelNotAllowedError(t *testing.T) {
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
			name: "AdminCenterError",
			err:  &client.AdminCenterError{Code: "EntityValidationFailed", Message: "cannot be cancelled"},
			want: true,
		},
		{
			name: "plain error",
			err:  errors.New("network error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsCancelNotAllowedError(tt.err)
			if got != tt.want {
				t.Errorf("IsCancelNotAllowedError() = %v, want %v", got, tt.want)
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
			name: "success 200",
			responseBody: Operation{
				ID:     "op-789",
				Type:   "uninstall",
				Status: OperationStatusScheduled,
			},
			responseStatus: http.StatusOK,
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
