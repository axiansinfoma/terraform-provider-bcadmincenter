// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
)

// mockTokenCredential is a mock implementation of azcore.TokenCredential for testing.
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
		name              string
		applicationFamily string
		responseBody      interface{}
		responseStatus    int
		wantErr           bool
		expectedEnvCount  int
	}{
		{
			name:              "successful response with environments",
			applicationFamily: "BusinessCentral",
			responseBody: EnvironmentListResponse{
				Value: []Environment{
					{
						Name:              "production",
						Type:              "Production",
						ApplicationFamily: "BusinessCentral",
						Status:            "Active",
					},
					{
						Name:              "sandbox",
						Type:              "Sandbox",
						ApplicationFamily: "BusinessCentral",
						Status:            "Active",
					},
				},
			},
			responseStatus:   http.StatusOK,
			wantErr:          false,
			expectedEnvCount: 2,
		},
		{
			name:              "empty response",
			applicationFamily: "BusinessCentral",
			responseBody: EnvironmentListResponse{
				Value: []Environment{},
			},
			responseStatus:   http.StatusOK,
			wantErr:          false,
			expectedEnvCount: 0,
		},
		{
			name:              "server error",
			applicationFamily: "BusinessCentral",
			responseBody: map[string]string{
				"error": "internal server error",
			},
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

			mockCred := &mockTokenCredential{token: "test-token"}
			c := &client.Client{}
			c.SetCredential(mockCred)
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			envs, err := svc.List(context.Background(), tt.applicationFamily)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(envs) != tt.expectedEnvCount {
				t.Errorf("List() returned %d environments, expected %d", len(envs), tt.expectedEnvCount)
			}
		})
	}
}

func TestService_Get(t *testing.T) {
	tests := []struct {
		name              string
		applicationFamily string
		environmentName   string
		responseBody      interface{}
		responseStatus    int
		wantErr           bool
	}{
		{
			name:              "successful retrieval",
			applicationFamily: "BusinessCentral",
			environmentName:   "production",
			responseBody: Environment{
				Name:              "production",
				Type:              "Production",
				ApplicationFamily: "BusinessCentral",
				Status:            "Active",
				CountryCode:       "US",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:              "environment not found",
			applicationFamily: "BusinessCentral",
			environmentName:   "nonexistent",
			responseBody: map[string]string{
				"error": "not found",
			},
			responseStatus: http.StatusNotFound,
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
			env, err := svc.Get(context.Background(), tt.applicationFamily, tt.environmentName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && env.Name != tt.environmentName {
				t.Errorf("Get() returned environment name %s, expected %s", env.Name, tt.environmentName)
			}
		})
	}
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name              string
		applicationFamily string
		request           *CreateEnvironmentRequest
		responseBody      interface{}
		responseStatus    int
		wantErr           bool
	}{
		{
			name:              "successful creation",
			applicationFamily: "BusinessCentral",
			request: &CreateEnvironmentRequest{
				EnvironmentType: "Sandbox",
				Name:            "test-env",
				CountryCode:     "US",
				RingName:        "PROD",
			},
			responseBody: Operation{
				ID:                "op-123",
				Type:              "CreateEnvironment",
				Status:            "Queued",
				ApplicationFamily: "BusinessCentral",
			},
			responseStatus: http.StatusAccepted,
			wantErr:        false,
		},
		{
			name:              "bad request",
			applicationFamily: "BusinessCentral",
			request: &CreateEnvironmentRequest{
				EnvironmentType: "Invalid",
				Name:            "test",
			},
			responseBody: map[string]string{
				"error": "invalid environment type",
			},
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
			operation, err := svc.Create(context.Background(), tt.applicationFamily, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && operation == nil {
				t.Error("Create() returned nil operation")
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name              string
		applicationFamily string
		environmentName   string
		responseBody      interface{}
		responseStatus    int
		wantErr           bool
		expectNilOp       bool
	}{
		{
			name:              "successful deletion",
			applicationFamily: "BusinessCentral",
			environmentName:   "test-env",
			responseBody: Operation{
				ID:     "op-456",
				Type:   "DeleteEnvironment",
				Status: "Queued",
			},
			responseStatus: http.StatusAccepted,
			wantErr:        false,
			expectNilOp:    false,
		},
		{
			name:              "already deleted",
			applicationFamily: "BusinessCentral",
			environmentName:   "test-env",
			responseBody:      nil,
			responseStatus:    http.StatusNoContent,
			wantErr:           false,
			expectNilOp:       true,
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
			operation, err := svc.Delete(context.Background(), tt.applicationFamily, tt.environmentName)

			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.expectNilOp && operation != nil {
				t.Error("Delete() expected nil operation but got one")
			}

			if !tt.expectNilOp && !tt.wantErr && operation == nil {
				t.Error("Delete() returned nil operation")
			}
		})
	}
}

func TestService_GetOperation(t *testing.T) {
	tests := []struct {
		name              string
		applicationFamily string
		environmentName   string
		operationID       string
		responseBody      interface{}
		responseStatus    int
		wantErr           bool
	}{
		{
			name:              "successful retrieval",
			applicationFamily: "BusinessCentral",
			environmentName:   "test-env",
			operationID:       "op-123",
			responseBody: Operation{
				ID:     "op-123",
				Type:   "CreateEnvironment",
				Status: "Succeeded",
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:              "operation not found",
			applicationFamily: "BusinessCentral",
			environmentName:   "test-env",
			operationID:       "op-999",
			responseBody: map[string]string{
				"error": "not found",
			},
			responseStatus: http.StatusNotFound,
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
			operation, err := svc.GetOperation(context.Background(), tt.applicationFamily, tt.environmentName, tt.operationID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && operation.ID != tt.operationID {
				t.Errorf("GetOperation() returned operation ID %s, expected %s", operation.ID, tt.operationID)
			}
		})
	}
}

func TestService_WaitForOperation(t *testing.T) {
	tests := []struct {
		name              string
		applicationFamily string
		environmentName   string
		operationID       string
		operationStatus   string
		errorMessage      string
		wantErr           bool
	}{
		{
			name:              "operation succeeds immediately",
			applicationFamily: "BusinessCentral",
			environmentName:   "test-env",
			operationID:       "op-123",
			operationStatus:   OperationStatusSucceeded,
			wantErr:           false,
		},
		{
			name:              "operation fails",
			applicationFamily: "BusinessCentral",
			environmentName:   "test-env",
			operationID:       "op-456",
			operationStatus:   OperationStatusFailed,
			errorMessage:      "Something went wrong",
			wantErr:           true,
		},
		{
			name:              "operation cancelled",
			applicationFamily: "BusinessCentral",
			environmentName:   "test-env",
			operationID:       "op-789",
			operationStatus:   OperationStatusCancelled,
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				if err := json.NewEncoder(w).Encode(Operation{
					ID:           tt.operationID,
					Status:       tt.operationStatus,
					ErrorMessage: tt.errorMessage,
				}); err != nil {

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
			err := svc.WaitForOperation(context.Background(), tt.applicationFamily, tt.environmentName, tt.operationID, 5*time.Second)

			if (err != nil) != tt.wantErr {
				t.Errorf("WaitForOperation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewService(t *testing.T) {
	mockCred := &mockTokenCredential{token: "test-token"}
	c := &client.Client{}
	c.SetCredential(mockCred)

	svc := NewService(c)

	if svc == nil {
		t.Error("NewService() returned nil")
		return
	}

	if svc.client == nil {
		t.Error("NewService() returned service with nil client")
	}
}

func TestIsEnvironmentNotFoundError(t *testing.T) {
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
			name: "admin center error code",
			err: &client.AdminCenterError{
				Code:    "EnvironmentNotFound",
				Message: "environment missing",
			},
			want: true,
		},
		{
			name: "wrapped admin center error code",
			err: fmt.Errorf("wrapped: %w", &client.AdminCenterError{
				Code:    "EnvironmentNotFound",
				Message: "environment missing",
			}),
			want: true,
		},
		{
			name: "fallback message match",
			err:  fmt.Errorf("request failed: EnvironmentNotFound"),
			want: true,
		},
		{
			name: "different error",
			err: &client.AdminCenterError{
				Code:    "ResourceNotFound",
				Message: "resource missing",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isEnvironmentNotFoundError(tt.err); got != tt.want {
				t.Errorf("isEnvironmentNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_GetUpdates(t *testing.T) {
	tests := []struct {
		name              string
		applicationFamily string
		environmentName   string
		responseBody      interface{}
		responseStatus    int
		wantErr           bool
		expectedCount     int
	}{
		{
			name:              "successful response with updates",
			applicationFamily: "BusinessCentral",
			environmentName:   "production",
			responseBody: EnvironmentUpdatesResponse{
				Value: []EnvironmentUpdate{
					{
						TargetVersion: "26.0",
						Available:     true,
						Selected:      false,
					},
					{
						TargetVersion: "26.1",
						Available:     true,
						Selected:      true,
						UpdateStatus:  UpdateStatusScheduled,
					},
				},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedCount:  2,
		},
		{
			name:              "empty updates list",
			applicationFamily: "BusinessCentral",
			environmentName:   "production",
			responseBody: EnvironmentUpdatesResponse{
				Value: []EnvironmentUpdate{},
			},
			responseStatus: http.StatusOK,
			wantErr:        false,
			expectedCount:  0,
		},
		{
			name:              "server error",
			applicationFamily: "BusinessCentral",
			environmentName:   "production",
			responseBody:      map[string]string{"error": "internal server error"},
			responseStatus:    http.StatusInternalServerError,
			wantErr:           true,
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
			updates, err := svc.GetUpdates(context.Background(), tt.applicationFamily, tt.environmentName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUpdates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(updates) != tt.expectedCount {
				t.Errorf("GetUpdates() returned %d updates, expected %d", len(updates), tt.expectedCount)
			}
		})
	}
}

func TestService_SelectUpdateVersion(t *testing.T) {
	tests := []struct {
		name               string
		targetVersion      string
		ignoreUpdateWindow bool
		responseStatus     int
		wantErr            bool
	}{
		{
			name:               "successful select",
			targetVersion:      "26.1",
			ignoreUpdateWindow: false,
			responseStatus:     http.StatusOK,
			wantErr:            false,
		},
		{
			name:               "successful select with ignore window",
			targetVersion:      "26.1",
			ignoreUpdateWindow: true,
			responseStatus:     http.StatusNoContent,
			wantErr:            false,
		},
		{
			name:           "bad request",
			targetVersion:  "invalid",
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
			err := svc.SelectUpdateVersion(context.Background(), "BusinessCentral", "production", tt.targetVersion, tt.ignoreUpdateWindow)

			if (err != nil) != tt.wantErr {
				t.Errorf("SelectUpdateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestService_SelectUpdateVersion_RetriesOnPastDateTimeError verifies that SelectUpdateVersion
// retries with a valid future selectedDateTime when the API rejects the first attempt due to a
// stale past selectedDateTime (EntityValidationFailed). On retry it caps the datetime to
// latestSelectableDateTime fetched from the updates list.
func TestService_SelectUpdateVersion_RetriesOnPastDateTimeError(t *testing.T) {
	// latestSelectableDateTime ~6 months from now so candidate (now+1h) stays within range.
	latestSelectable := time.Now().UTC().Add(6 * 30 * 24 * time.Hour).Format(time.RFC3339)

	patchBodies := make([]map[string]interface{}, 0, 2)
	getCalled := false
	patchCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getCalled = true
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"value": []map[string]interface{}{
					{
						"targetVersion": "27.2",
						"available":     true,
						"selected":      true,
						"scheduleDetails": map[string]interface{}{
							"latestSelectableDateTime": latestSelectable,
							"selectedDateTime":         "2026-01-12T21:00:00Z",
						},
					},
				},
			})
			return
		}
		// PATCH
		patchCount++
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
			patchBodies = append(patchBodies, body)
		}
		if patchCount == 1 {
			// First PATCH: simulate the "selected date time in the past" error.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"code":    "EntityValidationFailed",
				"message": "Update currently has selected date time in the past (2026-01-12T21:00:00.0000000Z) and cannot be selected. Modify the selected date time first.",
			})
		} else {
			// Second PATCH (retry): succeed.
			w.WriteHeader(http.StatusOK)
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
	err := svc.SelectUpdateVersion(context.Background(), "BusinessCentral", "production", "27.2", false)
	if err != nil {
		t.Fatalf("SelectUpdateVersion() unexpected error: %v", err)
	}

	if patchCount != 2 {
		t.Fatalf("expected 2 PATCH requests (initial + retry), got %d", patchCount)
	}
	if !getCalled {
		t.Error("expected GET updates call to resolve latestSelectableDateTime for retry")
	}

	// First PATCH: plain select — no selectedDateTime.
	firstBody := patchBodies[0]
	if selected, ok := firstBody["selected"].(bool); !ok || !selected {
		t.Errorf("first PATCH expected selected:true, got %v", firstBody["selected"])
	}
	if details, ok := firstBody["scheduleDetails"].(map[string]interface{}); ok {
		if _, hasDateTime := details["selectedDateTime"]; hasDateTime {
			t.Error("first PATCH should not include selectedDateTime")
		}
	}

	// Second PATCH (retry): must include selected:true and a valid future selectedDateTime.
	retryBody := patchBodies[1]
	if selected, ok := retryBody["selected"].(bool); !ok || !selected {
		t.Errorf("retry PATCH expected selected:true, got %v", retryBody["selected"])
	}
	details, hasDetails := retryBody["scheduleDetails"].(map[string]interface{})
	if !hasDetails {
		t.Fatal("retry PATCH must include 'scheduleDetails'")
	}
	selectedDateTime, hasDateTime := details["selectedDateTime"].(string)
	if !hasDateTime || selectedDateTime == "" {
		t.Error("retry PATCH scheduleDetails must include a non-empty 'selectedDateTime'")
	} else {
		dt, parseErr := time.Parse(time.RFC3339, selectedDateTime)
		if parseErr != nil {
			t.Errorf("retry selectedDateTime is not valid RFC3339: %v", selectedDateTime)
		} else {
			if !dt.After(time.Now()) {
				t.Errorf("retry selectedDateTime must be in the future, got: %v", selectedDateTime)
			}
			latest, _ := time.Parse(time.RFC3339, latestSelectable)
			if dt.After(latest) {
				t.Errorf("retry selectedDateTime %v exceeds latestSelectableDateTime %v", dt, latest)
			}
		}
	}
}

// TestService_SelectUpdateVersion_SinglePatchOnSuccess verifies that SelectUpdateVersion sends
// only ONE PATCH (no GET, no retry) when the first attempt succeeds.
func TestService_SelectUpdateVersion_SinglePatchOnSuccess(t *testing.T) {
	patchCount := 0
	getCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getCalled = true
			w.WriteHeader(http.StatusOK)
			return
		}
		patchCount++
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockCred := &mockTokenCredential{token: "test-token"}
	c := &client.Client{}
	c.SetCredential(mockCred)
	c.SetBaseURL(server.URL)
	c.SetAPIVersion(constants.DefaultAPIVersion)
	c.SetHTTPClient(&http.Client{})

	svc := NewService(c)
	err := svc.SelectUpdateVersion(context.Background(), "BusinessCentral", "production", "27.2", false)
	if err != nil {
		t.Errorf("SelectUpdateVersion() unexpected error: %v", err)
	}
	if patchCount != 1 {
		t.Errorf("expected exactly 1 PATCH (no retry needed), got %d", patchCount)
	}
	if getCalled {
		t.Error("GET should not be called when first PATCH succeeds")
	}
}

func TestService_ScheduleUpdateVersion(t *testing.T) {
	tests := []struct {
		name               string
		targetVersion      string
		scheduledDateTime  string
		ignoreUpdateWindow bool
		responseStatus     int
		wantErr            bool
	}{
		{
			name:               "successful schedule with datetime",
			targetVersion:      "26.1",
			scheduledDateTime:  "2026-04-01T02:00:00Z",
			ignoreUpdateWindow: false,
			responseStatus:     http.StatusOK,
			wantErr:            false,
		},
		{
			name:               "successful schedule without datetime",
			targetVersion:      "26.1",
			scheduledDateTime:  "",
			ignoreUpdateWindow: false,
			responseStatus:     http.StatusNoContent,
			wantErr:            false,
		},
		{
			name:           "server error",
			targetVersion:  "26.1",
			responseStatus: http.StatusInternalServerError,
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
			err := svc.ScheduleUpdateVersion(context.Background(), "BusinessCentral", "production", tt.targetVersion, tt.scheduledDateTime, tt.ignoreUpdateWindow)

			if (err != nil) != tt.wantErr {
				t.Errorf("ScheduleUpdateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_UpdateScheduleDetails(t *testing.T) {
	tests := []struct {
		name               string
		targetVersion      string
		scheduledDateTime  string
		ignoreUpdateWindow bool
		responseStatus     int
		wantErr            bool
	}{
		{
			name:               "successful update",
			targetVersion:      "26.1",
			scheduledDateTime:  "2026-04-01T04:00:00Z",
			ignoreUpdateWindow: true,
			responseStatus:     http.StatusOK,
			wantErr:            false,
		},
		{
			name:           "not found",
			targetVersion:  "26.1",
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
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			svc := NewService(c)
			err := svc.UpdateScheduleDetails(context.Background(), "BusinessCentral", "production", tt.targetVersion, tt.scheduledDateTime, tt.ignoreUpdateWindow)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateScheduleDetails() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
