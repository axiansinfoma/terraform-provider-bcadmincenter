// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package notificationrecipients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
)

// Service handles notification recipient operations for the Business Central Admin Center API.
type Service struct {
	client *client.Client
}

// NewService creates a new notification recipients service.
func NewService(c *client.Client) *Service {
	return &Service{
		client: c,
	}
}

// List retrieves all notification recipients.
// The tenant is determined by the client's OAuth token (see client.ForTenant).
func (s *Service) List(ctx context.Context) ([]NotificationRecipient, error) {
	path := "settings/notification/recipients"

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification recipients: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var recipientsResp NotificationRecipientsResponse
	if err := json.NewDecoder(resp.Body).Decode(&recipientsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return recipientsResp.Value, nil
}

// Get retrieves a specific notification recipient by ID.
// The tenant is determined by the client's OAuth token (see client.ForTenant).
func (s *Service) Get(ctx context.Context, id string) (*NotificationRecipient, error) {
	recipients, err := s.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, recipient := range recipients {
		if recipient.ID == id {
			return &recipient, nil
		}
	}

	return nil, fmt.Errorf("notification recipient with ID %s not found", id)
}

// Create creates a new notification recipient.
// The tenant is determined by the client's OAuth token (see client.ForTenant).
func (s *Service) Create(ctx context.Context, email, name string) (*NotificationRecipient, error) {
	path := "settings/notification/recipients"

	req := CreateNotificationRecipientRequest{
		Email: email,
		Name:  name,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := s.client.Put(ctx, path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create notification recipient: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var recipient NotificationRecipient
	if err := json.NewDecoder(resp.Body).Decode(&recipient); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &recipient, nil
}

// Delete deletes a notification recipient by ID.
// The tenant is determined by the client's OAuth token (see client.ForTenant).
func (s *Service) Delete(ctx context.Context, id string) error {
	path := fmt.Sprintf("settings/notification/recipients/%s", id)

	resp, err := s.client.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete notification recipient: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

// GetNotificationSettings retrieves the complete notification settings including all recipients.
// The tenant is determined by the client's OAuth token (see client.ForTenant).
func (s *Service) GetNotificationSettings(ctx context.Context) (*NotificationSettings, error) {
	path := "settings/notification"

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification settings: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var settings NotificationSettings
	if err := json.NewDecoder(resp.Body).Decode(&settings); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &settings, nil
}
