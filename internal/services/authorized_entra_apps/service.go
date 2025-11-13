// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package authorized_entra_apps

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// Service handles authorized Microsoft Entra apps operations.
type Service struct {
	client *client.Client
}

// NewService creates a new instance of the authorized apps service.
func NewService(c *client.Client) *Service {
	return &Service{
		client: c,
	}
}

// ListAuthorizedApps returns all authorized Microsoft Entra apps for the tenant.
// Note: This endpoint cannot be used when authenticated as an app.
func (s *Service) ListAuthorizedApps(ctx context.Context) ([]AuthorizedApp, error) {
	path := "authorizedAadApps"

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list authorized apps: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var apps AuthorizedAppsResponse
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return apps, nil
}

// AuthorizeApp authorizes a Microsoft Entra app to call the Business Central Admin Center API.
// Note: This endpoint cannot be used when authenticated as an app.
// This does not grant admin consent or assign permission sets in environments.
func (s *Service) AuthorizeApp(ctx context.Context, appID string) (*AuthorizedApp, error) {
	path := fmt.Sprintf("authorizedAadApps/%s", appID)

	resp, err := s.client.Put(ctx, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to authorize app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var app AuthorizedApp
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &app, nil
}

// RemoveAuthorizedApp removes a Microsoft Entra app from the authorized apps list.
// This does not revoke admin consent in Microsoft Entra ID nor remove permission sets in environments.
func (s *Service) RemoveAuthorizedApp(ctx context.Context, appID string) error {
	path := fmt.Sprintf("authorizedAadApps/%s", appID)

	resp, err := s.client.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to remove authorized app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetManageableTenants returns a list of tenants where the app is authorized.
// Note: This endpoint can only be used when authenticated as an app.
func (s *Service) GetManageableTenants(ctx context.Context) ([]ManageableTenant, error) {
	path := "authorizedAadApps/manageableTenants"

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get manageable tenants: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var response ManageableTenantsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.Value, nil
}
