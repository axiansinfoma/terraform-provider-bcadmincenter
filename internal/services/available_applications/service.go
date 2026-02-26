// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package available_applications

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// Service handles available applications-related operations for the Business Central Admin Center API.
type Service struct {
	client *client.Client
}

// NewService creates a new available applications service.
func NewService(c *client.Client) *Service {
	return &Service{
		client: c,
	}
}

// GetAvailableApplications retrieves the list of available application families with their countries and rings.
func (s *Service) GetAvailableApplications(ctx context.Context) (*AvailableApplicationsResponse, error) {
	path := "applications/"

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get available applications: %w", err)
	}
	defer resp.Body.Close()

	var appList AvailableApplicationsResponse
	if err := json.NewDecoder(resp.Body).Decode(&appList); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &appList, nil
}

// GetApplicationFamily retrieves information about a specific application family by name.
func (s *Service) GetApplicationFamily(ctx context.Context, familyName string) (*ApplicationFamily, error) {
	// Get all available applications.
	availableApps, err := s.GetAvailableApplications(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available applications: %w", err)
	}

	// Find the requested application family.
	for _, appFamily := range availableApps.Value {
		if appFamily.ApplicationFamily == familyName {
			return &appFamily, nil
		}
	}

	return nil, fmt.Errorf("application family '%s' not found", familyName)
}
