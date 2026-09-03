// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
)

// mockTokenCredential implements azcore.TokenCredential for testing.
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
		{
			name: "OIDC with static token",
			config: &Config{
				TenantID:  "test-tenant-id",
				ClientID:  "test-client-id",
				UseOIDC:   true,
				OIDCToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.test.sig",
			},
			wantErr: false,
		},
		{
			name: "OIDC with static token but missing client_id",
			config: &Config{
				TenantID:  "test-tenant-id",
				UseOIDC:   true,
				OIDCToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.test.sig",
			},
			wantErr: true,
			errMsg:  "client_id is required for OIDC authentication",
		},
		{
			name: "OIDC workload identity with token file path",
			config: &Config{
				TenantID:          "test-tenant-id",
				ClientID:          "test-client-id",
				UseOIDC:           true,
				OIDCTokenFilePath: "/var/run/secrets/token",
			},
			wantErr: false,
		},
		{
			name: "OIDC implied by oidc_token without use_oidc flag",
			config: &Config{
				TenantID:  "test-tenant-id",
				ClientID:  "test-client-id",
				OIDCToken: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.test.sig",
			},
			wantErr: false,
		},
		{
			name: "OIDC with GitHub Actions request URL",
			config: &Config{
				TenantID:         "test-tenant-id",
				ClientID:         "test-client-id",
				UseOIDC:          true,
				OIDCRequestURL:   "https://token.actions.githubusercontent.com/token",
				OIDCRequestToken: "gha-bearer-token",
			},
			wantErr: false,
		},
		{
			name: "OIDC with GitHub Actions request URL but missing client_id",
			config: &Config{
				TenantID:       "test-tenant-id",
				UseOIDC:        true,
				OIDCRequestURL: "https://token.actions.githubusercontent.com/token",
			},
			wantErr: true,
			errMsg:  "client_id is required for OIDC authentication",
		},
		{
			name: "OIDC with ADO pipeline service connection",
			config: &Config{
				TenantID:                       "test-tenant-id",
				ClientID:                       "test-client-id",
				UseOIDC:                        true,
				OIDCRequestURL:                 "https://dev.azure.com/org/_apis/distributedtask/hubs/build/plans/plan-id/jobs/job-id/oidctoken",
				OIDCRequestToken:               "ado-system-access-token",
				ADOPipelineServiceConnectionID: "service-conn-id",
			},
			wantErr: false,
		},
		{
			name: "OIDC use_oidc=true with no token source",
			config: &Config{
				TenantID: "test-tenant-id",
				ClientID: "test-client-id",
				UseOIDC:  true,
			},
			wantErr: true,
			errMsg:  "OIDC authentication requires one of: oidc_token, oidc_request_url, oidc_token_file_path, or AZURE_FEDERATED_TOKEN_FILE",
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

			// Check defaults.
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
			// Create test server.
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify authorization header.
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-token" {
					t.Errorf("Authorization header = %v, want Bearer test-token", authHeader)
				}

				// Verify content-type and accept headers.
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Content-Type header = %v, want application/json", r.Header.Get("Content-Type"))
				}
				if r.Header.Get("Accept") != "application/json" {
					t.Errorf("Accept header = %v, want application/json", r.Header.Get("Accept"))
				}

				// Verify request path contains API version and path.
				expectedPath := "/admin/" + constants.DefaultAPIVersion + "/" + tt.path
				if r.URL.Path != expectedPath {
					t.Errorf("Request path = %v, want %v", r.URL.Path, expectedPath)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.responseStatus)
				if err := json.NewEncoder(w).Encode(tt.responseBody); err != nil {

					t.Fatalf("Failed to encode response: %v", err)

				}
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
				// Verify error type.
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
		// This should never be called.
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

func TestClient_ForTenant(t *testing.T) {
	base := &Client{
		credential: &mockTokenCredential{token: "test-token"},
		httpClient: &http.Client{},
		baseURL:    "https://api.example.com",
		tenantID:   "original-tenant-id",
		apiVersion: constants.DefaultAPIVersion,
	}

	t.Run("same tenant returns original client", func(t *testing.T) {
		result := base.ForTenant("original-tenant-id")
		if result != base {
			t.Error("ForTenant() with same tenant ID should return the original client")
		}
	})

	t.Run("empty tenant returns original client", func(t *testing.T) {
		result := base.ForTenant("")
		if result != base {
			t.Error("ForTenant() with empty tenant ID should return the original client")
		}
	})

	t.Run("different tenant returns new client", func(t *testing.T) {
		result := base.ForTenant("other-tenant-id")
		if result == base {
			t.Error("ForTenant() with different tenant ID should return a new client")
		}
		if result.GetTenantID() != "other-tenant-id" {
			t.Errorf("ForTenant() tenant ID = %s, want other-tenant-id", result.GetTenantID())
		}
		if result.baseURL != base.baseURL {
			t.Errorf("ForTenant() baseURL = %s, want %s", result.baseURL, base.baseURL)
		}
		if result.apiVersion != base.apiVersion {
			t.Errorf("ForTenant() apiVersion = %s, want %s", result.apiVersion, base.apiVersion)
		}
		if result.httpClient != base.httpClient {
			t.Error("ForTenant() should reuse the same HTTP client")
		}
	})

	t.Run("per-tenant credential overrides tenant ID in token request", func(t *testing.T) {
		capturedTenantID := ""
		capturingCred := &capturingTokenCredential{
			onGetToken: func(opts policy.TokenRequestOptions) {
				capturedTenantID = opts.TenantID
			},
		}
		base2 := &Client{
			credential: capturingCred,
			httpClient: &http.Client{},
			baseURL:    "https://api.example.com",
			tenantID:   "original-tenant",
			apiVersion: constants.DefaultAPIVersion,
		}
		forked := base2.ForTenant("other-tenant")
		_, _ = forked.GetToken(context.Background())
		if capturedTenantID != "other-tenant" {
			t.Errorf("per-tenant credential passed TenantID = %s, want other-tenant", capturedTenantID)
		}
	})
}

// capturingTokenCredential captures TokenRequestOptions for test assertions.
type capturingTokenCredential struct {
	onGetToken func(policy.TokenRequestOptions)
}

func (c *capturingTokenCredential) GetToken(_ context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if c.onGetToken != nil {
		c.onGetToken(opts)
	}
	return azcore.AccessToken{Token: "captured-token"}, nil
}

func TestBuildOIDCAssertionCallback(t *testing.T) {
	t.Run("static token", func(t *testing.T) {
		cb, err := buildOIDCAssertionCallback(&Config{OIDCToken: "static-jwt"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, err := cb(context.Background())
		if err != nil || got != "static-jwt" {
			t.Errorf("callback() = %q, %v; want %q, nil", got, err, "static-jwt")
		}
	})

	t.Run("token file", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "oidc-token-*")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.WriteString("  file-jwt\n"); err != nil {
			t.Fatal(err)
		}
		f.Close()

		cb, err := buildOIDCAssertionCallback(&Config{OIDCTokenFilePath: f.Name()})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, err := cb(context.Background())
		if err != nil || got != "file-jwt" {
			t.Errorf("callback() = %q, %v; want %q, nil", got, err, "file-jwt")
		}

		// Simulate token rotation: overwrite file and check new value is returned.
		if err := os.WriteFile(f.Name(), []byte("rotated-jwt"), 0o600); err != nil {
			t.Fatal(err)
		}
		got2, err := cb(context.Background())
		if err != nil || got2 != "rotated-jwt" {
			t.Errorf("after rotation callback() = %q, %v; want %q, nil", got2, err, "rotated-jwt")
		}
	})

	t.Run("token file via AZURE_FEDERATED_TOKEN_FILE env", func(t *testing.T) {
		f, err := os.CreateTemp(t.TempDir(), "oidc-token-*")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.WriteString("env-jwt"); err != nil {
			t.Fatal(err)
		}
		f.Close()
		t.Setenv("AZURE_FEDERATED_TOKEN_FILE", f.Name())

		cb, err := buildOIDCAssertionCallback(&Config{UseOIDC: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, err := cb(context.Background())
		if err != nil || got != "env-jwt" {
			t.Errorf("callback() = %q, %v; want %q, nil", got, err, "env-jwt")
		}
	})

	t.Run("no source returns error", func(t *testing.T) {
		_, err := buildOIDCAssertionCallback(&Config{UseOIDC: true})
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
}

func TestBuildGitHubOIDCCallback(t *testing.T) {
	t.Run("fetches and returns token value", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer test-bearer" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if r.URL.Query().Get("audience") != "api://AzureADTokenExchange" {
				http.Error(w, "missing audience", http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"value":"fresh-oidc-jwt"}`)); err != nil {
				t.Errorf("writing response: %v", err)
			}
		}))
		defer server.Close()

		cb := buildGitHubOIDCCallback(server.URL+"/token", "test-bearer")
		got, err := cb(context.Background())
		if err != nil || got != "fresh-oidc-jwt" {
			t.Errorf("callback() = %q, %v; want %q, nil", got, err, "fresh-oidc-jwt")
		}
	})

	t.Run("appends audience to URL with existing query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("audience") != "api://AzureADTokenExchange" || q.Get("existing") != "param" {
				http.Error(w, "bad query: "+r.URL.RawQuery, http.StatusBadRequest)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"value":"jwt"}`)); err != nil {
				t.Errorf("writing response: %v", err)
			}
		}))
		defer server.Close()

		cb := buildGitHubOIDCCallback(server.URL+"/token?existing=param", "")
		if _, err := cb(context.Background()); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("non-200 response returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "forbidden", http.StatusForbidden)
		}))
		defer server.Close()

		cb := buildGitHubOIDCCallback(server.URL+"/token", "bad-token")
		if _, err := cb(context.Background()); err == nil {
			t.Error("expected error for non-200 response, got nil")
		}
	})

	t.Run("empty value field returns error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"value":""}`)); err != nil {
				t.Errorf("writing response: %v", err)
			}
		}))
		defer server.Close()

		cb := buildGitHubOIDCCallback(server.URL+"/token", "")
		if _, err := cb(context.Background()); err == nil {
			t.Error("expected error for empty value, got nil")
		}
	})
}

func TestClient_DoRequestWithOptions(t *testing.T) {
	newClient := func(baseURL string, timeout time.Duration) *Client {
		c := &Client{}
		c.SetCredential(&mockTokenCredential{token: "test-token"})
		c.SetBaseURL(baseURL)
		c.SetAPIVersion(constants.DefaultAPIVersion)
		c.SetHTTPClient(&http.Client{Timeout: timeout})
		return c
	}

	t.Run("nil options keep the JSON defaults", func(t *testing.T) {
		var gotContentType, gotAccept, gotAuth string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotContentType = r.Header.Get("Content-Type")
			gotAccept = r.Header.Get("Accept")
			gotAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := newClient(server.URL, 30*time.Second).DoRequestWithOptions(context.Background(), http.MethodGet, "things", nil, nil)
		if err != nil {
			t.Fatalf("DoRequestWithOptions() unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if gotContentType != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", gotContentType)
		}
		if gotAccept != "application/json" {
			t.Errorf("Accept = %q, want application/json", gotAccept)
		}
		if gotAuth != "Bearer test-token" {
			t.Errorf("Authorization = %q, want Bearer test-token", gotAuth)
		}
	})

	t.Run("content type and extra headers are applied", func(t *testing.T) {
		var gotContentType, gotCustom string

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotContentType = r.Header.Get("Content-Type")
			gotCustom = r.Header.Get("If-Match")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		resp, err := newClient(server.URL, 30*time.Second).DoRequestWithOptions(
			context.Background(), http.MethodPatch, "things", strings.NewReader("payload"),
			&RequestOptions{ContentType: "application/octet-stream", Headers: map[string]string{"If-Match": "*"}},
		)
		if err != nil {
			t.Fatalf("DoRequestWithOptions() unexpected error: %v", err)
		}
		defer resp.Body.Close()

		if gotContentType != "application/octet-stream" {
			t.Errorf("Content-Type = %q, want application/octet-stream", gotContentType)
		}
		if gotCustom != "*" {
			t.Errorf("If-Match = %q, want *", gotCustom)
		}
	})

	t.Run("per-request timeout overrides a shorter client timeout", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(150 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// The client's own timeout is far too short for this handler.
		c := newClient(server.URL, 20*time.Millisecond)

		if _, err := c.DoRequest(context.Background(), http.MethodGet, "things", nil); err == nil {
			t.Error("DoRequest() expected a timeout error with the short client timeout, got nil")
		}

		resp, err := c.DoRequestWithOptions(context.Background(), http.MethodGet, "things", nil, &RequestOptions{Timeout: 5 * time.Second})
		if err != nil {
			t.Fatalf("DoRequestWithOptions() with an extended timeout returned an error: %v", err)
		}
		defer resp.Body.Close()

		// The override must not leak back into the shared client.
		if c.httpClient.Timeout != 20*time.Millisecond {
			t.Errorf("client timeout = %v, want it left at 20ms", c.httpClient.Timeout)
		}
	})

	t.Run("error responses are decoded into AdminCenterError", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"code": "EntityValidationFailed", "message": "bad"})
		}))
		defer server.Close()

		_, err := newClient(server.URL, 30*time.Second).DoRequestWithOptions(context.Background(), http.MethodPost, "things", nil, &RequestOptions{})

		var apiErr *AdminCenterError
		if !errors.As(err, &apiErr) {
			t.Fatalf("error = %v, want an *AdminCenterError", err)
		}
		if apiErr.Code != "EntityValidationFailed" {
			t.Errorf("code = %q, want EntityValidationFailed", apiErr.Code)
		}
	})
}

func TestClient_PostMultipart(t *testing.T) {
	var gotMethod, gotPath, gotContentType, gotBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := &Client{}
	c.SetCredential(&mockTokenCredential{token: "test-token"})
	c.SetBaseURL(server.URL)
	c.SetAPIVersion(constants.DefaultAPIVersion)
	c.SetHTTPClient(&http.Client{Timeout: 30 * time.Second})

	resp, err := c.PostMultipart(context.Background(), "applications/BusinessCentral/environments/Prod/apps/pteInstall",
		strings.NewReader("--boundary--\r\n"), "multipart/form-data; boundary=boundary", 5*time.Minute)
	if err != nil {
		t.Fatalf("PostMultipart() unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	wantPath := "/admin/" + constants.DefaultAPIVersion + "/applications/BusinessCentral/environments/Prod/apps/pteInstall"
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}
	if gotContentType != "multipart/form-data; boundary=boundary" {
		t.Errorf("Content-Type = %q, want the boundary-carrying multipart type", gotContentType)
	}
	if !strings.Contains(gotBody, "boundary") {
		t.Errorf("body = %q, want the multipart payload to be forwarded", gotBody)
	}
}

// TestAdminCenterError_StatusCode covers the contract that error responses always carry
// their HTTP status, regardless of whether the body uses the documented {code, message}
// envelope. Without this, callers are forced to string-match the rendered message to
// detect a 404, which misfires on unrelated errors and misses non-conforming bodies.
func TestAdminCenterError_StatusCode(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantCode    string
		wantMessage string
		wantErrText string
		wantNotFnd  bool
	}{
		{
			name:        "documented envelope",
			status:      http.StatusBadRequest,
			body:        `{"code":"ValidationError","message":"bad request"}`,
			wantCode:    "ValidationError",
			wantMessage: "bad request",
			wantErrText: "ValidationError: bad request",
			wantNotFnd:  false,
		},
		{
			// The API answers some endpoints with this shape. encoding/json ignores the
			// unknown field, so Code and Message stay empty and the old renderer produced
			// the useless string ": ".
			name:        "non-conforming body still reports status and body",
			status:      http.StatusNotFound,
			body:        `{"error":"tenant not found"}`,
			wantCode:    "",
			wantMessage: "",
			wantErrText: `API returned status 404: {"error":"tenant not found"}`,
			wantNotFnd:  true,
		},
		{
			name:        "empty body falls back to the status line",
			status:      http.StatusNotFound,
			body:        "",
			wantErrText: "API returned status 404: 404 Not Found",
			wantNotFnd:  true,
		},
		{
			name:        "not-found code on a non-404 status",
			status:      http.StatusBadRequest,
			body:        `{"code":"EnvironmentNotFound","message":"no such environment"}`,
			wantCode:    "EnvironmentNotFound",
			wantMessage: "no such environment",
			wantErrText: "EnvironmentNotFound: no such environment",
			wantNotFnd:  false, // IsNotFound is status-based only; code checks are the caller's job.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				if tt.body != "" {
					if _, err := w.Write([]byte(tt.body)); err != nil {
						t.Errorf("failed to write response body: %v", err)
					}
				}
			}))
			defer server.Close()

			c := &Client{}
			c.SetCredential(&mockTokenCredential{token: "test-token"})
			c.SetBaseURL(server.URL)
			c.SetAPIVersion(constants.DefaultAPIVersion)
			c.SetHTTPClient(&http.Client{})

			resp, err := c.Get(context.Background(), "some/path")
			if resp != nil {
				t.Fatalf("expected a nil response on error, got %v", resp)
			}
			if err == nil {
				t.Fatal("expected an error")
			}

			var apiErr *AdminCenterError
			if !errors.As(err, &apiErr) {
				t.Fatalf("expected *AdminCenterError, got %T: %v", err, err)
			}
			if apiErr.StatusCode != tt.status {
				t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, tt.status)
			}
			if apiErr.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", apiErr.Code, tt.wantCode)
			}
			if apiErr.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", apiErr.Message, tt.wantMessage)
			}
			if got := err.Error(); got != tt.wantErrText {
				t.Errorf("Error() = %q, want %q", got, tt.wantErrText)
			}
			if got := IsNotFound(err); got != tt.wantNotFnd {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.wantNotFnd)
			}
			// Detection must survive wrapping, which every service layer does.
			if got := IsNotFound(fmt.Errorf("wrapped: %w", err)); got != tt.wantNotFnd {
				t.Errorf("IsNotFound(wrapped) = %v, want %v", got, tt.wantNotFnd)
			}
		})
	}
}

func TestIsNotFound_NonAPIErrors(t *testing.T) {
	if IsNotFound(nil) {
		t.Error("IsNotFound(nil) = true, want false")
	}
	// Message text must not drive the decision: "404" shows up in addresses and versions.
	if IsNotFound(fmt.Errorf("dial tcp 10.0.0.404:443: connection refused")) {
		t.Error("IsNotFound(untyped error containing 404) = true, want false")
	}
}

// TestNewClient_StaticTokenGate pins the security boundary added for the
// BCADMINCENTER_TEST_TOKEN backdoor: a static bearer token bypasses Azure AD entirely
// and must only be honoured in builds carrying the bcadmincenter_testing tag.
func TestNewClient_StaticTokenGate(t *testing.T) {
	c, err := NewClient(context.Background(), &Config{
		TenantID:    "00000000-0000-0000-0000-000000000000",
		AccessToken: "a-static-token",
	})

	if testingBuild {
		if err != nil {
			t.Fatalf("testing build should accept a static access token, got %v", err)
		}
		if c == nil {
			t.Fatal("expected a client")
		}
		return
	}

	if err == nil {
		t.Fatal("release build must refuse a static access token, got a usable client")
	}
	if !strings.Contains(err.Error(), "bcadmincenter_testing") {
		t.Errorf("error should name the build tag so the cause is actionable, got: %v", err)
	}
}

// TestValidateBaseURL guards against sending an Azure AD access token to an untrusted or
// plaintext endpoint: DoRequest attaches the Authorization header before it inspects the
// destination, so a bad base_url leaks a live credential.
func TestValidateBaseURL(t *testing.T) {
	alwaysValid := []string{
		"https://api.businesscentral.dynamics.com",
		"https://example.internal:8443/prefix",
	}
	for _, raw := range alwaysValid {
		t.Run("valid/"+raw, func(t *testing.T) {
			if err := validateBaseURL(raw); err != nil {
				t.Errorf("validateBaseURL(%q) = %v, want nil", raw, err)
			}
		})
	}

	alwaysInvalid := []struct{ name, raw string }{
		{"no scheme or host", "api.businesscentral.dynamics.com"},
		{"scheme without host", "https://"},
		{"unsupported scheme", "ftp://example.com"},
		{"control characters", "https://exa\x7fmple.com"},
	}
	for _, tt := range alwaysInvalid {
		t.Run("invalid/"+tt.name, func(t *testing.T) {
			if err := validateBaseURL(tt.raw); err == nil {
				t.Errorf("validateBaseURL(%q) = nil, want an error", tt.raw)
			}
		})
	}

	// http:// is the interesting case: test servers need it, release builds must not.
	t.Run("plaintext http", func(t *testing.T) {
		err := validateBaseURL("http://127.0.0.1:8080")
		if testingBuild {
			if err != nil {
				t.Errorf("testing build should allow http for local mock servers, got %v", err)
			}
			return
		}
		if err == nil {
			t.Error("release build must reject a plaintext http base_url")
		}
	})
}

// TestNewClient_RejectsPlaintextBaseURL checks the validation is actually wired into
// NewClient, not just available as a helper.
func TestNewClient_RejectsPlaintextBaseURL(t *testing.T) {
	if testingBuild {
		t.Skip("http:// is intentionally permitted in testing builds")
	}
	_, err := NewClient(context.Background(), &Config{
		TenantID: "00000000-0000-0000-0000-000000000000",
		ClientID: "client", ClientSecret: "secret",
		BaseURL: "http://attacker.example.com",
	})
	if err == nil {
		t.Fatal("NewClient accepted a plaintext base_url")
	}
	if !strings.Contains(err.Error(), "https") {
		t.Errorf("error should explain the https requirement, got: %v", err)
	}
}
