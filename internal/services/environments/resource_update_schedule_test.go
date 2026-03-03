// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestUpdateScheduleResource_Metadata(t *testing.T) {
	r := NewUpdateScheduleResource()
	req := resource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_environment_update_schedule"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestUpdateScheduleResource_Schema(t *testing.T) {
	r := NewUpdateScheduleResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	requiredAttrs := []string{"application_family", "environment_name", "target_version"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing required attribute: %s", attr)
		}
	}

	// Verify optional attributes exist.
	optionalAttrs := []string{"aad_tenant_id", "scheduled_datetime", "ignore_update_window"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing optional attribute: %s", attr)
		}
	}

	// Verify computed attributes exist.
	computedAttrs := []string{"id", "update_status", "rollout_status", "latest_selectable_datetime"}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing computed attribute: %s", attr)
		}
	}
}

func TestUpdateScheduleResource_Configure(t *testing.T) {
	r := &UpdateScheduleResource{}

	// Test with nil provider data (should not error, just skip).
	req := resource.ConfigureRequest{
		ProviderData: nil,
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil data returned errors: %v", resp.Diagnostics)
	}

	if r.client != nil {
		t.Error("Configure() with nil data should not set client")
	}
}

func TestUpdateScheduleResource_Configure_WithInvalidType(t *testing.T) {
	r := &UpdateScheduleResource{}

	req := resource.ConfigureRequest{
		ProviderData: "invalid-type",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure() with invalid type should return error")
	}
}

func TestUpdateScheduleResource_ImportState_InvalidID(t *testing.T) {
	tests := []struct {
		name      string
		importID  string
		wantError bool
	}{
		{
			name:      "invalid import ID - too few parts",
			importID:  "BusinessCentral/test-sandbox",
			wantError: true,
		},
		{
			name:      "invalid import ID - missing updateSchedule suffix",
			importID:  "/tenants/tenant123/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/prod",
			wantError: true,
		},
		{
			name:      "invalid import ID - empty",
			importID:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &UpdateScheduleResource{}
			req := resource.ImportStateRequest{
				ID: tt.importID,
			}
			resp := &resource.ImportStateResponse{}

			r.ImportState(context.Background(), req, resp)

			hasError := resp.Diagnostics.HasError()
			if hasError != tt.wantError {
				t.Errorf("ImportState() hasError = %v, wantError %v, diagnostics: %v", hasError, tt.wantError, resp.Diagnostics)
			}
		})
	}
}

func TestUpdateScheduleResourceModel(t *testing.T) {
	// Test that the model struct can be created and populated.
	model := UpdateScheduleResourceModel{}

	// Verify the struct has all expected fields.
	_ = model.ID
	_ = model.AADTenantID
	_ = model.ApplicationFamily
	_ = model.EnvironmentName
	_ = model.TargetVersion
	_ = model.ScheduledDatetime
	_ = model.IgnoreUpdateWindow
	_ = model.UpdateStatus
	_ = model.RolloutStatus
	_ = model.LatestSelectableDatetime
}

func TestNewUpdateScheduleResource(t *testing.T) {
	r := NewUpdateScheduleResource()

	if r == nil {
		t.Error("NewUpdateScheduleResource() returned nil")
	}
}

func TestFindSelectedUpdate(t *testing.T) {
	tests := []struct {
		name        string
		updates     []EnvironmentUpdate
		wantVersion string
		wantNil     bool
	}{
		{
			name: "finds selected update",
			updates: []EnvironmentUpdate{
				{TargetVersion: "26.0", Selected: false},
				{TargetVersion: "26.1", Selected: true, UpdateStatus: UpdateStatusScheduled},
			},
			wantVersion: "26.1",
			wantNil:     false,
		},
		{
			name: "no selected update",
			updates: []EnvironmentUpdate{
				{TargetVersion: "26.0", Selected: false},
				{TargetVersion: "26.1", Selected: false},
			},
			wantNil: true,
		},
		{
			name:    "empty list",
			updates: []EnvironmentUpdate{},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findSelectedUpdate(tt.updates)
			if tt.wantNil {
				if result != nil {
					t.Errorf("findSelectedUpdate() expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Error("findSelectedUpdate() expected non-nil, got nil")
					return
				}
				if result.TargetVersion != tt.wantVersion {
					t.Errorf("findSelectedUpdate() targetVersion = %v, want %v", result.TargetVersion, tt.wantVersion)
				}
			}
		})
	}
}

func TestApplyUpdatesDriftDetection(t *testing.T) {
	tests := []struct {
		name               string
		envVersion         string
		updates            []EnvironmentUpdate
		expectedVersion    string
	}{
		{
			name:       "scheduled update - suppress drift with target version",
			envVersion: "25.0",
			updates: []EnvironmentUpdate{
				{TargetVersion: "26.1", Selected: true, UpdateStatus: UpdateStatusScheduled},
			},
			expectedVersion: "26.1",
		},
		{
			name:       "running update - suppress drift with target version",
			envVersion: "25.0",
			updates: []EnvironmentUpdate{
				{TargetVersion: "26.1", Selected: true, UpdateStatus: UpdateStatusRunning},
			},
			expectedVersion: "26.1",
		},
		{
			name:       "failed update - drift with current running version",
			envVersion: "25.0",
			updates: []EnvironmentUpdate{
				{TargetVersion: "26.1", Selected: true, UpdateStatus: UpdateStatusFailed},
			},
			expectedVersion: "25.0",
		},
		{
			name:       "no selected update - use env version",
			envVersion: "25.0",
			updates: []EnvironmentUpdate{
				{TargetVersion: "26.0", Selected: false},
			},
			expectedVersion: "25.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &EnvironmentResource{}
			model := &EnvironmentResourceModel{}
			env := &Environment{
				Name:               "production",
				ApplicationFamily:  "BusinessCentral",
				ApplicationVersion: tt.envVersion,
				AADTenantID:        "tenant-id",
			}

			r.updateModelFromEnvironment(model, env)
			r.applyUpdatesDriftDetection(model, env, tt.updates)

			if model.ApplicationVersion.ValueString() != tt.expectedVersion {
				t.Errorf("application_version = %v, want %v", model.ApplicationVersion.ValueString(), tt.expectedVersion)
			}
		})
	}
}
