// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environmentsettings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/utils"
)

// Service handles environment settings operations for the Business Central Admin Center API.
type Service struct {
	client *client.Client
}

// NewService creates a new environment settings service.
func NewService(c *client.Client) *Service {
	return &Service{
		client: c,
	}
}

// GetUpdateSettings retrieves the update window settings for an environment.
func (s *Service) GetUpdateSettings(ctx context.Context, applicationFamily, environmentName string) (*UpdateSettings, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/upgrade", applicationFamily, environmentName)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get update settings: %w", err)
	}
	defer resp.Body.Close()

	// Return nil if no settings exist (null response)
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	var settings UpdateSettings
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &settings, nil
}

// SetUpdateSettings configures the update window for an environment.
func (s *Service) SetUpdateSettings(ctx context.Context, applicationFamily, environmentName string, settings *UpdateSettings) (*UpdateSettings, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/upgrade", applicationFamily, environmentName)

	body, err := json.Marshal(settings)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Put(ctx, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to set update settings: %w", err)
	}
	defer resp.Body.Close()

	var updatedSettings UpdateSettings
	if err := json.NewDecoder(resp.Body).Decode(&updatedSettings); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &updatedSettings, nil
}

// GetTimeZones retrieves the list of available time zones for update settings.
func (s *Service) GetTimeZones(ctx context.Context) ([]TimeZone, error) {
	path := "applications/settings/timezones"

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get time zones: %w", err)
	}
	defer resp.Body.Close()

	var tzList TimeZoneListResponse
	if err := json.NewDecoder(resp.Body).Decode(&tzList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tzList.Value, nil
}

// SetAppInsightsKey sets the Application Insights connection string for an environment.
// Note: This triggers an automatic environment restart.
func (s *Service) SetAppInsightsKey(ctx context.Context, applicationFamily, environmentName, key string) error {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/appinsightskey", applicationFamily, environmentName)

	req := AppInsightsKeyRequest{Key: key}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Post(ctx, path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to set Application Insights key: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	return nil
}

// GetSecurityGroup retrieves the Microsoft Entra security group assigned to an environment.
func (s *Service) GetSecurityGroup(ctx context.Context, applicationFamily, environmentName string) (*SecurityGroupResponse, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/securitygroupaccess", applicationFamily, environmentName)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get security group: %w", err)
	}
	defer resp.Body.Close()

	// 204 means no group is configured.
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	var group SecurityGroupResponse
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &group, nil
}

// SetSecurityGroup assigns a Microsoft Entra security group to an environment.
func (s *Service) SetSecurityGroup(ctx context.Context, applicationFamily, environmentName, groupID string) error {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/securitygroupaccess", applicationFamily, environmentName)

	req := SecurityGroupRequest{Value: groupID}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Post(ctx, path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to set security group: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	return nil
}

// ClearSecurityGroup removes the Microsoft Entra security group from an environment.
func (s *Service) ClearSecurityGroup(ctx context.Context, applicationFamily, environmentName string) error {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/securitygroupaccess", applicationFamily, environmentName)

	resp, err := s.client.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to clear security group: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	return nil
}

// GetAccessWithM365Licenses retrieves whether M365 license access is enabled.
// Returns nil if the setting is not available (404) or not configured.
func (s *Service) GetAccessWithM365Licenses(ctx context.Context, applicationFamily, environmentName string) (*AccessWithM365LicensesResponse, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/accesswithm365licenses", applicationFamily, environmentName)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		// Check if it's a 404 - feature not available on this environment.
		if apiErr, ok := err.(*client.AdminCenterError); ok && apiErr.Code == "ResourceNotFound" {
			return nil, nil // Return nil, nil to indicate feature not available
		}
		return nil, fmt.Errorf("failed to get M365 license access setting: %w", err)
	}
	defer resp.Body.Close()

	// Handle 204 No Content - setting not configured.
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	var accessSetting AccessWithM365LicensesResponse
	if err := json.NewDecoder(resp.Body).Decode(&accessSetting); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &accessSetting, nil
}

// SetAccessWithM365Licenses enables or disables M365 license access.
func (s *Service) SetAccessWithM365Licenses(ctx context.Context, applicationFamily, environmentName string, enabled bool) error {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/accesswithm365licenses", applicationFamily, environmentName)

	req := AccessWithM365LicensesRequest{Enabled: enabled}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Post(ctx, path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to set M365 license access: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	return nil
}

// SetAppUpdateCadence configures how frequently AppSource apps are updated.
func (s *Service) SetAppUpdateCadence(ctx context.Context, applicationFamily, environmentName, cadence string) error {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/appSourceAppsUpdateCadence", applicationFamily, environmentName)

	req := AppUpdateCadenceRequest{Value: cadence}
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Put(ctx, path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to set app update cadence: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	return nil
}

// GetPartnerAccess retrieves partner access settings for an environment.
func (s *Service) GetPartnerAccess(ctx context.Context, applicationFamily, environmentName string) (*PartnerAccessResponse, error) {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/partneraccess", applicationFamily, environmentName)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get partner access settings: %w", err)
	}
	defer resp.Body.Close()

	var accessSettings PartnerAccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&accessSettings); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &accessSettings, nil
}

// SetPartnerAccess configures partner access settings for an environment.
func (s *Service) SetPartnerAccess(ctx context.Context, applicationFamily, environmentName string, settings *PartnerAccessRequest) error {
	path := fmt.Sprintf("applications/%s/environments/%s/settings/partneraccess", applicationFamily, environmentName)

	body, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Put(ctx, path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to set partner access settings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, utils.ReadResponseBody(resp.Body))
	}

	return nil
}
