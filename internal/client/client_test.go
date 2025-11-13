// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package client


import (
	"github.com/vllni/terraform-provider-bcadmincenter/internal/constants"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// mockTokenCredential implements azcore.TokenCredential for testing
type mockTokenCredential struct {
	token string
	err   error
}

func (m *mockTokenCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if m.err != nil {
		return azcore.AccessToken{}, m.err
	}
	return azcore.AccessToken{
		Token: m.token,
	}, nil
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "missing tenant_id",
			config: &Config{
				ClientID:     "test-client-id",
				ClientSecret: "test-secret",
			},
			wantErr: true,
			errMsg:  "tenant_id is required",
		},
		{
			name: "valid config with client secret",
			config: &Config{
				TenantID:     "test-tenant-id",
				ClientID:     "test-client-id",
				ClientSecret: "test-secret",
			},
			wantErr: false,
		},
		{
			name: "valid config with defaults",
			config: &Config{
				TenantID: "test-tenant-id",
				ClientID: "test-client-id",
			},
			wantErr: false,
		},
		{
			name: "custom base URL",
			config: &Config{
				TenantID: "test-tenant-id",
				ClientID: "test-client-id",
				BaseURL:  "https://custom.api.example.com",
			},
			wantErr: false,
		},
		{
			name: "custom API version",
			config: &Config{
				TenantID:   "test-tenant-id",
				ClientID:   "test-client-id",
				APIVersion: "v3.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(context.Background(), tt.config)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewClient() expected error, got nil")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("NewClient() error = %v, want %v", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error = %v", err)
				return
			}

			if client == nil {
				t.Error("NewClient() returned nil client")
				return
			}

			// Check defaults
			if tt.config.BaseURL == "" && client.baseURL != constants.DefaultBaseURL {
				t.Errorf("Client baseURL = %v, want %v", client.baseURL, constants.DefaultBaseURL)
			}
			if tt.config.APIVersion == "" && client.apiVersion != constants.DefaultAPIVersion {
				t.Errorf("Client apiVersion = %v, want %v", client.apiVersion, constants.DefaultAPIVersion)
			}
			if client.tenantID != tt.config.TenantID {
				t.Errorf("Client tenantID = %v, want %v", client.tenantID, tt.config.TenantID)
			}
		})
	}
}

func TestClient_GetToken(t *testing.T) {
	tests := []struct {
		name      string
		mockToken string
		mockErr   error
		wantErr   bool
	}{
		{
			name:      "successful token retrieval",
			mockToken: "test-access-token",
			mockErr:   nil,
			wantErr:   false,
		},
		{
			name:      "token retrieval error",
			mockToken: "",
			mockErr:   errors.New("failed to get token"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				credential: &mockTokenCredential{
					token: tt.mockToken,
					err:   tt.mockErr,
				},
			}

			token, err := client.GetToken(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Error("GetToken() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("GetToken() unexpected error = %v", err)
				return
			}

			if token != tt.mockToken {
				t.Errorf("GetToken() = %v, want %v", token, tt.mockToken)
			}
		})
	}
}

func TestClient_DoRequest(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		responseStatus int
		responseBody   interface{}
		wantErr        bool
		wantErrType    string
	}{
		{
			name:           "successful GET request",
			method:         http.MethodGet,
			path:           "applications/BusinessCentral/environments",
			responseStatus: http.StatusOK,
			responseBody:   map[string]string{"status": "success"},
			wantErr:        false,
		},
		{
			name:           "404 not found",
			method:         http.MethodGet,
			path:           "nonexistent",
			responseStatus: http.StatusNotFound,
			responseBody: AdminCenterError{
				Code:    "NotFound",
				Message: "Resource not found",
			},
			wantErr:     true,
			wantErrType: "*client.AdminCenterError",
		},
		{
			name:           "400 bad request",
			method:         http.MethodPost,
			path:           "applications/BusinessCentral/environments",
			responseStatus: http.StatusBadRequest,
			responseBody: AdminCenterError{
				Code:    "InvalidRequest",
				Message: "Invalid request parameters",
				Target:  "environmentName",
			},
			wantErr:     true,
			wantErrType: "*client.AdminCenterError",
		},
		{
			name:           "500 internal server error",
			method:         http.MethodGet,
			path:           "applications",
			responseStatus: http.StatusInternalServerError,
			responseBody: AdminCenterError{
				Code:    "InternalServerError",
				Message: "An internal error occurred",
			},
			wantErr:     true,
			wantErrType: "*client.AdminCenterError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-token" {
					t.Errorf("Authorization header = %v, want Bearer test-token", authHeader)
				}

				// Verify content-type and accept headers
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Content-Type header = %v, want application/json", r.Header.Get("Content-Type"))
				}
				if r.Header.Get("Accept") != "application/json" {
					t.Errorf("Accept header = %v, want application/json", r.Header.Get("Accept"))
				}

				// Verify request path contains API version and path
				expectedPath := "/admin/" + constants.DefaultAPIVersion + "/" + tt.path
				if r.URL.Path != expectedPath {
					t.Errorf("Request path = %v, want %v", r.URL.Path, expectedPath)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client := &Client{
				credential: &mockTokenCredential{
					token: "test-token",
				},
				httpClient: server.Client(),
				baseURL:    server.URL,
				apiVersion: constants.DefaultAPIVersion,
			}

			resp, err := client.DoRequest(context.Background(), tt.method, tt.path, nil)

			if tt.wantErr {
				if err == nil {
					t.Error("DoRequest() expected error, got nil")
					return
				}
				// Verify error type
				if _, ok := err.(*AdminCenterError); !ok && tt.wantErrType == "*client.AdminCenterError" {
					t.Errorf("DoRequest() error type = %T, want %v", err, tt.wantErrType)
				}
				return
			}

			if err != nil {
				t.Errorf("DoRequest() unexpected error = %v", err)
				return
			}

			if resp == nil {
				t.Error("DoRequest() returned nil response")
				return
			}

			if resp.StatusCode != tt.responseStatus {
				t.Errorf("Response status = %v, want %v", resp.StatusCode, tt.responseStatus)
			}
		})
	}
}

func TestClient_HTTPMethods(t *testing.T) {
	methods := []struct {
		name       string
		methodFunc func(*Client, context.Context, string) (*http.Response, error)
		httpMethod string
	}{
		{
			name: "Get",
			methodFunc: func(c *Client, ctx context.Context, path string) (*http.Response, error) {
				return c.Get(ctx, path)
			},
			httpMethod: http.MethodGet,
		},
		{
			name: "Delete",
			methodFunc: func(c *Client, ctx context.Context, path string) (*http.Response, error) {
				return c.Delete(ctx, path)
			},
			httpMethod: http.MethodDelete,
		},
	}

	for _, tt := range methods {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != tt.httpMethod {
					t.Errorf("Request method = %v, want %v", r.Method, tt.httpMethod)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := &Client{
				credential: &mockTokenCredential{token: "test-token"},
				httpClient: server.Client(),
				baseURL:    server.URL,
				apiVersion: constants.DefaultAPIVersion,
			}

			resp, err := tt.methodFunc(client, context.Background(), "test")
			if err != nil {
				t.Errorf("%s() unexpected error = %v", tt.name, err)
			}
			if resp == nil {
				t.Errorf("%s() returned nil response", tt.name)
			}
		})
	}
}

func TestAdminCenterError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      AdminCenterError
		expected string
	}{
		{
			name: "error with target",
			err: AdminCenterError{
				Code:    "ValidationError",
				Message: "Field is required",
				Target:  "environmentName",
			},
			expected: "ValidationError: Field is required (target: environmentName)",
		},
		{
			name: "error without target",
			err: AdminCenterError{
				Code:    "NotFound",
				Message: "Resource not found",
			},
			expected: "NotFound: Resource not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This should never be called
		t.Error("Server should not be called with cancelled context")
	}))
	defer server.Close()

	client := &Client{
		credential: &mockTokenCredential{token: "test-token"},
		httpClient: server.Client(),
		baseURL:    server.URL,
		apiVersion: constants.DefaultAPIVersion,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.Get(ctx, "test")
	if err == nil {
		t.Error("Expected error with cancelled context, got nil")
	}
}
