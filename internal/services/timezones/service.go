// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package timezones

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// Service handles timezone operations.
type Service struct {
	client *client.Client
}

// NewService creates a new timezones service.
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// GetTimeZones retrieves the list of available time zones.
func (s *Service) GetTimeZones(ctx context.Context) (*TimeZoneResponse, error) {
	path := "applications/settings/timezones"

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get timezones: %w", err)
	}
	defer resp.Body.Close()

	var result TimeZoneResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
