// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"
)

const (
	testAppFamily = "BusinessCentral"
	testEnvName   = "TestEnv"
	testAppID     = "d0e4c7e2-1234-5678-abcd-ef0123456789"
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
	c.SetAPIVersion(constants.DefaultAPIVersion)
	c.SetHTTPClient(&http.Client{})
	return c
}

// expectedPath builds the full request path the client produces for an apps sub-path.
func expectedPath(suffix string) string {
	return fmt.Sprintf("/admin/%s/applications/%s/environments/%s/apps%s",
		constants.DefaultAPIVersion, testAppFamily, testEnvName, suffix)
}

func TestService_UploadAndInstall(t *testing.T) {
	t.Run("immediate install", func(t *testing.T) {
		var (
			gotPath        string
			gotMethod      string
			gotContentType string
			gotFileName    string
			gotFileBytes   []byte
			gotFields      map[string]string
		)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotMethod = r.Method
			gotContentType = r.Header.Get("Content-Type")

			if err := r.ParseMultipartForm(10 << 20); err != nil {
				t.Errorf("failed to parse multipart form: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			files := r.MultipartForm.File["extensionFile"]
			if len(files) == 1 {
				gotFileName = files[0].Filename
				f, err := files[0].Open()
				if err == nil {
					defer f.Close()
					gotFileBytes, _ = io.ReadAll(f)
				}
			}

			gotFields = map[string]string{}
			for k, v := range r.MultipartForm.Value {
				if len(v) > 0 {
					gotFields[k] = v[0]
				}
			}

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":               "op-1",
				"type":             "install",
				"status":           "running",
				"appId":            testAppID,
				"targetAppVersion": "1.0.0.0",
				"scheduleKind":     DeploymentScheduleImmediate,
			})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		op, err := svc.UploadAndInstall(context.Background(), testAppFamily, testEnvName, &PteInstallRequest{
			FileName:                          "MyExtension_1.0.0.0.app",
			Content:                           []byte("fake-app-bytes"),
			DeploymentSchedule:                DeploymentScheduleImmediate,
			SyncMode:                          SyncModeAdd,
			LanguageID:                        "en-US",
			AcceptIsvEula:                     true,
			InstallOrUpdateNeededDependencies: true,
		})
		if err != nil {
			t.Fatalf("UploadAndInstall() unexpected error: %v", err)
		}

		if gotMethod != http.MethodPost {
			t.Errorf("method = %q, want POST", gotMethod)
		}
		if want := expectedPath("/pteInstall"); gotPath != want {
			t.Errorf("path = %q, want %q", gotPath, want)
		}
		if !strings.HasPrefix(gotContentType, "multipart/form-data") {
			t.Errorf("Content-Type = %q, want multipart/form-data", gotContentType)
		}
		if gotFileName != "MyExtension_1.0.0.0.app" {
			t.Errorf("uploaded file name = %q, want MyExtension_1.0.0.0.app", gotFileName)
		}
		if string(gotFileBytes) != "fake-app-bytes" {
			t.Errorf("uploaded file bytes = %q, want fake-app-bytes", string(gotFileBytes))
		}

		wantFields := map[string]string{
			"deploymentSchedule":                DeploymentScheduleImmediate,
			"syncMode":                          SyncModeAdd,
			"languageId":                        "en-US",
			"acceptIsvEula":                     "true",
			"installOrUpdateNeededDependencies": "true",
		}
		for k, want := range wantFields {
			if gotFields[k] != want {
				t.Errorf("form field %s = %q, want %q", k, gotFields[k], want)
			}
		}

		if op.AppID != testAppID {
			t.Errorf("AppID = %q, want %q", op.AppID, testAppID)
		}
		if op.TargetVersionValue() != "1.0.0.0" {
			t.Errorf("TargetVersionValue() = %q, want 1.0.0.0", op.TargetVersionValue())
		}
		if op.IsScheduled() {
			t.Error("IsScheduled() = true, want false for a running operation")
		}
	})

	t.Run("legacy schedule and sync mode names are normalized", func(t *testing.T) {
		var gotFields map[string]string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if err := r.ParseMultipartForm(10 << 20); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			gotFields = map[string]string{}
			for k, v := range r.MultipartForm.Value {
				if len(v) > 0 {
					gotFields[k] = v[0]
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "op-1", "status": "running", "appId": testAppID})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		if _, err := svc.UploadAndInstall(context.Background(), testAppFamily, testEnvName, &PteInstallRequest{
			FileName:           "ext.app",
			Content:            []byte("bytes"),
			DeploymentSchedule: "Next minor version",
			SyncMode:           "Force Sync",
			AcceptIsvEula:      true,
		}); err != nil {
			t.Fatalf("UploadAndInstall() unexpected error: %v", err)
		}

		if gotFields["deploymentSchedule"] != DeploymentScheduleNextMinorUpdate {
			t.Errorf("deploymentSchedule = %q, want %q", gotFields["deploymentSchedule"], DeploymentScheduleNextMinorUpdate)
		}
		if gotFields["syncMode"] != SyncModeForceSync {
			t.Errorf("syncMode = %q, want %q", gotFields["syncMode"], SyncModeForceSync)
		}
	})

	t.Run("scheduled install", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id":               "op-2",
				"type":             "install",
				"status":           "scheduled",
				"appId":            testAppID,
				"targetAppVersion": "2.0.0.0",
				"scheduleKind":     DeploymentScheduleNextMajorUpdate,
			})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		op, err := svc.UploadAndInstall(context.Background(), testAppFamily, testEnvName, &PteInstallRequest{
			FileName:           "ext.app",
			Content:            []byte("bytes"),
			DeploymentSchedule: DeploymentScheduleNextMajorUpdate,
			AcceptIsvEula:      true,
		})
		if err != nil {
			t.Fatalf("UploadAndInstall() unexpected error: %v", err)
		}
		if !op.IsScheduled() {
			t.Error("IsScheduled() = false, want true")
		}
		if op.ScheduleKind != DeploymentScheduleNextMajorUpdate {
			t.Errorf("ScheduleKind = %q, want %q", op.ScheduleKind, DeploymentScheduleNextMajorUpdate)
		}
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"code":    "EntityValidationFailed",
				"message": "acceptIsvEula must be true",
			})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		if _, err := svc.UploadAndInstall(context.Background(), testAppFamily, testEnvName, &PteInstallRequest{
			FileName: "ext.app",
			Content:  []byte("bytes"),
		}); err == nil {
			t.Error("UploadAndInstall() expected error for 400 response, got nil")
		}
	})

	t.Run("response without operation id", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"status": "running"})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		if _, err := svc.UploadAndInstall(context.Background(), testAppFamily, testEnvName, &PteInstallRequest{
			FileName: "ext.app",
			Content:  []byte("bytes"),
		}); err == nil {
			t.Error("UploadAndInstall() expected error when operation id is missing, got nil")
		}
	})
}

func TestBuildPteInstallForm_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     *PteInstallRequest
		wantErr string
	}{
		{
			name:    "nil request",
			req:     nil,
			wantErr: "cannot be nil",
		},
		{
			name:    "empty content",
			req:     &PteInstallRequest{FileName: "ext.app"},
			wantErr: "empty",
		},
		{
			name:    "wrong file extension",
			req:     &PteInstallRequest{FileName: "ext.zip", Content: []byte("bytes")},
			wantErr: ".app extension",
		},
		{
			name:    "oversized package",
			req:     &PteInstallRequest{FileName: "ext.app", Content: make([]byte, MaxExtensionFileSize+1)},
			wantErr: "exceeds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := buildPteInstallForm(tt.req)
			if err == nil {
				t.Fatalf("buildPteInstallForm() expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("buildPteInstallForm() error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}

	t.Run("defaults are applied", func(t *testing.T) {
		body, contentType, err := buildPteInstallForm(&PteInstallRequest{Content: []byte("bytes")})
		if err != nil {
			t.Fatalf("buildPteInstallForm() unexpected error: %v", err)
		}
		if !strings.HasPrefix(contentType, "multipart/form-data") {
			t.Errorf("contentType = %q, want multipart/form-data", contentType)
		}
		encoded := body.String()
		for _, want := range []string{"extension.app", DeploymentScheduleImmediate, SyncModeAdd} {
			if !strings.Contains(encoded, want) {
				t.Errorf("form body missing %q", want)
			}
		}
	})
}

func TestService_GetApp(t *testing.T) {
	tests := []struct {
		name           string
		responseStatus int
		responseBody   interface{}
		wantApp        bool
		wantErr        bool
	}{
		{
			name:           "app found",
			responseStatus: http.StatusOK,
			responseBody: AppListResponse{Value: []App{
				{AppID: "other-app", Name: "Other"},
				{AppID: testAppID, Name: "My Extension", Publisher: "Contoso", Version: "1.0.0.0", State: AppStateInstalled, AppType: AppTypeTenant},
			}},
			wantApp: true,
		},
		{
			name:           "app not installed",
			responseStatus: http.StatusOK,
			responseBody:   AppListResponse{Value: []App{{AppID: "other-app"}}},
			wantApp:        false,
		},
		{
			name:           "empty list",
			responseStatus: http.StatusOK,
			responseBody:   AppListResponse{Value: []App{}},
			wantApp:        false,
		},
		{
			name:           "api error",
			responseStatus: http.StatusForbidden,
			responseBody:   map[string]string{"code": "Forbidden", "message": "no access"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if want := expectedPath(""); r.URL.Path != want {
					t.Errorf("path = %q, want %q", r.URL.Path, want)
				}
				w.WriteHeader(tt.responseStatus)
				_ = json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			svc := NewService(newTestClient(t, server.URL))

			app, err := svc.GetApp(context.Background(), testAppFamily, testEnvName, testAppID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetApp() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if (app != nil) != tt.wantApp {
				t.Fatalf("GetApp() app = %v, wantApp %v", app, tt.wantApp)
			}
			if tt.wantApp && app.Version != "1.0.0.0" {
				t.Errorf("GetApp() version = %q, want 1.0.0.0", app.Version)
			}
		})
	}
}

func TestService_GetOperation(t *testing.T) {
	t.Run("bare operation object", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if want := expectedPath("/" + testAppID + "/operations/op-1"); r.URL.Path != want {
				t.Errorf("path = %q, want %q", r.URL.Path, want)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "op-1", "status": "Succeeded", "type": "install", "targetVersion": "1.0.0.0",
			})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		op, err := svc.GetOperation(context.Background(), testAppFamily, testEnvName, testAppID, "op-1")
		if err != nil {
			t.Fatalf("GetOperation() unexpected error: %v", err)
		}
		if op.ID != "op-1" {
			t.Errorf("ID = %q, want op-1", op.ID)
		}
		if op.TargetVersionValue() != "1.0.0.0" {
			t.Errorf("TargetVersionValue() = %q, want 1.0.0.0", op.TargetVersionValue())
		}
	})

	t.Run("list wrapper", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(AppOperationListResponse{Value: []AppOperation{
				{ID: "op-2", Status: "succeeded", TargetAppVersion: "2.0.0.0"},
			}})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		op, err := svc.GetOperation(context.Background(), testAppFamily, testEnvName, testAppID, "op-2")
		if err != nil {
			t.Fatalf("GetOperation() unexpected error: %v", err)
		}
		if op.ID != "op-2" {
			t.Errorf("ID = %q, want op-2", op.ID)
		}
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "ResourceDoesNotExist", "message": "gone"})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		if _, err := svc.GetOperation(context.Background(), testAppFamily, testEnvName, testAppID, "op-3"); err == nil {
			t.Error("GetOperation() expected error, got nil")
		}
	})
}

func TestService_WaitForOperation(t *testing.T) {
	tests := []struct {
		name                string
		status              string
		scheduledIsTerminal bool
		wantErr             bool
	}{
		{name: "succeeded", status: "succeeded"},
		{name: "succeeded capitalised", status: "Succeeded"},
		{name: "deferred scheduled returns without waiting", status: "scheduled", scheduledIsTerminal: true},
		{name: "failed", status: "failed", wantErr: true},
		{name: "canceled", status: "canceled", wantErr: true},
		{name: "skipped", status: "skipped", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"id": "op-1", "status": tt.status, "errorMessage": "boom",
				})
			}))
			defer server.Close()

			svc := NewService(newTestClient(t, server.URL))

			op, err := svc.WaitForOperation(context.Background(), testAppFamily, testEnvName, testAppID, "op-1", 30*time.Second, tt.scheduledIsTerminal)
			if (err != nil) != tt.wantErr {
				t.Fatalf("WaitForOperation() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && op == nil {
				t.Error("WaitForOperation() returned nil operation without an error")
			}
		})
	}

	t.Run("times out while still running", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "op-1", "status": "running"})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		_, err := svc.WaitForOperation(context.Background(), testAppFamily, testEnvName, testAppID, "op-1", 100*time.Millisecond, false)
		if err == nil {
			t.Fatal("WaitForOperation() expected a timeout error, got nil")
		}
		if !strings.Contains(err.Error(), "timed out") {
			t.Errorf("WaitForOperation() error = %q, want a timeout error", err.Error())
		}
	})
}

func TestService_Uninstall(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotBody UninstallAppRequest
		var gotPath string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "op-9", "type": "uninstall", "status": "running"})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		op, err := svc.Uninstall(context.Background(), testAppFamily, testEnvName, testAppID, &UninstallAppRequest{
			DeleteData:          true,
			UninstallDependents: true,
		})
		if err != nil {
			t.Fatalf("Uninstall() unexpected error: %v", err)
		}
		if want := expectedPath("/" + testAppID + "/uninstall"); gotPath != want {
			t.Errorf("path = %q, want %q", gotPath, want)
		}
		if !gotBody.DeleteData || !gotBody.UninstallDependents {
			t.Errorf("request body = %+v, want deleteData and uninstallDependents true", gotBody)
		}
		if op.ID != "op-9" {
			t.Errorf("ID = %q, want op-9", op.ID)
		}
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "EntityValidationFailed", "message": "dependents exist"})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		if _, err := svc.Uninstall(context.Background(), testAppFamily, testEnvName, testAppID, &UninstallAppRequest{}); err == nil {
			t.Error("Uninstall() expected error, got nil")
		}
	})
}

func TestService_ListScheduledPteOperations(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotPath string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"value": []map[string]interface{}{
					{
						"id":               "op-1",
						"type":             "Install",
						"status":           "scheduled",
						"appId":            testAppID,
						"targetAppVersion": "2.0.0.0",
						"scheduleKind":     DeploymentScheduleNextMinorUpdate,
						"parameters": map[string]interface{}{
							"name":      "My Extension",
							"publisher": "Contoso",
							"syncMode":  SyncModeAdd,
						},
					},
				},
			})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		ops, err := svc.ListScheduledPteOperations(context.Background(), testAppFamily, testEnvName)
		if err != nil {
			t.Fatalf("ListScheduledPteOperations() unexpected error: %v", err)
		}
		if want := expectedPath("/scheduledPteOperations"); gotPath != want {
			t.Errorf("path = %q, want %q", gotPath, want)
		}
		if len(ops) != 1 {
			t.Fatalf("got %d operations, want 1", len(ops))
		}
		if ops[0].Parameters == nil || ops[0].Parameters.Name != "My Extension" {
			t.Errorf("parameters = %+v, want name %q", ops[0].Parameters, "My Extension")
		}
		if ops[0].TargetVersionValue() != "2.0.0.0" {
			t.Errorf("TargetVersionValue() = %q, want 2.0.0.0", ops[0].TargetVersionValue())
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(ScheduledPteOperationListResponse{Value: []ScheduledPteOperation{}})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		ops, err := svc.ListScheduledPteOperations(context.Background(), testAppFamily, testEnvName)
		if err != nil {
			t.Fatalf("ListScheduledPteOperations() unexpected error: %v", err)
		}
		if len(ops) != 0 {
			t.Errorf("got %d operations, want 0", len(ops))
		}
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "NotFound", "message": "unsupported api version"})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		if _, err := svc.ListScheduledPteOperations(context.Background(), testAppFamily, testEnvName); err == nil {
			t.Error("ListScheduledPteOperations() expected error, got nil")
		}
	})
}

func TestService_GetScheduledPteOperationsForApp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(ScheduledPteOperationListResponse{Value: []ScheduledPteOperation{
			{AppOperation: AppOperation{ID: "op-1", AppID: testAppID, TargetAppVersion: "2.0.0.0"}},
			{AppOperation: AppOperation{ID: "op-2", AppID: "another-app", TargetAppVersion: "3.0.0.0"}},
			{AppOperation: AppOperation{ID: "op-3", AppID: strings.ToUpper(testAppID), TargetAppVersion: "4.0.0.0"}},
		}})
	}))
	defer server.Close()

	svc := NewService(newTestClient(t, server.URL))

	ops, err := svc.GetScheduledPteOperationsForApp(context.Background(), testAppFamily, testEnvName, testAppID)
	if err != nil {
		t.Fatalf("GetScheduledPteOperationsForApp() unexpected error: %v", err)
	}
	if len(ops) != 2 {
		t.Fatalf("got %d operations, want 2 (app id match is case-insensitive)", len(ops))
	}
}

func TestService_RemoveScheduledPteVersion(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var gotPath string
		var gotBody RemoveScheduledPteVersionRequest

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			_ = json.NewDecoder(r.Body).Decode(&gotBody)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "op-1", "type": "install", "status": "canceled", "targetAppVersion": "2.0.0.0",
			})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		op, err := svc.RemoveScheduledPteVersion(context.Background(), testAppFamily, testEnvName, testAppID, &RemoveScheduledPteVersionRequest{
			TargetVersion: "2.0.0.0",
			ScheduleKind:  DeploymentScheduleNextMinorUpdate,
		})
		if err != nil {
			t.Fatalf("RemoveScheduledPteVersion() unexpected error: %v", err)
		}
		if want := expectedPath("/" + testAppID + "/removeScheduledPteVersion"); gotPath != want {
			t.Errorf("path = %q, want %q", gotPath, want)
		}
		if gotBody.TargetVersion != "2.0.0.0" || gotBody.ScheduleKind != DeploymentScheduleNextMinorUpdate {
			t.Errorf("request body = %+v, want target version 2.0.0.0 and schedule kind %q", gotBody, DeploymentScheduleNextMinorUpdate)
		}
		if op.Status != "canceled" {
			t.Errorf("status = %q, want canceled", op.Status)
		}
	})

	t.Run("missing required fields", func(t *testing.T) {
		svc := NewService(newTestClient(t, "http://127.0.0.1:1"))

		for _, req := range []*RemoveScheduledPteVersionRequest{
			nil,
			{ScheduleKind: DeploymentScheduleNextMinorUpdate},
			{TargetVersion: "2.0.0.0"},
		} {
			if _, err := svc.RemoveScheduledPteVersion(context.Background(), testAppFamily, testEnvName, testAppID, req); err == nil {
				t.Errorf("RemoveScheduledPteVersion(%+v) expected error, got nil", req)
			}
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"code":    "ResourceDoesNotExist",
				"message": "App with id 'x' is not installed on environment or the environment does not exist.",
			})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		_, err := svc.RemoveScheduledPteVersion(context.Background(), testAppFamily, testEnvName, testAppID, &RemoveScheduledPteVersionRequest{
			TargetVersion: "2.0.0.0",
			ScheduleKind:  DeploymentScheduleNextMinorUpdate,
		})
		if err == nil {
			t.Fatal("RemoveScheduledPteVersion() expected error, got nil")
		}
		if !IsNotFoundError(err) {
			t.Errorf("IsNotFoundError() = false for %v, want true", err)
		}
	})
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil error", err: nil, want: false},
		{name: "plain error", err: fmt.Errorf("network unreachable"), want: false},
		{name: "resource does not exist", err: &client.AdminCenterError{Code: "ResourceDoesNotExist"}, want: true},
		{name: "not found code", err: &client.AdminCenterError{Code: "NotFound"}, want: true},
		{
			name: "message based match",
			err:  &client.AdminCenterError{Code: "Unknown", Message: "App with id 'x' is not installed on environment"},
			want: true,
		},
		{name: "other api error", err: &client.AdminCenterError{Code: "EntityValidationFailed", Message: "bad request"}, want: false},
		{
			name: "wrapped api error",
			err:  fmt.Errorf("failed to uninstall: %w", &client.AdminCenterError{Code: "ResourceDoesNotExist"}),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFoundError(tt.err); got != tt.want {
				t.Errorf("IsNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeDeploymentSchedule(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", DeploymentScheduleImmediate},
		{DeploymentScheduleImmediate, DeploymentScheduleImmediate},
		{DeploymentScheduleUpdateWindow, DeploymentScheduleUpdateWindow},
		{DeploymentScheduleNextMinorUpdate, DeploymentScheduleNextMinorUpdate},
		{DeploymentScheduleNextMajorUpdate, DeploymentScheduleNextMajorUpdate},
		{"Current version", DeploymentScheduleImmediate},
		{"Next minor version", DeploymentScheduleNextMinorUpdate},
		{"Next major version", DeploymentScheduleNextMajorUpdate},
		{"  next minor version  ", DeploymentScheduleNextMinorUpdate},
		{"SomethingElse", "SomethingElse"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := NormalizeDeploymentSchedule(tt.input); got != tt.want {
				t.Errorf("NormalizeDeploymentSchedule(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeSyncMode(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", SyncModeAdd},
		{SyncModeAdd, SyncModeAdd},
		{SyncModeForceSync, SyncModeForceSync},
		{"Force Sync", SyncModeForceSync},
		{"force sync", SyncModeForceSync},
		{"SomethingElse", "SomethingElse"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := NormalizeSyncMode(tt.input); got != tt.want {
				t.Errorf("NormalizeSyncMode(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestAppOperation_VersionAccessors(t *testing.T) {
	tests := []struct {
		name           string
		op             AppOperation
		wantSource     string
		wantTarget     string
		wantIsSchedule bool
	}{
		{
			name:       "pteInstall field names",
			op:         AppOperation{SourceAppVersion: "1.0.0.0", TargetAppVersion: "2.0.0.0", Status: "running"},
			wantSource: "1.0.0.0",
			wantTarget: "2.0.0.0",
		},
		{
			name:       "app operations field names",
			op:         AppOperation{SourceVersion: "1.0.0.0", TargetVersion: "2.0.0.0", Status: "running"},
			wantSource: "1.0.0.0",
			wantTarget: "2.0.0.0",
		},
		{
			name:           "scheduled status is case-insensitive",
			op:             AppOperation{Status: "Scheduled"},
			wantIsSchedule: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.SourceVersionValue(); got != tt.wantSource {
				t.Errorf("SourceVersionValue() = %q, want %q", got, tt.wantSource)
			}
			if got := tt.op.TargetVersionValue(); got != tt.wantTarget {
				t.Errorf("TargetVersionValue() = %q, want %q", got, tt.wantTarget)
			}
			if got := tt.op.IsScheduled(); got != tt.wantIsSchedule {
				t.Errorf("IsScheduled() = %v, want %v", got, tt.wantIsSchedule)
			}
		})
	}
}

// TestService_GetApp_LiveWireShape pins the shapes the live API actually returns, which
// differ from the published documentation: the identity field is "id" (not "appId") and
// state is camelCase (not TitleCase).
func TestService_GetApp_LiveWireShape(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"value":[
			{"id":"other-app","name":"Other","appType":"global"},
			{"id":"` + testAppID + `","name":"Infoma Cloud Migration Tools","publisher":"Axians Infoma",
			 "version":"2.0.3.0","state":"installed","appType":"tenant","canBeUninstalled":true}
		]}`))
	}))
	defer server.Close()

	svc := NewService(newTestClient(t, server.URL))

	app, err := svc.GetApp(context.Background(), testAppFamily, testEnvName, testAppID)
	if err != nil {
		t.Fatalf("GetApp() unexpected error: %v", err)
	}
	if app == nil {
		t.Fatal("GetApp() returned nil for an app the API reports under the \"id\" field")
	}
	if app.Identity() != testAppID {
		t.Errorf("Identity() = %q, want %q", app.Identity(), testAppID)
	}
	if app.Version != "2.0.3.0" {
		t.Errorf("Version = %q, want 2.0.3.0", app.Version)
	}
	if app.State != AppStateInstalled {
		t.Errorf("State = %q, want %q", app.State, AppStateInstalled)
	}
	if app.AppType != AppTypeTenant {
		t.Errorf("AppType = %q, want %q", app.AppType, AppTypeTenant)
	}
}

// TestApp_Identity covers both field spellings, since the documentation and the live API
// disagree about which one carries the app identity.
func TestApp_Identity(t *testing.T) {
	tests := []struct {
		name string
		app  App
		want string
	}{
		{name: "live shape uses id", app: App{ID: testAppID}, want: testAppID},
		{name: "documented shape uses appId", app: App{AppID: testAppID}, want: testAppID},
		{name: "id wins when both are present", app: App{ID: testAppID, AppID: "other"}, want: testAppID},
		{name: "neither present", app: App{}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.app.Identity(); got != tt.want {
				t.Errorf("Identity() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestService_GetOperation_StaleHistoryGuard pins the behaviour of the operations
// endpoint observed live: when it cannot narrow to a single operation it returns the
// app's entire, unordered operation history. Picking the first entry would report the
// status of an unrelated long-completed operation, so a list must be searched by ID.
func TestService_GetOperation_StaleHistoryGuard(t *testing.T) {
	// Value[0] is a stale success; the operation actually being polled is still running.
	history := AppOperationListResponse{Value: []AppOperation{
		{ID: "old-install", Type: "install", Status: "succeeded", CreatedOn: "2026-08-26T12:59:02.997Z"},
		{ID: "op-current", Type: "uninstall", Status: "running", CreatedOn: "2026-08-28T09:58:14.850Z"},
	}}

	t.Run("list response is matched by id, not position", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(history)
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		op, err := svc.GetOperation(context.Background(), testAppFamily, testEnvName, testAppID, "op-current")
		if err != nil {
			t.Fatalf("GetOperation() unexpected error: %v", err)
		}
		if op.ID != "op-current" {
			t.Fatalf("GetOperation() returned %q, want op-current (the first entry is a stale success)", op.ID)
		}
		if op.Status != "running" {
			t.Errorf("Status = %q, want running", op.Status)
		}
	})

	t.Run("unknown id in a list response is an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(history)
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		if _, err := svc.GetOperation(context.Background(), testAppFamily, testEnvName, testAppID, "op-missing"); err == nil {
			t.Error("GetOperation() expected an error when the id is absent from the returned list, got nil")
		}
	})

	t.Run("empty operation id is rejected without a request", func(t *testing.T) {
		var called bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			_ = json.NewEncoder(w).Encode(history)
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		if _, err := svc.GetOperation(context.Background(), testAppFamily, testEnvName, testAppID, ""); err == nil {
			t.Error("GetOperation() expected an error for an empty operation id, got nil")
		}
		if called {
			t.Error("GetOperation() sent a request for an empty operation id; the API would answer with the full history")
		}
	})

	t.Run("WaitForOperation does not succeed on a stale entry", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_ = json.NewEncoder(w).Encode(history)
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		_, err := svc.WaitForOperation(context.Background(), testAppFamily, testEnvName, testAppID, "op-current", 100*time.Millisecond, false)
		if err == nil {
			t.Fatal("WaitForOperation() returned success while the polled operation was still running")
		}
		if !strings.Contains(err.Error(), "timed out") {
			t.Errorf("WaitForOperation() error = %q, want a timeout", err.Error())
		}
	})
}

func TestService_WaitForAppRemoval(t *testing.T) {
	t.Run("returns once the app leaves the list", func(t *testing.T) {
		var calls int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls++
			if calls == 1 {
				// Still winding down, exactly as the live API reports it.
				_, _ = w.Write([]byte(`{"value":[{"id":"` + testAppID + `","state":"uninstallPending","appType":"tenant"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"value":[]}`))
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		if err := svc.WaitForAppRemoval(context.Background(), testAppFamily, testEnvName, testAppID, 60*time.Second); err != nil {
			t.Fatalf("WaitForAppRemoval() unexpected error: %v", err)
		}
		if calls < 2 {
			t.Errorf("WaitForAppRemoval() polled %d times, want it to keep polling past uninstallPending", calls)
		}
	})

	t.Run("returns immediately when already absent", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"value":[]}`))
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		if err := svc.WaitForAppRemoval(context.Background(), testAppFamily, testEnvName, testAppID, 60*time.Second); err != nil {
			t.Fatalf("WaitForAppRemoval() unexpected error: %v", err)
		}
	})

	t.Run("times out while still pending", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(`{"value":[{"id":"` + testAppID + `","state":"uninstallPending","appType":"tenant"}]}`))
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		err := svc.WaitForAppRemoval(context.Background(), testAppFamily, testEnvName, testAppID, 100*time.Millisecond)
		if err == nil {
			t.Fatal("WaitForAppRemoval() expected a timeout error, got nil")
		}
		if !strings.Contains(err.Error(), "uninstallPending") {
			t.Errorf("WaitForAppRemoval() error = %q, want it to report the last observed state", err.Error())
		}
	})
}

// TestService_WaitForOperation_ScheduledIsTransient pins the behaviour observed live: an
// immediate uninstall is reported as "scheduled" while it is queued, before it starts
// running. Only a caller that asked for a deferred deployment may treat that as terminal.
func TestService_WaitForOperation_ScheduledIsTransient(t *testing.T) {
	t.Run("queued immediate operation is waited through", func(t *testing.T) {
		var calls int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls++
			status := "scheduled"
			if calls > 1 {
				status = "succeeded"
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"id": "op-1", "type": "uninstall", "status": status})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		op, err := svc.WaitForOperation(context.Background(), testAppFamily, testEnvName, testAppID, "op-1", 60*time.Second, false)
		if err != nil {
			t.Fatalf("WaitForOperation() unexpected error: %v", err)
		}
		if !strings.EqualFold(op.Status, OperationStatusSucceeded) {
			t.Errorf("Status = %q, want succeeded; a queued \"scheduled\" must not end the wait", op.Status)
		}
		if calls < 2 {
			t.Errorf("polled %d times, want it to keep polling past the queued state", calls)
		}
	})

	t.Run("deferred operation returns on the first scheduled poll", func(t *testing.T) {
		var calls int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			calls++
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "op-1", "type": "install", "status": "scheduled", "scheduleKind": "NextMinorUpdate",
			})
		}))
		defer server.Close()

		svc := NewService(newTestClient(t, server.URL))

		op, err := svc.WaitForOperation(context.Background(), testAppFamily, testEnvName, testAppID, "op-1", 60*time.Second, true)
		if err != nil {
			t.Fatalf("WaitForOperation() unexpected error: %v", err)
		}
		if !op.IsScheduled() {
			t.Errorf("Status = %q, want scheduled", op.Status)
		}
		if calls != 1 {
			t.Errorf("polled %d times, want exactly 1 for a genuinely deferred operation", calls)
		}
	})
}
