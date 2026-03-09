// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package environments

import (
	"context"
	"fmt"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &environmentUpdatesDataSource{}
	_ datasource.DataSourceWithConfigure = &environmentUpdatesDataSource{}
)

// NewEnvironmentUpdatesDataSource is a helper function to simplify the provider implementation.
func NewEnvironmentUpdatesDataSource() datasource.DataSource {
	return &environmentUpdatesDataSource{}
}

// environmentUpdatesDataSource is the data source implementation.
type environmentUpdatesDataSource struct {
	client *client.Client
}

// environmentUpdatesDataSourceModel maps the data source schema data.
type environmentUpdatesDataSourceModel struct {
	ApplicationFamily types.String `tfsdk:"application_family"`
	EnvironmentName   types.String `tfsdk:"environment_name"`
	AadTenantID       types.String `tfsdk:"aad_tenant_id"`
	Updates           types.List   `tfsdk:"updates"`
}

// updateItemAttrTypes defines the object attribute types for a single update entry.
var updateItemAttrTypes = map[string]attr.Type{
	"target_version":             types.StringType,
	"available":                  types.BoolType,
	"selected":                   types.BoolType,
	"update_status":              types.StringType,
	"target_version_type":        types.StringType,
	"scheduled_datetime":         types.StringType,
	"ignore_update_window":       types.BoolType,
	"rollout_status":             types.StringType,
	"latest_selectable_datetime": types.StringType,
}

// Metadata returns the data source type name.
func (d *environmentUpdatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environment_updates"
}

// Schema defines the schema for the data source.
func (d *environmentUpdatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves the list of available version updates for a Business Central environment. " +
			"Each entry represents a version that can be scheduled for upgrade, along with its current scheduling state.",

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
			"updates": schema.ListNestedAttribute{
				MarkdownDescription: "The list of available updates for the environment.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"target_version": schema.StringAttribute{
							MarkdownDescription: "The target version for this update entry.",
							Computed:            true,
						},
						"available": schema.BoolAttribute{
							MarkdownDescription: "Whether this version is available for selection.",
							Computed:            true,
						},
						"selected": schema.BoolAttribute{
							MarkdownDescription: "Whether this version is currently selected for upgrade.",
							Computed:            true,
						},
						"update_status": schema.StringAttribute{
							MarkdownDescription: "The current update status (e.g. `\"scheduled\"`, `\"running\"`, `\"failed\"`).",
							Computed:            true,
						},
						"target_version_type": schema.StringAttribute{
							MarkdownDescription: "The type of the target version (e.g. `\"Cumulative\"`, `\"Mandatory\"`).",
							Computed:            true,
						},
						"scheduled_datetime": schema.StringAttribute{
							MarkdownDescription: "The datetime at which the upgrade is scheduled to run.",
							Computed:            true,
						},
						"ignore_update_window": schema.BoolAttribute{
							MarkdownDescription: "Whether the upgrade ignores the environment's configured update window.",
							Computed:            true,
						},
						"rollout_status": schema.StringAttribute{
							MarkdownDescription: "The rollout status of the update (e.g. `\"Active\"`, `\"UnderMaintenance\"`, `\"Postponed\"`).",
							Computed:            true,
						},
						"latest_selectable_datetime": schema.StringAttribute{
							MarkdownDescription: "The latest datetime by which this update can be scheduled.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *environmentUpdatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read fetches the available updates for an environment.
func (d *environmentUpdatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state environmentUpdatesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tenantID := d.client.GetTenantID()
	if !state.AadTenantID.IsNull() && !state.AadTenantID.IsUnknown() {
		tenantID = state.AadTenantID.ValueString()
	}

	svc := NewService(d.client.ForTenant(tenantID))

	updates, err := svc.GetUpdates(ctx, state.ApplicationFamily.ValueString(), state.EnvironmentName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Environment Updates",
			fmt.Sprintf("An error occurred while reading updates for environment %q: %s",
				state.EnvironmentName.ValueString(), err.Error()),
		)
		return
	}

	state.AadTenantID = types.StringValue(tenantID)

	updateObjects := make([]attr.Value, 0, len(updates))
	for _, u := range updates {
		scheduledDatetime := types.StringNull()
		ignoreUpdateWindow := types.BoolValue(false)
		rolloutStatus := types.StringNull()
		latestSelectableDateTime := types.StringNull()

		if u.ScheduleDetails != nil {
			if u.ScheduleDetails.SelectedDateTime != "" {
				scheduledDatetime = types.StringValue(u.ScheduleDetails.SelectedDateTime)
			}
			ignoreUpdateWindow = types.BoolValue(u.ScheduleDetails.IgnoreUpdateWindow)
			if u.ScheduleDetails.RolloutStatus != "" {
				rolloutStatus = types.StringValue(u.ScheduleDetails.RolloutStatus)
			}
			if u.ScheduleDetails.LatestSelectableDateTime != "" {
				latestSelectableDateTime = types.StringValue(u.ScheduleDetails.LatestSelectableDateTime)
			}
		}

		updateStatus := types.StringNull()
		if u.UpdateStatus != "" {
			updateStatus = types.StringValue(u.UpdateStatus)
		}

		targetVersionType := types.StringNull()
		if u.TargetVersionType != "" {
			targetVersionType = types.StringValue(u.TargetVersionType)
		}

		obj, diags := types.ObjectValue(updateItemAttrTypes, map[string]attr.Value{
			"target_version":             types.StringValue(u.TargetVersion),
			"available":                  types.BoolValue(u.Available),
			"selected":                   types.BoolValue(u.Selected),
			"update_status":              updateStatus,
			"target_version_type":        targetVersionType,
			"scheduled_datetime":         scheduledDatetime,
			"ignore_update_window":       ignoreUpdateWindow,
			"rollout_status":             rolloutStatus,
			"latest_selectable_datetime": latestSelectableDateTime,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateObjects = append(updateObjects, obj)
	}

	updatesList, diags := types.ListValue(types.ObjectType{AttrTypes: updateItemAttrTypes}, updateObjects)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Updates = updatesList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
