// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package notificationrecipients

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestNotificationRecipientResource_Metadata(t *testing.T) {
	r := NewNotificationRecipientResource()
	req := resource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &resource.MetadataResponse{}

	r.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_notification_recipient"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestNotificationRecipientResource_Schema(t *testing.T) {
	r := NewNotificationRecipientResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}

	r.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify required attributes exist.
	requiredAttrs := []string{"email", "name"}
	for _, attr := range requiredAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing required attribute: %s", attr)
		}
	}

	// Verify computed id exists.
	if _, ok := resp.Schema.Attributes["id"]; !ok {
		t.Error("Schema missing computed 'id' attribute")
	}
}

func TestNotificationRecipientResource_Configure(t *testing.T) {
	r, ok := NewNotificationRecipientResource().(*NotificationRecipientResource)
	if !ok {
		t.Fatal("Failed to cast to NotificationRecipientResource")
	}

	// Test with nil provider data (should not error)
	req := resource.ConfigureRequest{ProviderData: nil}
	resp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil data should not error, got: %v", resp.Diagnostics)
	}
}

func TestNotificationSettingsDataSource_Metadata(t *testing.T) {
	d := NewNotificationSettingsDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_notification_settings"
	if resp.TypeName != expected {
		t.Errorf("TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestNotificationSettingsDataSource_Schema(t *testing.T) {
	d := NewNotificationSettingsDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() errors: %v", resp.Diagnostics)
	}

	// Verify computed attributes exist.
	computedAttrs := []string{"id", "aad_tenant_id", "recipients"}
	for _, attr := range computedAttrs {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Schema missing computed attribute: %s", attr)
		}
	}
}

func TestNotificationSettingsDataSource_Configure(t *testing.T) {
	d, ok := NewNotificationSettingsDataSource().(*NotificationSettingsDataSource)
	if !ok {
		t.Fatal("Failed to cast to NotificationSettingsDataSource")
	}

	// Test with nil provider data (should not error)
	req := datasource.ConfigureRequest{ProviderData: nil}
	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil data should not error, got: %v", resp.Diagnostics)
	}
}
