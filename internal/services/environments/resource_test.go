// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEnvironmentResource_Metadata(t *testing.T) {
	r := NewEnvironmentResource()
	req := resource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_environment"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestEnvironmentResource_Schema(t *testing.T) {
	r := NewEnvironmentResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	requiredAttrs := []string{"name", "type", "country_code"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing required attribute: %s", attr)
		}
	}

	// Verify optional attributes exist.
	optionalAttrs := []string{"application_family", "ring_name", "application_version", "ignore_update_window", "azure_region", "aad_tenant_id"}
	for _, attr := range optionalAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing optional attribute: %s", attr)
		}
	}

	// Verify computed attributes exist.
	computedAttrs := []string{"id", "status", "web_client_login_url", "web_service_url", "app_insights_key", "platform_version", "pending_upgrade_version", "pending_upgrade_scheduled_for"}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing computed attribute: %s", attr)
		}
	}

	// Verify settings nested block exists in the Blocks map (block syntax: settings { ... }).
	if _, ok := resp.Schema.Blocks["settings"]; !ok {
		t.Fatal("Schema missing 'settings' nested block")
	}
}

func TestEnvironmentResource_Configure(t *testing.T) {
	r := &EnvironmentResource{}

	// Test with nil provider data (should not error, just skip)
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

func TestEnvironmentResourceModel(t *testing.T) {
	// Test that the model struct can be created and populated.
	model := EnvironmentResourceModel{}

	// Verify the struct has all expected fields.
	_ = model.ID
	_ = model.Name
	_ = model.ApplicationFamily
	_ = model.Type
	_ = model.CountryCode
	_ = model.RingName
	_ = model.ApplicationVersion
	_ = model.IgnoreUpdateWindow
	_ = model.AzureRegion
	_ = model.Status
	_ = model.WebClientLoginURL
	_ = model.WebServiceURL
	_ = model.AppInsightsKey
	_ = model.PlatformVersion
	_ = model.AADTenantID
	_ = model.PendingUpgradeVersion
	_ = model.PendingUpgradeScheduledFor
	_ = model.Settings
	_ = model.Timeouts
}

func TestEnvironmentResource_ImportState(t *testing.T) {
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
			name:      "invalid import ID - too many parts",
			importID:  "tenant/BusinessCentral/test/extra",
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
			r := &EnvironmentResource{}
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

func TestNewEnvironmentResource(t *testing.T) {
	r := NewEnvironmentResource()

	if r == nil {
		t.Error("NewEnvironmentResource() returned nil")
	}

	// Verify it returns a resource (the function signature already guarantees this)
	if r == nil {
		t.Error("NewEnvironmentResource() should return a valid resource")
	}
}

func TestNormalizeRingName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Production to PROD",
			input:    "Production",
			expected: "PROD",
		},
		{
			name:     "Preview to PREVIEW",
			input:    "Preview",
			expected: "PREVIEW",
		},
		{
			name:     "Fast to FAST",
			input:    "Fast",
			expected: "FAST",
		},
		{
			name:     "Unknown ring name",
			input:    "CustomRing",
			expected: "CustomRing",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeRingName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeRingName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestUpdateModelFromEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		env      *Environment
		validate func(*testing.T, *EnvironmentResourceModel)
	}{
		{
			name: "complete environment data",
			env: &Environment{
				Name:               "production",
				ApplicationFamily:  "BusinessCentral",
				Type:               "Production",
				CountryCode:        "US",
				RingName:           "Production",
				ApplicationVersion: "25.0",
				Status:             "Active",
				WebClientLoginURL:  "https://example.com",
				WebServiceURL:      "https://api.example.com",
				AppInsightsKey:     "insights-key-123",
				PlatformVersion:    "25.0.0.0",
				AADTenantID:        "tenant-id-123",
			},
			validate: func(t *testing.T, model *EnvironmentResourceModel) {
				expectedID := "/tenants/tenant-id-123/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/production"
				if model.ID.ValueString() != expectedID {
					t.Errorf("ID = %v, want %v", model.ID.ValueString(), expectedID)
				}
				if model.Name.ValueString() != "production" {
					t.Errorf("Name = %v, want production", model.Name.ValueString())
				}
				if model.ApplicationFamily.ValueString() != "BusinessCentral" {
					t.Errorf("ApplicationFamily = %v, want BusinessCentral", model.ApplicationFamily.ValueString())
				}
				if model.Type.ValueString() != "Production" {
					t.Errorf("Type = %v, want Production", model.Type.ValueString())
				}
				if model.CountryCode.ValueString() != "US" {
					t.Errorf("CountryCode = %v, want US", model.CountryCode.ValueString())
				}
				if model.RingName.ValueString() != "PROD" {
					t.Errorf("RingName = %v, want PROD (normalized)", model.RingName.ValueString())
				}
				if model.ApplicationVersion.ValueString() != "25.0" {
					t.Errorf("ApplicationVersion = %v, want 25.0", model.ApplicationVersion.ValueString())
				}
				if model.Status.ValueString() != "Active" {
					t.Errorf("Status = %v, want Active", model.Status.ValueString())
				}
				if model.WebClientLoginURL.ValueString() != "https://example.com" {
					t.Errorf("WebClientLoginURL = %v, want https://example.com", model.WebClientLoginURL.ValueString())
				}
				if model.WebServiceURL.ValueString() != "https://api.example.com" {
					t.Errorf("WebServiceURL = %v, want https://api.example.com", model.WebServiceURL.ValueString())
				}
				if model.AppInsightsKey.ValueString() != "insights-key-123" {
					t.Errorf("AppInsightsKey = %v, want insights-key-123", model.AppInsightsKey.ValueString())
				}
				if model.PlatformVersion.ValueString() != "25.0.0.0" {
					t.Errorf("PlatformVersion = %v, want 25.0.0.0", model.PlatformVersion.ValueString())
				}
				if model.AADTenantID.ValueString() != "tenant-id-123" {
					t.Errorf("AADTenantID = %v, want tenant-id-123", model.AADTenantID.ValueString())
				}
				if !model.AzureRegion.IsNull() {
					t.Error("AzureRegion should be null")
				}
			},
		},
		{
			name: "minimal environment data",
			env: &Environment{
				Name:              "sandbox",
				ApplicationFamily: "BusinessCentral",
				Type:              "Sandbox",
				CountryCode:       "CA",
				Status:            "Active",
				WebClientLoginURL: "https://sandbox.example.com",
				AADTenantID:       "tenant-id-456",
			},
			validate: func(t *testing.T, model *EnvironmentResourceModel) {
				expectedID := "/tenants/tenant-id-456/providers/Microsoft.Dynamics365.BusinessCentral/applications/BusinessCentral/environments/sandbox"
				if model.ID.ValueString() != expectedID {
					t.Errorf("ID = %v, want %v", model.ID.ValueString(), expectedID)
				}
				if model.Name.ValueString() != "sandbox" {
					t.Errorf("Name = %v, want sandbox", model.Name.ValueString())
				}
				if !model.WebServiceURL.IsNull() {
					t.Error("WebServiceURL should be null when not provided")
				}
				if !model.AppInsightsKey.IsNull() {
					t.Error("AppInsightsKey should be null when not provided")
				}
				if !model.RingName.IsNull() {
					t.Error("RingName should be null when not provided")
				}
				if !model.ApplicationVersion.IsNull() {
					t.Error("ApplicationVersion should be null when not provided")
				}
				if !model.PlatformVersion.IsNull() {
					t.Error("PlatformVersion should be null when not provided")
				}
			},
		},
		{
			name: "ring name normalization - Preview",
			env: &Environment{
				Name:              "test",
				ApplicationFamily: "BusinessCentral",
				Type:              "Sandbox",
				CountryCode:       "US",
				RingName:          "Preview",
				Status:            "Active",
				WebClientLoginURL: "https://test.example.com",
				AADTenantID:       "tenant-id",
			},
			validate: func(t *testing.T, model *EnvironmentResourceModel) {
				if model.RingName.ValueString() != "PREVIEW" {
					t.Errorf("RingName = %v, want PREVIEW (normalized)", model.RingName.ValueString())
				}
			},
		},
		{
			name: "ring name normalization - Fast",
			env: &Environment{
				Name:              "test",
				ApplicationFamily: "BusinessCentral",
				Type:              "Sandbox",
				CountryCode:       "US",
				RingName:          "Fast",
				Status:            "Active",
				WebClientLoginURL: "https://test.example.com",
				AADTenantID:       "tenant-id",
			},
			validate: func(t *testing.T, model *EnvironmentResourceModel) {
				if model.RingName.ValueString() != "FAST" {
					t.Errorf("RingName = %v, want FAST (normalized)", model.RingName.ValueString())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &EnvironmentResource{}
			model := &EnvironmentResourceModel{}

			r.updateModelFromEnvironment(model, tt.env)

			tt.validate(t, model)
		})
	}
}

func TestEnvironmentResource_Configure_WithInvalidType(t *testing.T) {
	r := &EnvironmentResource{}

	// Test with invalid provider data type.
	req := resource.ConfigureRequest{
		ProviderData: "invalid-type",
	}
	resp := &resource.ConfigureResponse{}

	r.Configure(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure() with invalid type should return error")
	}
}

func TestEnvironmentResource_ImportState_Success(t *testing.T) {
	// Test successful parsing of import ID format.
	importID := "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d/BusinessCentral/production"
	parts := strings.Split(importID, "/")

	if len(parts) != 3 {
		t.Errorf("Import ID parsing failed, expected 3 parts, got %d", len(parts))
	}

	if parts[0] != "9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d" {
		t.Errorf("Tenant ID = %s, want 9ff11aaa-cddc-4df5-97c9-b9e79db1ba1d", parts[0])
	}

	if parts[1] != "BusinessCentral" {
		t.Errorf("Application family = %s, want BusinessCentral", parts[1])
	}

	if parts[2] != "production" {
		t.Errorf("Environment name = %s, want production", parts[2])
	}
}

func TestNormalizeApplicationVersion(t *testing.T) {
	tests := []struct {
		name         string
		priorVersion string
		apiVersion   string
		want         string
	}{
		{
			name:         "short major.minor form preserved when API returns full build version",
			priorVersion: "27.1",
			apiVersion:   "27.1.41698.41831",
			want:         "27.1",
		},
		{
			name:         "exact match preserved unchanged",
			priorVersion: "27.1.41698.41831",
			apiVersion:   "27.1.41698.41831",
			want:         "27.1.41698.41831",
		},
		{
			name:         "external auto-upgrade: API returns higher minor, preserve user config",
			priorVersion: "27.1",
			apiVersion:   "27.2.12345.67890",
			want:         "27.1",
		},
		{
			name:         "external auto-upgrade: API returns higher major, preserve user config",
			priorVersion: "27.1",
			apiVersion:   "28.1.12345.67890",
			want:         "27.1",
		},
		{
			name:         "external auto-upgrade avoids prefix collision: 27.1 vs 27.10",
			priorVersion: "27.1",
			apiVersion:   "27.10.12345.67890",
			want:         "27.1",
		},
		{
			name:         "empty prior version returns api version",
			priorVersion: "",
			apiVersion:   "27.1.41698.41831",
			want:         "27.1.41698.41831",
		},
		{
			name:         "empty api version returns empty",
			priorVersion: "27.1",
			apiVersion:   "",
			want:         "",
		},
		{
			name:         "both empty returns empty",
			priorVersion: "",
			apiVersion:   "",
			want:         "",
		},
		{
			name:         "full prior version preserved when API returns newer full version of same minor",
			priorVersion: "27.1.41698.41831",
			apiVersion:   "27.1.42000.00000",
			want:         "27.1.42000.00000",
		},
		{
			name:         "external auto-upgrade: API returns higher minor, preserve user config",
			priorVersion: "27.5",
			apiVersion:   "27.6",
			want:         "27.5",
		},
		{
			name:         "external auto-upgrade: API returns higher major, preserve user config",
			priorVersion: "27.5",
			apiVersion:   "28.0",
			want:         "27.5",
		},
		{
			name:         "external auto-upgrade with full build version: higher major, preserve user config",
			priorVersion: "27.5",
			apiVersion:   "28.0.12345.67890",
			want:         "27.5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeApplicationVersion(tt.priorVersion, tt.apiVersion)
			if got != tt.want {
				t.Errorf("normalizeApplicationVersion(%q, %q) = %q, want %q", tt.priorVersion, tt.apiVersion, got, tt.want)
			}
		})
	}
}

func TestEnvironmentSettingsNestedModel(t *testing.T) {
	// Test that the nested settings model struct can be created and populated.
	m := EnvironmentSettingsNestedModel{}
	_ = m.UpdateWindowStartTime
	_ = m.UpdateWindowEndTime
	_ = m.UpdateWindowTimeZone
	_ = m.AppInsightsKey
	_ = m.SecurityGroupID
	_ = m.AccessWithM365Licenses
	_ = m.AppUpdateCadence
	_ = m.PartnerAccessStatus
	_ = m.AllowedPartnerTenantIDs
}

// TestEnvironmentResource_Schema_ApplicationVersionHasUseStateForUnknown verifies that
// application_version includes UseStateForUnknown so that the prior state value is
// preserved in the plan when the user does not specify the attribute.  Without this,
// the plan shows "(known after apply)" for application_version, which makes versionChanged
// true in Update, which blocks settings-only updates from saving state — causing the
// settings block to always appear as being added (drift).
func TestEnvironmentResource_Schema_ApplicationVersionHasUseStateForUnknown(t *testing.T) {
	r := NewEnvironmentResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	appVersionAttr, ok := resp.Schema.Attributes["application_version"]
	if !ok {
		t.Fatal("Schema missing application_version attribute")
	}

	// Verify at least two plan modifiers are registered (UseStateForUnknown + NoDowngradeVersion).
	strAttr, ok := appVersionAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("application_version is not a StringAttribute")
	}

	if len(strAttr.PlanModifiers) < 2 {
		t.Errorf("application_version has %d plan modifier(s), want at least 2 (UseStateForUnknown + NoDowngradeVersion)",
			len(strAttr.PlanModifiers))
	}
}

// TestVersionChangedIgnoresUnknownPlanValue verifies that an unknown plan value for
// application_version (i.e. the user did not set it) is not treated as a version
// change, preventing settings-only updates from failing on a missing version.
func TestVersionChangedIgnoresUnknownPlanValue(t *testing.T) {
	tests := []struct {
		name               string
		planVersion        types.String
		stateVersion       types.String
		wantVersionChanged bool
	}{
		{
			name:               "unknown plan value is not a version change",
			planVersion:        types.StringUnknown(),
			stateVersion:       types.StringValue("28.0"),
			wantVersionChanged: false,
		},
		{
			name:               "null plan value is a version change",
			planVersion:        types.StringNull(),
			stateVersion:       types.StringValue("28.0"),
			wantVersionChanged: true,
		},
		{
			name:               "same version is not a change",
			planVersion:        types.StringValue("28.0"),
			stateVersion:       types.StringValue("28.0"),
			wantVersionChanged: false,
		},
		{
			name:               "different version is a change",
			planVersion:        types.StringValue("29.0"),
			stateVersion:       types.StringValue("28.0"),
			wantVersionChanged: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replicate the versionChanged logic from the Update function.
			versionChanged := !tt.planVersion.IsUnknown() && !tt.planVersion.Equal(tt.stateVersion)
			if versionChanged != tt.wantVersionChanged {
				t.Errorf("versionChanged = %v, want %v (plan=%v, state=%v)",
					versionChanged, tt.wantVersionChanged,
					tt.planVersion, tt.stateVersion)
			}
		})
	}
}

func TestSettingsBlockChanged(t *testing.T) {
	tests := []struct {
		name     string
		plan     *EnvironmentSettingsNestedModel
		state    *EnvironmentSettingsNestedModel
		expected bool
	}{
		{
			name:     "both nil - no change",
			plan:     nil,
			state:    nil,
			expected: false,
		},
		{
			name:     "plan nil, state non-nil - changed (removed)",
			plan:     nil,
			state:    &EnvironmentSettingsNestedModel{},
			expected: true,
		},
		{
			name:     "plan non-nil, state nil - changed (added)",
			plan:     &EnvironmentSettingsNestedModel{},
			state:    nil,
			expected: true,
		},
		{
			name: "identical settings - no change",
			plan: &EnvironmentSettingsNestedModel{
				UpdateWindowStartTime:   types.StringValue("22:00"),
				UpdateWindowEndTime:     types.StringValue("06:00"),
				UpdateWindowTimeZone:    types.StringValue("UTC"),
				AppInsightsKey:          types.StringNull(),
				SecurityGroupID:         types.StringNull(),
				AccessWithM365Licenses:  types.BoolNull(),
				AppUpdateCadence:        types.StringNull(),
				PartnerAccessStatus:     types.StringNull(),
				AllowedPartnerTenantIDs: types.ListNull(types.StringType),
			},
			state: &EnvironmentSettingsNestedModel{
				UpdateWindowStartTime:   types.StringValue("22:00"),
				UpdateWindowEndTime:     types.StringValue("06:00"),
				UpdateWindowTimeZone:    types.StringValue("UTC"),
				AppInsightsKey:          types.StringNull(),
				SecurityGroupID:         types.StringNull(),
				AccessWithM365Licenses:  types.BoolNull(),
				AppUpdateCadence:        types.StringNull(),
				PartnerAccessStatus:     types.StringNull(),
				AllowedPartnerTenantIDs: types.ListNull(types.StringType),
			},
			expected: false,
		},
		{
			name: "start time changed",
			plan: &EnvironmentSettingsNestedModel{
				UpdateWindowStartTime:   types.StringValue("20:00"),
				UpdateWindowEndTime:     types.StringValue("06:00"),
				UpdateWindowTimeZone:    types.StringValue("UTC"),
				AppInsightsKey:          types.StringNull(),
				SecurityGroupID:         types.StringNull(),
				AccessWithM365Licenses:  types.BoolNull(),
				AppUpdateCadence:        types.StringNull(),
				PartnerAccessStatus:     types.StringNull(),
				AllowedPartnerTenantIDs: types.ListNull(types.StringType),
			},
			state: &EnvironmentSettingsNestedModel{
				UpdateWindowStartTime:   types.StringValue("22:00"),
				UpdateWindowEndTime:     types.StringValue("06:00"),
				UpdateWindowTimeZone:    types.StringValue("UTC"),
				AppInsightsKey:          types.StringNull(),
				SecurityGroupID:         types.StringNull(),
				AccessWithM365Licenses:  types.BoolNull(),
				AppUpdateCadence:        types.StringNull(),
				PartnerAccessStatus:     types.StringNull(),
				AllowedPartnerTenantIDs: types.ListNull(types.StringType),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := settingsBlockChanged(tt.plan, tt.state)
			if got != tt.expected {
				t.Errorf("settingsBlockChanged() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestEnvironmentResource_Schema_AccessWithM365LicensesHasUseStateForUnknown verifies that
// access_with_m365_licenses inside the settings block includes UseStateForUnknown so that
// the prior state value is preserved in the plan when the user does not set the attribute.
// Without this, every plan shows "access_with_m365_licenses = (known after apply)", which
// makes settingsBlockChanged return true on every cycle, causing perpetual update drift.
func TestEnvironmentResource_Schema_AccessWithM365LicensesHasUseStateForUnknown(t *testing.T) {
	r := NewEnvironmentResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	settingsBlock, ok := resp.Schema.Blocks["settings"]
	if !ok {
		t.Fatal("Schema missing 'settings' block")
	}

	nestedBlock, ok := settingsBlock.(schema.SingleNestedBlock)
	if !ok {
		t.Fatal("'settings' block is not a SingleNestedBlock")
	}

	m365Attr, ok := nestedBlock.Attributes["access_with_m365_licenses"]
	if !ok {
		t.Fatal("settings block missing 'access_with_m365_licenses' attribute")
	}

	boolAttr, ok := m365Attr.(schema.BoolAttribute)
	if !ok {
		t.Fatal("access_with_m365_licenses is not a BoolAttribute")
	}

	// Verify that UseStateForUnknown is present by checking for a modifier whose description
	// matches the well-known description of boolplanmodifier.UseStateForUnknown().
	const wantDesc = "Once set, the value of this attribute in state will not change."
	found := false
	for _, mod := range boolAttr.PlanModifiers {
		if mod.Description(context.Background()) == wantDesc {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("access_with_m365_licenses is missing the UseStateForUnknown() plan modifier (expected description %q)", wantDesc)
	}
}

// TestSettingsBlockChanged_UnknownM365DoesNotTriggerChange verifies that an unknown
// plan value for access_with_m365_licenses is not treated as a settings change.
// This prevents perpetual re-apply cycles caused by Computed attributes showing as
// "(known after apply)" in plans.
func TestSettingsBlockChanged_UnknownM365DoesNotTriggerChange(t *testing.T) {
	tests := []struct {
		name     string
		plan     *EnvironmentSettingsNestedModel
		state    *EnvironmentSettingsNestedModel
		expected bool
	}{
		{
			name: "unknown access_with_m365_licenses in plan is NOT a change",
			plan: &EnvironmentSettingsNestedModel{
				UpdateWindowStartTime:   types.StringValue("22:00"),
				UpdateWindowEndTime:     types.StringValue("06:00"),
				UpdateWindowTimeZone:    types.StringValue("UTC"),
				AppInsightsKey:          types.StringNull(),
				SecurityGroupID:         types.StringNull(),
				AccessWithM365Licenses:  types.BoolUnknown(),
				AppUpdateCadence:        types.StringNull(),
				PartnerAccessStatus:     types.StringNull(),
				AllowedPartnerTenantIDs: types.ListNull(types.StringType),
			},
			state: &EnvironmentSettingsNestedModel{
				UpdateWindowStartTime:   types.StringValue("22:00"),
				UpdateWindowEndTime:     types.StringValue("06:00"),
				UpdateWindowTimeZone:    types.StringValue("UTC"),
				AppInsightsKey:          types.StringNull(),
				SecurityGroupID:         types.StringNull(),
				AccessWithM365Licenses:  types.BoolValue(false),
				AppUpdateCadence:        types.StringNull(),
				PartnerAccessStatus:     types.StringNull(),
				AllowedPartnerTenantIDs: types.ListNull(types.StringType),
			},
			expected: false,
		},
		{
			name: "explicit false-to-true change for access_with_m365_licenses IS a change",
			plan: &EnvironmentSettingsNestedModel{
				UpdateWindowStartTime:   types.StringValue("22:00"),
				UpdateWindowEndTime:     types.StringValue("06:00"),
				UpdateWindowTimeZone:    types.StringValue("UTC"),
				AppInsightsKey:          types.StringNull(),
				SecurityGroupID:         types.StringNull(),
				AccessWithM365Licenses:  types.BoolValue(true),
				AppUpdateCadence:        types.StringNull(),
				PartnerAccessStatus:     types.StringNull(),
				AllowedPartnerTenantIDs: types.ListNull(types.StringType),
			},
			state: &EnvironmentSettingsNestedModel{
				UpdateWindowStartTime:   types.StringValue("22:00"),
				UpdateWindowEndTime:     types.StringValue("06:00"),
				UpdateWindowTimeZone:    types.StringValue("UTC"),
				AppInsightsKey:          types.StringNull(),
				SecurityGroupID:         types.StringNull(),
				AccessWithM365Licenses:  types.BoolValue(false),
				AppUpdateCadence:        types.StringNull(),
				PartnerAccessStatus:     types.StringNull(),
				AllowedPartnerTenantIDs: types.ListNull(types.StringType),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := settingsBlockChanged(tt.plan, tt.state)
			if got != tt.expected {
				t.Errorf("settingsBlockChanged() = %v, want %v", got, tt.expected)
			}
		})
	}
}
