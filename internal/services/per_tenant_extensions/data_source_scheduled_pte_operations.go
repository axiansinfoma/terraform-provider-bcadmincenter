// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &scheduledPteOperationsDataSource{}
	_ datasource.DataSourceWithConfigure = &scheduledPteOperationsDataSource{}
)

// NewScheduledPteOperationsDataSource is a helper function to simplify the provider implementation.
func NewScheduledPteOperationsDataSource() datasource.DataSource {
	return &scheduledPteOperationsDataSource{}
}

// scheduledPteOperationsDataSource is the data source implementation.
type scheduledPteOperationsDataSource struct {
	client *client.Client
}

// scheduledPteOperationsDataSourceModel maps the data source schema data.
type scheduledPteOperationsDataSourceModel struct {
	ApplicationFamily types.String `tfsdk:"application_family"`
	EnvironmentName   types.String `tfsdk:"environment_name"`
	AadTenantID       types.String `tfsdk:"aad_tenant_id"`
	AppID             types.String `tfsdk:"app_id"`
	Operations        types.List   `tfsdk:"operations"`
}

// scheduledPteOperationAttrTypes defines the object attribute types for a single
// scheduled operation entry.
var scheduledPteOperationAttrTypes = map[string]attr.Type{
	"id":                     types.StringType,
	"type":                   types.StringType,
	"app_id":                 types.StringType,
	"name":                   types.StringType,
	"publisher":              types.StringType,
	"source_app_version":     types.StringType,
	"target_app_version":     types.StringType,
	"status":                 types.StringType,
	"schedule_kind":          types.StringType,
	"sync_mode":              types.StringType,
	"language_id":            types.StringType,
	"country_code":           types.StringType,
	"target_release":         types.StringType,
	"created_on":             types.StringType,
	"created_by":             types.StringType,
	"creator_principal_type": types.StringType,
}

// Metadata returns the data source type name.
func (d *scheduledPteOperationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scheduled_pte_operations"
}

// Schema defines the schema for the data source.
func (d *scheduledPteOperationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves the per-tenant extension (PTE) installs and updates that are scheduled for a future " +
			"deployment window on a Business Central environment.\n\n" +
			"Scheduled PTE versions survive an uninstall, so this data source is useful for auditing what will still be " +
			"deployed to an environment.\n\n" +
			"~> **Note:** Requires Admin Center API version 2.29 or later.",

		Attributes: map[string]schema.Attribute{
			"application_family": schema.StringAttribute{
				MarkdownDescription: "The application family of the environment (e.g. `\"BusinessCentral\"`).",
				Required:            true,
			},
			"environment_name": schema.StringAttribute{
				MarkdownDescription: "The name of the environment.",
				Required:            true,
			},
			"aad_tenant_id": schema.StringAttribute{
				MarkdownDescription: "The Azure AD tenant ID. If not specified, the provider's configured tenant ID is used.",
				Optional:            true,
				Computed:            true,
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "Optional filter. When set, only scheduled operations targeting this extension are returned.",
				Optional:            true,
			},
			"operations": schema.ListNestedAttribute{
				MarkdownDescription: "The scheduled per-tenant extension operations for the environment.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "ID of the scheduled operation.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Operation type. Always `\"Install\"` for this endpoint.",
							Computed:            true,
						},
						"app_id": schema.StringAttribute{
							MarkdownDescription: "ID of the targeted extension.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the extension, read from the uploaded `.app` package metadata.",
							Computed:            true,
						},
						"publisher": schema.StringAttribute{
							MarkdownDescription: "Publisher of the extension, read from the uploaded `.app` package metadata.",
							Computed:            true,
						},
						"source_app_version": schema.StringAttribute{
							MarkdownDescription: "Source version. Empty for scheduled installs.",
							Computed:            true,
						},
						"target_app_version": schema.StringAttribute{
							MarkdownDescription: "Version of the extension that will be installed.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "Operation status. Always `\"scheduled\"` for this endpoint.",
							Computed:            true,
						},
						"schedule_kind": schema.StringAttribute{
							MarkdownDescription: "When the install will run (`\"Immediate\"`, `\"UpdateWindow\"`, `\"NextMinorUpdate\"`, or `\"NextMajorUpdate\"`).",
							Computed:            true,
						},
						"sync_mode": schema.StringAttribute{
							MarkdownDescription: "Schema synchronisation mode supplied at upload time (`\"Add\"` or `\"ForceSync\"`).",
							Computed:            true,
						},
						"language_id": schema.StringAttribute{
							MarkdownDescription: "Microsoft Language Code ID the install was scheduled with (e.g. `\"en-US\"`).",
							Computed:            true,
						},
						"country_code": schema.StringAttribute{
							MarkdownDescription: "Country code the install was scheduled with (e.g. `\"US\"`).",
							Computed:            true,
						},
						"target_release": schema.StringAttribute{
							MarkdownDescription: "Platform release the staged entry targets (e.g. `\"29.0.0.0\"`).",
							Computed:            true,
						},
						"created_on": schema.StringAttribute{
							MarkdownDescription: "Date and time the schedule entry was created.",
							Computed:            true,
						},
						"created_by": schema.StringAttribute{
							MarkdownDescription: "Email address if scheduled by a user, or app ID if scheduled by a service principal.",
							Computed:            true,
						},
						"creator_principal_type": schema.StringAttribute{
							MarkdownDescription: "Principal type that created the operation (`\"User\"` or `\"App\"`).",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *scheduledPteOperationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

// stringOrNull converts an API string into a Terraform value, mapping "" to null.
func stringOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

// Read fetches the scheduled per-tenant extension operations for an environment.
func (d *scheduledPteOperationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state scheduledPteOperationsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := d.client.GetTenantID()
	if !state.AadTenantID.IsNull() && !state.AadTenantID.IsUnknown() {
		tenantID = state.AadTenantID.ValueString()
	}

	svc := NewService(d.client.ForTenant(tenantID))

	operations, err := svc.ListScheduledPteOperations(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Scheduled Per-Tenant Extension Operations",
			fmt.Sprintf("An error occurred while reading scheduled PTE operations for environment %q: %s",
				state.EnvironmentName.ValueString(), err.Error()),
		)
		return
	}

	state.AadTenantID = types.StringValue(tenantID)

	appIDFilter := ""
	if !state.AppID.IsNull() && !state.AppID.IsUnknown() {
		appIDFilter = state.AppID.ValueString()
	}

	operationObjects := make([]attr.Value, 0, len(operations))
	for i := range operations {
		op := &operations[i]

		if appIDFilter != "" && !strings.EqualFold(op.AppID, appIDFilter) {
			continue
		}

		var params ScheduledPteOperationParameters
		if op.Parameters != nil {
			params = *op.Parameters
		}

		targetVersion := op.TargetVersionValue()
		if targetVersion == "" {
			targetVersion = params.TargetAppVersion
		}

		scheduleKind := op.ScheduleKind
		if scheduleKind == "" {
			scheduleKind = params.ScheduleKind
		}

		obj, diags := types.ObjectValue(scheduledPteOperationAttrTypes, map[string]attr.Value{
			"id":                     types.StringValue(op.ID),
			"type":                   stringOrNull(op.Type),
			"app_id":                 stringOrNull(op.AppID),
			"name":                   stringOrNull(params.Name),
			"publisher":              stringOrNull(params.Publisher),
			"source_app_version":     stringOrNull(op.SourceVersionValue()),
			"target_app_version":     stringOrNull(targetVersion),
			"status":                 stringOrNull(op.Status),
			"schedule_kind":          stringOrNull(scheduleKind),
			"sync_mode":              stringOrNull(params.SyncMode),
			"language_id":            stringOrNull(params.LanguageID),
			"country_code":           stringOrNull(params.CountryCode),
			"target_release":         stringOrNull(params.TargetRelease),
			"created_on":             stringOrNull(op.CreatedOn),
			"created_by":             stringOrNull(op.CreatedBy),
			"creator_principal_type": stringOrNull(op.CreatorPrincipalType),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		operationObjects = append(operationObjects, obj)
	}

	operationsList, diags := types.ListValue(types.ObjectType{AttrTypes: scheduledPteOperationAttrTypes}, operationObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Operations = operationsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
