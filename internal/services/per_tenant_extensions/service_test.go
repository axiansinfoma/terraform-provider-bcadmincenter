// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
)

// mockTokenCredential is a mock implementation of azcore.TokenCredential for testing.
type mockTokenCredential struct {
	token string
}

func (m *mockTokenCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: m.token}, nil
}

func newTestClient(t *testing.T, serverURL string) *client.Client {
	t.Helper()
	mockCred := &mockTokenCredential{token: "test-token"}
	c := &client.Client{}
	c.SetCredential(mockCred)
	c.SetBaseURL(serverURL)
	c.SetAPIVersion("v2.27")
	c.SetHTTPClient(&http.Client{})
	return c
}

func TestService_GetFirstCompany(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantCompanyID  string
	}{
		{
			name: "returns first company",
			responseBody: CompanyListResponse{
				Value: []Company{
					{ID: "company-1", Name: "CRONUS"},
					{ID: "company-2", Name: "Other"},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantCompanyID:  "company-1",
		},
		{
			name: "no companies",
			responseBody: CompanyListResponse{
				Value: []Company{},
			},
			responseStatus: http.StatusOK,
			wantErr:        true,
		},
		{
			name:           "HTTP error",
			responseBody:   map[string]interface{}{"code": "InternalError", "message": "internal error"},
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				_ = json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			got, err := svc.GetFirstCompany(context.Background(), "Production")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFirstCompany() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantCompanyID {
				t.Errorf("GetFirstCompany() = %v, want %v", got, tt.wantCompanyID)
			}
		})
	}
}

func TestService_CreateExtensionUpload(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantSystemID   string
	}{
		{
			name: "successful upload record creation",
			responseBody: ExtensionUpload{
				SystemID: "upload-system-id-123",
			},
			responseStatus: http.StatusCreated,
			wantErr:        false,
			wantSystemID:   "upload-system-id-123",
		},
		{
			name: "response with 200 OK",
			responseBody: ExtensionUpload{
				SystemID: "upload-system-id-456",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantSystemID:   "upload-system-id-456",
		},
		{
			name:           "server error",
			responseBody:   map[string]interface{}{"code": "BadRequest", "message": "bad request"},
			responseStatus: http.StatusBadRequest,
			wantErr:        true,
		},
		{
			name:           "missing system ID in response",
			responseBody:   ExtensionUpload{SystemID: ""},
			responseStatus: http.StatusCreated,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				_ = json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			req := &ExtensionUploadRequest{
				Schedule:       DefaultSchedule,
				SchemaSyncMode: DefaultSchemaSyncMode,
			}

			got, err := svc.CreateExtensionUpload(context.Background(), "Production", "company-1", req)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateExtensionUpload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.wantSystemID {
				t.Errorf("CreateExtensionUpload() = %v, want %v", got, tt.wantSystemID)
			}
		})
	}
}

func TestService_UploadExtensionContent(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful content upload",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "no content response (204)",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "server error",
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify required headers.
				if r.Header.Get("Content-Type") != "application/octet-stream" {
					t.Errorf("Content-Type = %v, want application/octet-stream", r.Header.Get("Content-Type"))
				}
				if r.Header.Get("If-Match") != "*" {
					t.Errorf("If-Match = %v, want *", r.Header.Get("If-Match"))
				}

				if tt.responseStatus >= 400 {
					w.WriteHeader(tt.responseStatus)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "Error", "message": "error"})
				} else {
					w.WriteHeader(tt.responseStatus)
				}
			}))
			defer server.Close()

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			err := svc.UploadExtensionContent(context.Background(), "Production", "company-1", "upload-id", []byte("fake-app-bytes"))
			if (err != nil) != tt.wantErr {
				t.Errorf("UploadExtensionContent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_TriggerInstall(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful trigger",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "200 OK trigger",
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "server error",
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.responseStatus >= 400 {
					w.WriteHeader(tt.responseStatus)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "Error", "message": "error"})
				} else {
					w.WriteHeader(tt.responseStatus)
				}
			}))
			defer server.Close()

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			err := svc.TriggerInstall(context.Background(), "Production", "company-1", "upload-id")
			if (err != nil) != tt.wantErr {
				t.Errorf("TriggerInstall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_GetExtensionByPackageID(t *testing.T) {
	tests := []struct {
		name           string
		packageID      string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantNil        bool
	}{
		{
			name:      "extension found",
			packageID: "pkg-id-1",
			responseBody: ExtensionListResponse{
				Value: []Extension{
					{PackageID: "pkg-id-1", ID: "app-id-1", DisplayName: "My Extension", Publisher: "Contoso", IsInstalled: true},
					{PackageID: "pkg-id-2", ID: "app-id-2", DisplayName: "Other Extension", Publisher: "Contoso", IsInstalled: true},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        false,
		},
		{
			name:      "extension not found",
			packageID: "missing-pkg",
			responseBody: ExtensionListResponse{
				Value: []Extension{
					{PackageID: "other-pkg", ID: "app-id-1", DisplayName: "Other", Publisher: "Contoso", IsInstalled: true},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        true,
		},
		{
			name:           "HTTP error",
			packageID:      "pkg-id-1",
			responseBody:   map[string]interface{}{"code": "Error", "message": "error"},
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				_ = json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			ext, err := svc.GetExtensionByPackageID(context.Background(), "Production", "company-1", tt.packageID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetExtensionByPackageID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (ext == nil) != tt.wantNil {
				t.Errorf("GetExtensionByPackageID() nil = %v, want nil = %v", ext == nil, tt.wantNil)
			}
		})
	}
}

func TestService_GetInstalledExtensionByNameAndPublisher(t *testing.T) {
	tests := []struct {
		name           string
		extName        string
		publisher      string
		responseBody   interface{}
		responseStatus int
		wantErr        bool
		wantNil        bool
		wantPackageID  string
	}{
		{
			name:      "extension found",
			extName:   "My Extension",
			publisher: "Contoso",
			responseBody: ExtensionListResponse{
				Value: []Extension{
					{PackageID: "pkg-id-1", ID: "app-id-1", DisplayName: "My Extension", Publisher: "Contoso", IsInstalled: true},
					{PackageID: "pkg-id-2", ID: "app-id-2", DisplayName: "Other Extension", Publisher: "Contoso", IsInstalled: true},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        false,
			wantPackageID:  "pkg-id-1",
		},
		{
			name:      "skips uninstalled extension with same name",
			extName:   "My Extension",
			publisher: "Contoso",
			responseBody: ExtensionListResponse{
				Value: []Extension{
					{PackageID: "old-pkg", ID: "app-id-1", DisplayName: "My Extension", Publisher: "Contoso", IsInstalled: false},
					{PackageID: "new-pkg", ID: "app-id-1", DisplayName: "My Extension", Publisher: "Contoso", IsInstalled: true},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        false,
			wantPackageID:  "new-pkg",
		},
		{
			name:      "extension not found",
			extName:   "Missing Extension",
			publisher: "Contoso",
			responseBody: ExtensionListResponse{
				Value: []Extension{
					{PackageID: "pkg-id-1", ID: "app-id-1", DisplayName: "Other Extension", Publisher: "Contoso", IsInstalled: true},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			wantNil:        true,
		},
		{
			name:           "HTTP error",
			extName:        "My Extension",
			publisher:      "Contoso",
			responseBody:   map[string]interface{}{"code": "Error", "message": "error"},
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				_ = json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			ext, err := svc.GetInstalledExtensionByNameAndPublisher(context.Background(), "Production", "company-1", tt.extName, tt.publisher)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetInstalledExtensionByNameAndPublisher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (ext == nil) != tt.wantNil {
				t.Errorf("GetInstalledExtensionByNameAndPublisher() nil = %v, want nil = %v", ext == nil, tt.wantNil)
				return
			}
			if !tt.wantNil && ext.PackageID != tt.wantPackageID {
				t.Errorf("GetInstalledExtensionByNameAndPublisher() PackageID = %v, want %v", ext.PackageID, tt.wantPackageID)
			}
		})
	}
}

func TestService_Uninstall(t *testing.T) {
	tests := []struct {
		name           string
		deleteData     bool
		responseStatus int
		wantErr        bool
		wantPathSuffix string
	}{
		{
			name:           "uninstall without data deletion",
			deleteData:     false,
			responseStatus: http.StatusNoContent,
			wantErr:        false,
			wantPathSuffix: "/Microsoft.NAV.uninstall",
		},
		{
			name:           "uninstall with data deletion",
			deleteData:     true,
			responseStatus: http.StatusNoContent,
			wantErr:        false,
			wantPathSuffix: "/Microsoft.NAV.uninstallAndDeleteExtensionData",
		},
		{
			name:           "server error",
			deleteData:     false,
			responseStatus: http.StatusInternalServerError,
			wantErr:        true,
			wantPathSuffix: "/Microsoft.NAV.uninstall",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedPath string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				capturedPath = r.URL.Path
				if tt.responseStatus >= 400 {
					w.WriteHeader(tt.responseStatus)
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "Error", "message": "error"})
				} else {
					w.WriteHeader(tt.responseStatus)
				}
			}))
			defer server.Close()

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			err := svc.Uninstall(context.Background(), "Production", "company-1", "pkg-id-1", tt.deleteData)
			if (err != nil) != tt.wantErr {
				t.Errorf("Uninstall() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if len(capturedPath) < len(tt.wantPathSuffix) || capturedPath[len(capturedPath)-len(tt.wantPathSuffix):] != tt.wantPathSuffix {
					t.Errorf("URL path suffix = %v, want suffix %v", capturedPath, tt.wantPathSuffix)
				}
			}
		})
	}
}

func TestService_Unpublish(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		wantErr        bool
	}{
		{
			name:           "successful unpublish",
			responseStatus: http.StatusNoContent,
			wantErr:        false,
		},
		{
			name:           "404 gracefully ignored",
			responseStatus: http.StatusNotFound,
			wantErr:        false,
		},
		{
			name:           "405 gracefully ignored",
			responseStatus: http.StatusMethodNotAllowed,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseStatus)
				if tt.responseStatus >= 400 {
					_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": "NotFound", "message": "not found"})
				}
			}))
			defer server.Close()

			c := newTestClient(t, server.URL)
			svc := NewService(c)

			err := svc.Unpublish(context.Background(), "Production", "company-1", "pkg-id-1")
			if (err != nil) != tt.wantErr {
				t.Errorf("Unpublish() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_WaitForDeployment_StaleEntryFiltered(t *testing.T) {
	// A "Failed" status entry that started BEFORE notBefore must be ignored.
	// WaitForDeployment should return nil (keep waiting) until a fresh entry appears.
	staleTime := time.Now().Add(-60 * time.Second)
	freshTime := time.Now().Add(-1 * time.Second) // fresh entry starts 1 s ago

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var entries []ExtensionDeploymentStatus
		if callCount == 1 {
			// First poll: only a stale failed entry exists.
			entries = []ExtensionDeploymentStatus{
				{Name: "MyPTE", Publisher: "Me", OperationType: "Upload", Status: DeploymentStatusFailed, StartedOn: staleTime.UTC().Format(time.RFC3339)},
			}
		} else {
			// Second poll: fresh completed entry appears.
			entries = []ExtensionDeploymentStatus{
				{Name: "MyPTE", Publisher: "Me", OperationType: "Upload", Status: DeploymentStatusCompleted, StartedOn: freshTime.UTC().Format(time.RFC3339)},
				{Name: "MyPTE", Publisher: "Me", OperationType: "Upload", Status: DeploymentStatusFailed, StartedOn: staleTime.UTC().Format(time.RFC3339)},
			}
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ExtensionDeploymentStatusListResponse{Value: entries})
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	svc := NewService(c)

	// notBefore is the zero time for staleTime but after it; any entry before notBefore is stale.
	notBefore := staleTime.Add(30 * time.Second) // between stale and fresh

	// Use a very short initial delay / ticker for the test by overriding the timeout.
	// We can't easily shorten the 5 s initial delay in tests without refactoring, but we
	// can use a short timeout and just verify the stale-filter behaviour via the low-level helper.
	result, err := svc.getLatestDeploymentStatus(context.Background(), "Production", "company-1", notBefore)
	if err != nil {
		t.Fatalf("getLatestDeploymentStatus() error = %v", err)
	}
	if result != nil {
		t.Errorf("getLatestDeploymentStatus() expected nil (stale entry filtered), got status=%s startedOn=%s", result.Status, result.StartedOn)
	}
}

func TestService_WaitForDeployment_FreshCompletedEntry(t *testing.T) {
	// A "Completed" entry started AFTER notBefore should be returned immediately.
	freshTime := time.Now().Add(-1 * time.Second)
	notBefore := freshTime.Add(-2 * time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries := []ExtensionDeploymentStatus{
			{Name: "MyPTE", Publisher: "Me", OperationType: "Upload", Status: DeploymentStatusCompleted, StartedOn: freshTime.UTC().Format(time.RFC3339)},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ExtensionDeploymentStatusListResponse{Value: entries})
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	svc := NewService(c)

	result, err := svc.getLatestDeploymentStatus(context.Background(), "Production", "company-1", notBefore)
	if err != nil {
		t.Fatalf("getLatestDeploymentStatus() error = %v", err)
	}
	if result == nil {
		t.Fatal("getLatestDeploymentStatus() expected non-nil result for fresh completed entry")
	}
	if result.Status != DeploymentStatusCompleted {
		t.Errorf("getLatestDeploymentStatus() status = %v, want %v", result.Status, DeploymentStatusCompleted)
	}
}

func TestService_WaitForDeployment_FreshFailedEntry(t *testing.T) {
	// A "Failed" entry started AFTER notBefore should be returned (not filtered).
	freshTime := time.Now().Add(-1 * time.Second)
	notBefore := freshTime.Add(-2 * time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries := []ExtensionDeploymentStatus{
			{Name: "MyPTE", Publisher: "Me", OperationType: "Upload", Status: DeploymentStatusFailed, StartedOn: freshTime.UTC().Format(time.RFC3339)},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ExtensionDeploymentStatusListResponse{Value: entries})
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	svc := NewService(c)

	result, err := svc.getLatestDeploymentStatus(context.Background(), "Production", "company-1", notBefore)
	if err != nil {
		t.Fatalf("getLatestDeploymentStatus() error = %v", err)
	}
	if result == nil {
		t.Fatal("getLatestDeploymentStatus() expected non-nil result for fresh failed entry")
	}
	if result.Status != DeploymentStatusFailed {
		t.Errorf("getLatestDeploymentStatus() status = %v, want %v", result.Status, DeploymentStatusFailed)
	}
}

func TestService_WaitForDeployment_EmptyList(t *testing.T) {
	// An empty deployment status list should return (nil, nil) — keep waiting.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ExtensionDeploymentStatusListResponse{Value: []ExtensionDeploymentStatus{}})
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	svc := NewService(c)

	result, err := svc.getLatestDeploymentStatus(context.Background(), "Production", "company-1", time.Now().Add(-5*time.Second))
	if err != nil {
		t.Fatalf("getLatestDeploymentStatus() error = %v", err)
	}
	if result != nil {
		t.Errorf("getLatestDeploymentStatus() expected nil for empty list, got %+v", result)
	}
}

func TestService_WaitForDeployment_UnparsableTimestampSkipped(t *testing.T) {
	// An entry with an unparseable StartedOn timestamp should be skipped conservatively.
	notBefore := time.Now().Add(-60 * time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		entries := []ExtensionDeploymentStatus{
			{Name: "MyPTE", Publisher: "Me", OperationType: "Upload", Status: DeploymentStatusCompleted, StartedOn: "not-a-timestamp"},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(ExtensionDeploymentStatusListResponse{Value: entries})
	}))
	defer server.Close()

	c := newTestClient(t, server.URL)
	svc := NewService(c)

	result, err := svc.getLatestDeploymentStatus(context.Background(), "Production", "company-1", notBefore)
	if err != nil {
		t.Fatalf("getLatestDeploymentStatus() error = %v", err)
	}
	if result != nil {
		t.Errorf("getLatestDeploymentStatus() expected nil for unparseable timestamp (conservative skip), got %+v", result)
	}
}
