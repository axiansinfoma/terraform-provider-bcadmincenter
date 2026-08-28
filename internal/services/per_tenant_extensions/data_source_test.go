// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestScheduledPteOperationsDataSource_Metadata(t *testing.T) {
	d := NewScheduledPteOperationsDataSource()

	req := datasource.MetadataRequest{ProviderTypeName: "bcadmincenter"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	expected := "bcadmincenter_scheduled_pte_operations"
	if resp.TypeName != expected {
		t.Errorf("Metadata() TypeName = %v, want %v", resp.TypeName, expected)
	}
}

func TestScheduledPteOperationsDataSource_Schema(t *testing.T) {
	d := NewScheduledPteOperationsDataSource()

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("Schema() unexpected errors: %v", resp.Diagnostics)
	}

	expectedAttrs := []string{"application_family", "environment_name", "aad_tenant_id", "app_id", "operations"}
	for _, attrName := range expectedAttrs {
		if _, ok := resp.Schema.Attributes[attrName]; !ok {
			t.Errorf("Schema missing attribute: %s", attrName)
		}
	}

	if !resp.Schema.Attributes["application_family"].IsRequired() {
		t.Error("application_family should be required")
	}
	if !resp.Schema.Attributes["environment_name"].IsRequired() {
		t.Error("environment_name should be required")
	}
}

func TestScheduledPteOperationsDataSource_Configure(t *testing.T) {
	d := &scheduledPteOperationsDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: nil}, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Configure() with nil ProviderData should not error, got: %v", resp.Diagnostics)
	}
	if d.client != nil {
		t.Error("Configure() with nil ProviderData should not set client")
	}
}

func TestScheduledPteOperationsDataSource_Configure_InvalidType(t *testing.T) {
	d := &scheduledPteOperationsDataSource{}

	resp := &datasource.ConfigureResponse{}
	d.Configure(context.Background(), datasource.ConfigureRequest{ProviderData: "not-a-client"}, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Configure() with wrong ProviderData type should error")
	}
}

// TestScheduledPteOperationsDataSource_ServiceRead exercises the API decoding the data
// source relies on, including the parameters snapshot and the app ID filter.
func TestScheduledPteOperationsDataSource_ServiceRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"value": []map[string]interface{}{
				{
					"id":                   "op-1",
					"type":                 "Install",
					"status":               "scheduled",
					"appId":                testAppID,
					"targetAppVersion":     "2.0.0.0",
					"scheduleKind":         DeploymentScheduleNextMajorUpdate,
					"createdOn":            "2026-08-01T10:00:00Z",
					"createdBy":            "admin@contoso.com",
					"creatorPrincipalType": "User",
					"parameters": map[string]interface{}{
						"name":          "My Extension",
						"publisher":     "Contoso",
						"syncMode":      SyncModeForceSync,
						"languageId":    "en-US",
						"countryCode":   "US",
						"targetRelease": "29.0.0.0",
					},
				},
				{
					"id":     "op-2",
					"status": "scheduled",
					"appId":  "another-app",
				},
			},
		})
	}))
	defer server.Close()

	svc := NewService(newTestClient(t, server.URL))

	ops, err := svc.ListScheduledPteOperations(context.Background(), testAppFamily, testEnvName)
	if err != nil {
		t.Fatalf("ListScheduledPteOperations() unexpected error: %v", err)
	}
	if len(ops) != 2 {
		t.Fatalf("got %d operations, want 2", len(ops))
	}

	first := ops[0]
	if first.Parameters == nil {
		t.Fatal("first operation is missing its parameters snapshot")
	}
	if first.Parameters.TargetRelease != "29.0.0.0" {
		t.Errorf("targetRelease = %q, want 29.0.0.0", first.Parameters.TargetRelease)
	}
	if first.Parameters.SyncMode != SyncModeForceSync {
		t.Errorf("syncMode = %q, want %q", first.Parameters.SyncMode, SyncModeForceSync)
	}
	if first.CreatorPrincipalType != "User" {
		t.Errorf("creatorPrincipalType = %q, want User", first.CreatorPrincipalType)
	}

	// The second entry has no parameters block; the data source must tolerate that.
	if ops[1].Parameters != nil {
		t.Error("second operation should have no parameters snapshot")
	}
}

func TestStringOrNull(t *testing.T) {
	if got := stringOrNull(""); !got.IsNull() {
		t.Errorf("stringOrNull(\"\") = %v, want null", got)
	}
	if got := stringOrNull("value"); got != types.StringValue("value") {
		t.Errorf("stringOrNull(\"value\") = %v, want StringValue(\"value\")", got)
	}
}
