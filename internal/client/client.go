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

	"github.com/vllni/terraform-provider-bcadmincenter/internal/constants"
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

	// If ClientID and ClientSecret are provided, use ClientSecretCredential.
	if config.ClientID != "" && config.ClientSecret != "" {
		credential, err = azidentity.NewClientSecretCredential(
			config.TenantID,
			config.ClientID,
			config.ClientSecret,
			nil,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create client secret credential: %w", err)
		}
	} else {
		// Otherwise, use DefaultAzureCredential for other auth methods.
		// Pass the tenant ID to ensure it's used for Azure CLI, Azure Developer CLI, and workload identity.
		credential, err = azidentity.NewDefaultAzureCredential(&azidentity.DefaultAzureCredentialOptions{
			TenantID: config.TenantID,
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
