// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package quotas

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/vllni/terraform-provider-bcadmincenter/internal/client"
)

// Service handles quotas operations
type Service struct {
	client *client.Client
}

// NewService creates a new quotas service
func NewService(c *client.Client) *Service {
	return &Service{client: c}
}

// GetQuotas retrieves the environment quotas for the tenant
func (s *Service) GetQuotas(ctx context.Context) (*QuotasResponse, error) {
	path := "environments/quotas"

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotas: %w", err)
	}
	defer resp.Body.Close()

	var result QuotasResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}
