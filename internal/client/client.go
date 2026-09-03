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
	"net/url"
	"os"
	"strings"
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
	// UseOIDC forces Workload Identity / federated credential authentication.
	// Equivalent to setting AZURE_USE_OIDC=true.
	UseOIDC bool
	// OIDCToken is a static JWT bearer token used as the OIDC client assertion.
	// Setting this implies UseOIDC=true.
	OIDCToken string
	// OIDCTokenFilePath is the path to a file containing the federated OIDC token.
	// Falls back to AZURE_FEDERATED_TOKEN_FILE when empty.
	OIDCTokenFilePath string
	// OIDCRequestURL is the URL of the OIDC token endpoint (e.g. the GitHub Actions
	// OIDC endpoint provided via ACTIONS_ID_TOKEN_REQUEST_URL). A fresh token is
	// fetched from this endpoint on every Azure AD token refresh, which prevents
	// failures caused by short-lived OIDC JWTs expiring mid-run.
	OIDCRequestURL string
	// OIDCRequestToken is the bearer token used to authenticate requests to
	// OIDCRequestURL (e.g. ACTIONS_ID_TOKEN_REQUEST_TOKEN in GitHub Actions).
	OIDCRequestToken string
	// ADOPipelineServiceConnectionID is the Azure DevOps service connection ID used
	// when authenticating via ADO Pipeline OIDC (SYSTEM_OIDCREQUESTURI / SYSTEM_ACCESSTOKEN).
	// When set together with OIDCRequestURL, the ADO OIDC endpoint is used instead of GitHub.
	ADOPipelineServiceConnectionID string
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
//
// StatusCode, Status and RawBody are populated from the HTTP response rather than the
// JSON body. The API does not always return the documented {code, message} envelope —
// some endpoints answer with shapes like {"error": "..."} — and encoding/json ignores
// unknown fields, so Code and Message can both be empty for a perfectly valid error
// response. Callers must therefore branch on StatusCode (see IsNotFound) rather than
// pattern-matching Code or the rendered message.
type AdminCenterError struct {
	// StatusCode is the HTTP status of the response that produced this error.
	StatusCode int `json:"-"`
	// Status is the HTTP status line (e.g. "404 Not Found").
	Status string `json:"-"`
	// RawBody is the trimmed response body, retained so a non-conforming error body
	// still produces a useful diagnostic instead of an empty "code: message".
	RawBody string `json:"-"`

	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Target     string                 `json:"target,omitempty"`
	Details    []AdminCenterError     `json:"details,omitempty"`
	InnerError map[string]interface{} `json:"innererror,omitempty"`
}

func (e *AdminCenterError) Error() string {
	if e.Code == "" && e.Message == "" {
		detail := e.RawBody
		if detail == "" {
			detail = e.Status
		}
		return fmt.Sprintf("API returned status %d: %s", e.StatusCode, detail)
	}
	if e.Target != "" {
		return fmt.Sprintf("%s: %s (target: %s)", e.Code, e.Message, e.Target)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// IsNotFound reports whether err is (or wraps) an Admin Center API error carrying
// HTTP 404. Prefer this over inspecting Code or the message text.
func IsNotFound(err error) bool {
	var apiErr *AdminCenterError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound
}

// newAdminCenterError builds a typed error from a >= 400 response. The status is always
// carried, even when the body is empty or does not use the documented envelope, so
// callers can detect a 404 without string-matching. The caller closes resp.Body.
func newAdminCenterError(resp *http.Response) *AdminCenterError {
	apiError := &AdminCenterError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err == nil && len(body) > 0 {
		// Unknown fields are ignored, so a non-conforming body leaves Code/Message
		// empty rather than failing; RawBody keeps the text for the diagnostic.
		_ = json.Unmarshal(body, apiError)
		apiError.RawBody = strings.TrimSpace(string(body))
	}
	return apiError
}

// BuildPath assembles an Admin Center API path from individual segments, percent-escaping
// each one.
//
// Callers pass values that come from Terraform configuration — environment names,
// application families, app ids, target versions. Interpolating those into a path with
// fmt.Sprintf lets a value change the request's structure rather than just its content: a
// "?" starts the query string, a "#" starts a fragment that is never transmitted, and a
// "/" or "../" re-targets the request at a different resource. Each of those turns a
// request meant for one environment into a silently successful request against another.
//
// Literal segments may be passed alongside dynamic ones; escaping is a no-op for them.
func BuildPath(segments ...string) string {
	escaped := make([]string, len(segments))
	for i, segment := range segments {
		escaped[i] = url.PathEscape(segment)
	}
	return strings.Join(escaped, "/")
}

// validateBaseURL checks a caller-supplied base URL before it is used.
//
// Every request built from this URL carries a live Azure AD bearer token in an
// Authorization header, and that header is attached before the destination is
// examined. A plaintext or malformed base URL therefore leaks a usable credential, so
// http:// is accepted only in testing builds where the target is a local test server.
func validateBaseURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid base_url %q: %w", raw, err)
	}
	if parsed.Host == "" {
		return fmt.Errorf("invalid base_url %q: must be an absolute URL including a host", raw)
	}
	if parsed.Scheme == "https" {
		return nil
	}
	if testingBuild && parsed.Scheme == "http" {
		return nil
	}
	return fmt.Errorf("invalid base_url %q: scheme must be https, because every request to it "+
		"carries an Azure AD access token", raw)
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

	// A static access token bypasses Azure AD entirely, so it is refused unless this
	// binary was built for testing. Failing loudly beats silently ignoring the value:
	// a release provider that quietly discarded the token would authenticate as
	// somebody else without saying so.
	if config.AccessToken != "" && !testingBuild {
		return nil, fmt.Errorf("a static access token was supplied (BCADMINCENTER_TEST_TOKEN), " +
			"but static tokens bypass Azure AD authentication and are only honoured in builds " +
			"tagged 'bcadmincenter_testing'")
	}

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
	} else if config.UseOIDC || config.OIDCToken != "" || config.OIDCTokenFilePath != "" || config.OIDCRequestURL != "" {
		// Workload Identity / OIDC (federated credential) authentication.
		// All OIDC variants use ClientAssertionCredential with a callback so the
		// Azure SDK can obtain a fresh assertion on every token refresh, preventing
		// failures caused by short-lived OIDC JWTs expiring during long runs.
		if config.ClientID == "" {
			return nil, fmt.Errorf("client_id is required for OIDC authentication")
		}
		callback, cbErr := buildOIDCAssertionCallback(config)
		if cbErr != nil {
			return nil, cbErr
		}
		credential, err = azidentity.NewClientAssertionCredential(
			config.TenantID,
			config.ClientID,
			callback,
			&azidentity.ClientAssertionCredentialOptions{
				AdditionallyAllowedTenants: []string{"*"},
			},
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OIDC credential: %w", err)
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
	} else if err := validateBaseURL(baseURL); err != nil {
		return nil, err
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

// buildOIDCAssertionCallback returns the assertion callback used by ClientAssertionCredential.
// Source priority:
//  1. Static oidc_token – returned as-is (caller is responsible for keeping it valid).
//  2. GitHub Actions OIDC endpoint – a fresh JWT is fetched on every invocation.
//  3. Token file (oidc_token_file_path / AZURE_FEDERATED_TOKEN_FILE) – re-read on every
//     invocation so that token rotation by the platform is picked up automatically.
func buildOIDCAssertionCallback(config *Config) (func(context.Context) (string, error), error) {
	switch {
	case config.OIDCToken != "":
		return func(_ context.Context) (string, error) {
			return config.OIDCToken, nil
		}, nil

	case config.OIDCRequestURL != "":
		if config.ADOPipelineServiceConnectionID != "" {
			return buildADOOIDCCallback(config.OIDCRequestURL, config.OIDCRequestToken, config.ADOPipelineServiceConnectionID), nil
		}
		return buildGitHubOIDCCallback(config.OIDCRequestURL, config.OIDCRequestToken), nil

	default:
		// File-based: resolve path now (env var is typically static), re-read contents
		// on every invocation so a rotating token file is always current.
		filePath := config.OIDCTokenFilePath
		if filePath == "" {
			filePath = os.Getenv("AZURE_FEDERATED_TOKEN_FILE")
		}
		if filePath == "" {
			return nil, fmt.Errorf("OIDC authentication requires one of: oidc_token, oidc_request_url, oidc_token_file_path, or AZURE_FEDERATED_TOKEN_FILE")
		}
		return func(_ context.Context) (string, error) {
			data, readErr := os.ReadFile(filePath)
			if readErr != nil {
				return "", fmt.Errorf("reading OIDC token file %q: %w", filePath, readErr)
			}
			return strings.TrimSpace(string(data)), nil
		}, nil
	}
}

// buildGitHubOIDCCallback returns a callback that fetches a fresh OIDC JWT from the
// GitHub Actions (or compatible) OIDC endpoint on every invocation.
// The audience parameter required by Azure AD token exchange is appended automatically.
func buildGitHubOIDCCallback(requestURL, bearerToken string) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		parsedURL, parseErr := url.Parse(requestURL)
		if parseErr != nil {
			return "", fmt.Errorf("parsing GitHub OIDC request URL: %w", parseErr)
		}
		query, _ := url.ParseQuery(parsedURL.RawQuery)
		if query.Get("audience") == "" {
			query.Set("audience", "api://AzureADTokenExchange")
			parsedURL.RawQuery = query.Encode()
		}

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, parsedURL.String(), nil)
		if reqErr != nil {
			return "", fmt.Errorf("building GitHub OIDC token request: %w", reqErr)
		}
		req.Header.Set("Authorization", "Bearer "+bearerToken)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, doErr := http.DefaultClient.Do(req)
		if doErr != nil {
			return "", fmt.Errorf("fetching GitHub OIDC token: %w", doErr)
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if readErr != nil {
			return "", fmt.Errorf("reading GitHub OIDC token response: %w", readErr)
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return "", fmt.Errorf("GitHub OIDC token request failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var result struct {
			Count *int    `json:"count"`
			Value *string `json:"value"`
		}
		if decodeErr := json.Unmarshal(body, &result); decodeErr != nil {
			return "", fmt.Errorf("decoding GitHub OIDC token response: %w", decodeErr)
		}
		if result.Value == nil || *result.Value == "" {
			return "", fmt.Errorf("GitHub OIDC token response contained empty value field")
		}
		return *result.Value, nil
	}
}

// buildADOOIDCCallback returns a callback that fetches a fresh OIDC JWT from an Azure DevOps
// Pipeline OIDC endpoint on every invocation. The service connection ID and required
// query parameters are appended automatically, matching the behaviour of go-azure-sdk's
// ADOPipelineOIDCAuthorizer.
func buildADOOIDCCallback(requestURL, bearerToken, serviceConnectionID string) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		parsedURL, parseErr := url.Parse(requestURL)
		if parseErr != nil {
			return "", fmt.Errorf("parsing ADO OIDC request URL: %w", parseErr)
		}
		query, _ := url.ParseQuery(parsedURL.RawQuery)
		if query.Get("api-version") == "" {
			query.Set("api-version", "7.1")
		}
		if query.Get("serviceConnectionId") == "" {
			query.Set("serviceConnectionId", serviceConnectionID)
		}
		if query.Get("audience") == "" {
			query.Set("audience", "api://AzureADTokenExchange")
		}
		parsedURL.RawQuery = query.Encode()

		req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, parsedURL.String(), nil)
		if reqErr != nil {
			return "", fmt.Errorf("building ADO OIDC token request: %w", reqErr)
		}
		req.Header.Set("Authorization", "Bearer "+bearerToken)
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, doErr := http.DefaultClient.Do(req)
		if doErr != nil {
			return "", fmt.Errorf("fetching ADO OIDC token: %w", doErr)
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		if readErr != nil {
			return "", fmt.Errorf("reading ADO OIDC token response: %w", readErr)
		}

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return "", fmt.Errorf("ADO OIDC token request failed (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var result struct {
			OIDCToken *string `json:"oidcToken"`
		}
		if decodeErr := json.Unmarshal(body, &result); decodeErr != nil {
			return "", fmt.Errorf("decoding ADO OIDC token response: %w", decodeErr)
		}
		if result.OIDCToken == nil || *result.OIDCToken == "" {
			return "", fmt.Errorf("ADO OIDC token response contained empty oidcToken field")
		}
		return *result.OIDCToken, nil
	}
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

// RequestOptions customizes a single Admin Center API request. A nil *RequestOptions
// selects the defaults: a JSON content type, no extra headers, and the client's
// configured HTTP timeout.
type RequestOptions struct {
	// ContentType overrides the default "application/json" request content type.
	ContentType string
	// Headers are additional request headers applied after the defaults.
	Headers map[string]string
	// Timeout overrides the client's HTTP timeout for this request only. Uploads of
	// large payloads (for example a 50 MB .app package) need more than the default.
	Timeout time.Duration
}

// DoRequest performs an authenticated HTTP request to the Business Central Admin Center API.
func (c *Client) DoRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	return c.DoRequestWithOptions(ctx, method, path, body, nil)
}

// DoRequestWithOptions performs an authenticated HTTP request to the Business Central
// Admin Center API, allowing the content type, extra headers, and HTTP timeout to be
// overridden for this request only.
func (c *Client) DoRequestWithOptions(ctx context.Context, method, path string, body io.Reader, opts *RequestOptions) (*http.Response, error) {
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
	contentType := "application/json"
	if opts != nil && opts.ContentType != "" {
		contentType = opts.ContentType
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")
	if opts != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	// Execute request, applying a per-request timeout when one is requested.
	httpClient := c.httpClient
	if opts != nil && opts.Timeout > 0 && opts.Timeout != httpClient.Timeout {
		clientCopy := *httpClient
		clientCopy.Timeout = opts.Timeout
		httpClient = &clientCopy
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	// Check for error responses.
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		return nil, newAdminCenterError(resp)
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

// PostMultipart performs an authenticated POST request carrying a multipart/form-data
// payload. contentType must be the boundary-carrying value produced by
// multipart.Writer.FormDataContentType(). timeout overrides the client's HTTP timeout
// when greater than zero, which large uploads require.
func (c *Client) PostMultipart(ctx context.Context, path string, body io.Reader, contentType string, timeout time.Duration) (*http.Response, error) {
	return c.DoRequestWithOptions(ctx, http.MethodPost, path, body, &RequestOptions{
		ContentType: contentType,
		Timeout:     timeout,
	})
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
