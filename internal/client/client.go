// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/constants"
)

// Client represents a Business Central Admin Center API client.
type Client struct {
	credential azcore.TokenCredential
	httpClient *http.Client
	baseURL    string
	tenantID   string
	apiVersion string
}

// Config holds the configuration for creating a new client.
type Config struct {
	ClientID     string
	ClientSecret string
	TenantID     string
	Environment  string
	BaseURL      string
	APIVersion   string
	// AccessToken is a static token used for testing to bypass Azure AD authentication.
	// This should only be set in test environments.
	AccessToken string
}

// staticTokenCredential is a token credential that returns a static pre-obtained token.
// It is intended for use in tests only.
type staticTokenCredential struct {
	token string
}

func (s *staticTokenCredential) GetToken(_ context.Context, _ policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}

// AdminCenterError represents an error response from the Business Central Admin Center API.
type AdminCenterError struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Target     string                 `json:"target,omitempty"`
	Details    []AdminCenterError     `json:"details,omitempty"`
	InnerError map[string]interface{} `json:"innererror,omitempty"`
}

func (e *AdminCenterError) Error() string {
	if e.Target != "" {
		return fmt.Sprintf("%s: %s (target: %s)", e.Code, e.Message, e.Target)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewClient creates a new Business Central Admin Center API client.
func NewClient(ctx context.Context, config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// Initialize credential.
	var credential azcore.TokenCredential
	var err error

	// If a static access token is provided (for testing only), use it directly.
	if config.AccessToken != "" {
		credential = &staticTokenCredential{token: config.AccessToken}
	} else if config.ClientID != "" && config.ClientSecret != "" {
		credential, err = azidentity.NewClientSecretCredential(
			config.TenantID,
			config.ClientID,
			config.ClientSecret,
			&azidentity.ClientSecretCredentialOptions{
				AdditionallyAllowedTenants: []string{"*"},
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create client secret credential: %w", err)
		}
	} else {
		// Otherwise, use DefaultAzureCredential for other auth methods.
		// Pass the tenant ID to ensure it's used for Azure CLI, Azure Developer CLI, and workload identity.
		credential, err = azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
			TenantID:                   config.TenantID,
			AdditionallyAllowedTenants: []string{"*"},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create default credential: %w", err)
		}
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = constants.DefaultBaseURL
	}

	apiVersion := config.APIVersion
	if apiVersion == "" {
		apiVersion = constants.DefaultAPIVersion
	}

	client := &Client{
		credential: credential,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:    baseURL,
		tenantID:   config.TenantID,
		apiVersion: apiVersion,
	}

	return client, nil
}

// GetToken retrieves an access token for the Business Central Admin Center API.
func (c *Client) GetToken(ctx context.Context) (string, error) {
	token, err := c.credential.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{fmt.Sprintf("%s/.default", constants.BusinessCentralResourceID)},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}
	return token.Token, nil
}

// DoRequest performs an authenticated HTTP request to the Business Central Admin Center API.
func (c *Client) DoRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	// Get authentication token.
	token, err := c.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	// Build request URL.
	url := fmt.Sprintf("%s/admin/%s/%s", c.baseURL, c.apiVersion, path)

	// Create request.
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers.
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Check for error responses.
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		var apiError AdminCenterError
		if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
			return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, resp.Status)
		}

		return nil, &apiError
	}

	return resp, nil
}

// Get performs an authenticated GET request.
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodGet, path, nil)
}

// Post performs an authenticated POST request.
func (c *Client) Post(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodPost, path, body)
}

// Put performs an authenticated PUT request.
func (c *Client) Put(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodPut, path, body)
}

// Delete performs an authenticated DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodDelete, path, nil)
}

// Patch performs an authenticated PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body io.Reader) (*http.Response, error) {
	return c.DoRequest(ctx, http.MethodPatch, path, body)
}

// DoAutomationRequest performs an authenticated HTTP request to the Business Central Automation API.
// The Automation API uses a different base URL pattern:
// {baseURL}/v2.0/{environmentName}/api/microsoft/automation/v2.0/{path}
// contentType overrides the default "application/json" when non-empty.
// extraHeaders contains additional headers (e.g. If-Match for PATCH requests).
func (c *Client) DoAutomationRequest(ctx context.Context, method, environmentName, path string, body io.Reader, contentType string, extraHeaders map[string]string) (*http.Response, error) {
	// Get authentication token.
	token, err := c.GetToken(ctx)
	if err != nil {
		return nil, err
	}

	// Build Automation API URL.
	url := fmt.Sprintf("%s/v2.0/%s/api/microsoft/automation/v2.0/%s", c.baseURL, environmentName, path)

	// Create request.
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create automation request: %w", err)
	}

	// Set headers.
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	// Execute request.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute automation request: %w", err)
	}

	// Check for error responses.
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		var apiError AdminCenterError
		if err := json.NewDecoder(resp.Body).Decode(&apiError); err != nil {
			return nil, fmt.Errorf("automation API returned status %d: %s", resp.StatusCode, resp.Status)
		}

		return nil, &apiError
	}

	return resp, nil
}

// SetCredential sets the credential for testing purposes.
func (c *Client) SetCredential(credential azcore.TokenCredential) {
	c.credential = credential
}

// SetBaseURL sets the base URL for testing purposes.
func (c *Client) SetBaseURL(baseURL string) {
	c.baseURL = baseURL
}

// SetAPIVersion sets the API version for testing purposes.
func (c *Client) SetAPIVersion(apiVersion string) {
	c.apiVersion = apiVersion
}

// SetHTTPClient sets the HTTP client for testing purposes.
func (c *Client) SetHTTPClient(httpClient *http.Client) {
	c.httpClient = httpClient
}

// GetTenantID returns the configured tenant ID.
func (c *Client) GetTenantID() string {
	return c.tenantID
}

// ForTenant returns a new Client that authenticates against the specified tenant.
// When aad_tenant_id is set to a tenant other than the provider's configured tenant_id,
// use this method to ensure API calls are directed to the correct tenant.
// The underlying credential must support multi-tenant access (AdditionallyAllowedTenants).
func (c *Client) ForTenant(tenantID string) *Client {
	if tenantID == "" || tenantID == c.tenantID {
		return c
	}
	return &Client{
		credential: &tenantOverrideCredential{
			base:     c.credential,
			tenantID: tenantID,
		},
		httpClient: c.httpClient,
		baseURL:    c.baseURL,
		tenantID:   tenantID,
		apiVersion: c.apiVersion,
	}
}

// tenantOverrideCredential wraps an azcore.TokenCredential to request tokens for a specific tenant.
type tenantOverrideCredential struct {
	base     azcore.TokenCredential
	tenantID string
}

func (t *tenantOverrideCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	options.TenantID = t.tenantID
	return t.base.GetToken(ctx, options)
}
