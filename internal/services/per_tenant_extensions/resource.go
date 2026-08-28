// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &PerTenantExtensionResource{}
	_ resource.ResourceWithConfigure   = &PerTenantExtensionResource{}
	_ resource.ResourceWithImportState = &PerTenantExtensionResource{}
)

// defaultOperationTimeout bounds how long the resource waits for an install, update, or
// uninstall operation to reach a terminal state.
const defaultOperationTimeout = 60 * time.Minute

// NewPerTenantExtensionResource is a helper function to simplify the provider implementation.
func NewPerTenantExtensionResource() resource.Resource {
	return &PerTenantExtensionResource{}
}

// PerTenantExtensionResource is the resource implementation.
type PerTenantExtensionResource struct {
	client *client.Client
}

// PerTenantExtensionResourceModel describes the resource data model.
type PerTenantExtensionResourceModel struct {
	ID                                types.String `tfsdk:"id"`
	AADTenantID                       types.String `tfsdk:"aad_tenant_id"`
	EnvironmentName                   types.String `tfsdk:"environment_name"`
	ApplicationFamily                 types.String `tfsdk:"application_family"`
	FilePath                          types.String `tfsdk:"file_path"`
	FileContent                       types.String `tfsdk:"file_content"`
	FileName                          types.String `tfsdk:"file_name"`
	FileSHA256                        types.String `tfsdk:"file_sha256"`
	DeploymentSchedule                types.String `tfsdk:"deployment_schedule"`
	SyncMode                          types.String `tfsdk:"sync_mode"`
	LanguageID                        types.String `tfsdk:"language_id"`
	AcceptIsvEula                     types.Bool   `tfsdk:"accept_isv_eula"`
	InstallOrUpdateNeededDependencies types.Bool   `tfsdk:"install_or_update_needed_dependencies"`
	DeleteData                        types.Bool   `tfsdk:"delete_data"`
	UninstallDependents               types.Bool   `tfsdk:"uninstall_dependents"`
	UninstallInUpdateWindow           types.Bool   `tfsdk:"uninstall_in_update_window"`
	CancelScheduledOnDestroy          types.Bool   `tfsdk:"cancel_scheduled_on_destroy"`
	Timeouts                          types.Object `tfsdk:"timeouts"`
	AppID                             types.String `tfsdk:"app_id"`
	DisplayName                       types.String `tfsdk:"display_name"`
	Publisher                         types.String `tfsdk:"publisher"`
	Version                           types.String `tfsdk:"version"`
	State                             types.String `tfsdk:"state"`
	AppType                           types.String `tfsdk:"app_type"`
	LastOperationID                   types.String `tfsdk:"last_operation_id"`
	PendingTargetVersion              types.String `tfsdk:"pending_target_version"`
	PendingScheduleKind               types.String `tfsdk:"pending_schedule_kind"`
}

// Metadata returns the resource type name.
func (r *PerTenantExtensionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_per_tenant_extension"
}

// Schema defines the schema for the resource.
func (r *PerTenantExtensionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the full lifecycle of a Per-Tenant Extension (PTE) in a Business Central environment.\n\n" +
			"This resource uploads a `.app` extension package, installs or schedules it, updates it when the package changes, " +
			"and uninstalls it on destroy. All operations use the **Business Central Admin Center API** PTE endpoints, " +
			"which require **API version 2.29 or later**.\n\n" +
			"~> **Note:** Exactly one of `file_path` or `file_content` must be set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ARM-like resource ID (format: `/tenants/{tenantId}/providers/Microsoft.Dynamics365.BusinessCentral/applications/{applicationFamily}/environments/{environmentName}/perTenantExtensions/{appId}`).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"aad_tenant_id": schema.StringAttribute{
				MarkdownDescription: "The Azure AD tenant ID. If not specified, defaults to the provider's configured tenant ID. Changing this forces a new resource to be created.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"environment_name": schema.StringAttribute{
				MarkdownDescription: "The name of the target environment. Changing this forces a new resource to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_family": schema.StringAttribute{
				MarkdownDescription: "The application family of the environment (e.g. `\"BusinessCentral\"`). Changing this forces a new resource to be created.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_path": schema.StringAttribute{
				MarkdownDescription: "Local path to the `.app` file. Mutually exclusive with `file_content`.",
				Optional:            true,
			},
			"file_content": schema.StringAttribute{
				MarkdownDescription: "Base64-encoded `.app` file bytes. Mutually exclusive with `file_path`. " +
					"Enables passing content directly from a data source (e.g. `azurerm_storage_blob`). " +
					"When set, `file_name` should also be set because the API requires an uploaded file name ending in `.app`.",
				Optional:  true,
				Sensitive: true,
			},
			"file_name": schema.StringAttribute{
				MarkdownDescription: "File name submitted with the upload. Must end in `.app`. " +
					"Defaults to the base name of `file_path`, or `extension.app` when `file_content` is used.",
				Optional: true,
				Computed: true,
			},
			"file_sha256": schema.StringAttribute{
				MarkdownDescription: "SHA-256 hash of the `.app` file content. Drives change detection — changing this value triggers an update.",
				Required:            true,
			},
			"deployment_schedule": schema.StringAttribute{
				MarkdownDescription: "When the uploaded package is installed. One of `\"Immediate\"` (default), `\"UpdateWindow\"`, " +
					"`\"NextMinorUpdate\"`, or `\"NextMajorUpdate\"`. Must be `\"Immediate\"` or `\"UpdateWindow\"` for a PTE that is not yet " +
					"installed; updates to an installed PTE may use any value. The pre-2.29 names `\"Current version\"`, " +
					"`\"Next minor version\"`, and `\"Next major version\"` are accepted and mapped to their modern equivalents.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(DefaultDeploymentSchedule),
				Validators: []validator.String{
					stringvalidator.OneOf(
						DeploymentScheduleImmediate,
						DeploymentScheduleUpdateWindow,
						DeploymentScheduleNextMinorUpdate,
						DeploymentScheduleNextMajorUpdate,
						"Current version",
						"Next minor version",
						"Next major version",
					),
				},
			},
			"sync_mode": schema.StringAttribute{
				MarkdownDescription: "Schema synchronisation mode applied during install. One of `\"Add\"` (default) or `\"ForceSync\"`. " +
					"The pre-2.29 name `\"Force Sync\"` is accepted and mapped to `\"ForceSync\"`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(DefaultSyncMode),
				Validators: []validator.String{
					stringvalidator.OneOf(SyncModeAdd, SyncModeForceSync, "Force Sync"),
				},
			},
			"language_id": schema.StringAttribute{
				MarkdownDescription: "Microsoft Language Code ID applied during install (e.g. `\"en-US\"`). Defaults to `\"en-US\"`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(DefaultLanguageID),
			},
			"accept_isv_eula": schema.BoolAttribute{
				MarkdownDescription: "Must be set to `true` for the installation to proceed. Setting it to `true` accepts the publisher's " +
					"end-user license terms and the associated Microsoft Marketplace terms, exactly as when installing the extension " +
					"through the Business Central admin center.",
				Required: true,
			},
			"install_or_update_needed_dependencies": schema.BoolAttribute{
				MarkdownDescription: "When `true`, dependencies of the uploaded extension are installed or updated automatically. " +
					"When `false` (default), the install fails and returns the missing dependencies as error details.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"delete_data": schema.BoolAttribute{
				MarkdownDescription: "When `true`, the uninstall performed on destroy also deletes the extension's data and syncs it clean (irreversible). Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"uninstall_dependents": schema.BoolAttribute{
				MarkdownDescription: "When `true`, apps that depend on this extension are uninstalled along with it on destroy. " +
					"When `false` (default), the uninstall fails and returns the dependent apps as error details.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"uninstall_in_update_window": schema.BoolAttribute{
				MarkdownDescription: "When `true`, the uninstall performed on destroy runs only inside the environment's update window " +
					"and is reported as `scheduled` until then. Defaults to `false`, which uninstalls immediately.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"cancel_scheduled_on_destroy": schema.BoolAttribute{
				MarkdownDescription: "When `true` (default), any versions of this extension still scheduled for a future deployment window " +
					"are cancelled on destroy before the uninstall runs. Uninstalling alone does not remove scheduled versions, so they " +
					"would otherwise reinstall the extension later.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(true),
			},
			"timeouts": schema.SingleNestedAttribute{
				MarkdownDescription: "Timeout configuration for the resource operations. Values are Go duration strings (e.g. `\"90m\"`).",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"create": schema.StringAttribute{
						MarkdownDescription: "Timeout for create operations. Defaults to 60 minutes.",
						Optional:            true,
					},
					"update": schema.StringAttribute{
						MarkdownDescription: "Timeout for update operations. Defaults to 60 minutes.",
						Optional:            true,
					},
					"delete": schema.StringAttribute{
						MarkdownDescription: "Timeout for delete operations. Defaults to 60 minutes.",
						Optional:            true,
					},
				},
			},
			"app_id": schema.StringAttribute{
				MarkdownDescription: "Stable extension identity read from the uploaded `.app` package. Remains constant across version updates.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "Display name of the extension.",
				Computed:            true,
			},
			"publisher": schema.StringAttribute{
				MarkdownDescription: "Publisher of the extension.",
				Computed:            true,
			},
			"version": schema.StringAttribute{
				MarkdownDescription: "Currently installed version. Empty while the first install is still scheduled for a future deployment window.",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Install state reported by the API (e.g. `\"installed\"`, `\"updatePending\"`, `\"updating\"`).",
				Computed:            true,
			},
			"app_type": schema.StringAttribute{
				MarkdownDescription: "App type reported by the API. `\"tenant\"` for per-tenant extensions.",
				Computed:            true,
			},
			"last_operation_id": schema.StringAttribute{
				MarkdownDescription: "ID of the most recent install or update operation triggered for this extension.",
				Computed:            true,
			},
			"pending_target_version": schema.StringAttribute{
				MarkdownDescription: "Target version of an install or update that is still scheduled for a future deployment window. " +
					"Empty when nothing is pending.",
				Computed: true,
			},
			"pending_schedule_kind": schema.StringAttribute{
				MarkdownDescription: "Schedule kind of a pending install or update (e.g. `\"UpdateWindow\"`, `\"NextMinorUpdate\"`). " +
					"Empty when nothing is pending.",
				Computed: true,
			},
		},
	}
}

// Configure stores the provider client on the resource.
func (r *PerTenantExtensionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// resolveFileBytes returns the raw .app bytes from either file_path or file_content.
func resolveFileBytes(data *PerTenantExtensionResourceModel) ([]byte, error) {
	if hasValue(data.FilePath) {
		return os.ReadFile(data.FilePath.ValueString())
	}

	if hasValue(data.FileContent) {
		raw, err := base64.StdEncoding.DecodeString(data.FileContent.ValueString())
		if err != nil {
			return nil, fmt.Errorf("failed to base64-decode file_content: %w", err)
		}
		return raw, nil
	}

	return nil, fmt.Errorf("exactly one of file_path or file_content must be set")
}

// hasValue reports whether a string attribute carries a usable non-empty value.
func hasValue(v types.String) bool {
	return !v.IsNull() && !v.IsUnknown() && v.ValueString() != ""
}

// validateFileInputs checks that exactly one of file_path / file_content is provided.
func validateFileInputs(data *PerTenantExtensionResourceModel) error {
	hasFilePath := hasValue(data.FilePath)
	hasFileContent := hasValue(data.FileContent)

	if hasFilePath && hasFileContent {
		return fmt.Errorf("exactly one of file_path or file_content must be set, but both are set")
	}

	if !hasFilePath && !hasFileContent {
		return fmt.Errorf("exactly one of file_path or file_content must be set, but neither is set")
	}

	return nil
}

// resolveFileName determines the file name submitted with the upload. The API requires a
// name ending in .app, so an explicit file_name wins, then the base name of file_path,
// then a generic fallback for base64 content.
func resolveFileName(data *PerTenantExtensionResourceModel) string {
	if hasValue(data.FileName) {
		return data.FileName.ValueString()
	}
	if hasValue(data.FilePath) {
		return filepath.Base(data.FilePath.ValueString())
	}
	return "extension.app"
}

// operationTimeout reads one timeout from the optional `timeouts` block, falling back to
// defaultOperationTimeout when unset or unparseable.
func operationTimeout(ctx context.Context, timeouts types.Object, key string) time.Duration {
	if timeouts.IsNull() || timeouts.IsUnknown() {
		return defaultOperationTimeout
	}

	attrs := timeouts.Attributes()
	raw, ok := attrs[key]
	if !ok {
		return defaultOperationTimeout
	}

	value, ok := raw.(types.String)
	if !ok || !hasValue(value) {
		return defaultOperationTimeout
	}

	parsed, err := time.ParseDuration(value.ValueString())
	if err != nil || parsed <= 0 {
		tflog.Warn(ctx, "Ignoring unparseable timeout value", map[string]interface{}{
			"timeout": key,
			"value":   value.ValueString(),
		})
		return defaultOperationTimeout
	}

	return parsed
}

// uploadAndWait uploads the .app package and waits for the resulting operation to reach a
// terminal state, unless it was deferred to a future deployment window.
func (r *PerTenantExtensionResource) uploadAndWait(ctx context.Context, data *PerTenantExtensionResourceModel, svc *Service, timeout time.Duration) (*AppOperation, error) {
	if !data.AcceptIsvEula.ValueBool() {
		return nil, fmt.Errorf("accept_isv_eula must be set to true; the API rejects per-tenant extension installs without it")
	}

	fileBytes, err := resolveFileBytes(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read extension file: %w", err)
	}

	installReq := &PteInstallRequest{
		FileName:                          resolveFileName(data),
		Content:                           fileBytes,
		DeploymentSchedule:                data.DeploymentSchedule.ValueString(),
		SyncMode:                          data.SyncMode.ValueString(),
		LanguageID:                        data.LanguageID.ValueString(),
		AcceptIsvEula:                     data.AcceptIsvEula.ValueBool(),
		InstallOrUpdateNeededDependencies: data.InstallOrUpdateNeededDependencies.ValueBool(),
	}

	operation, err := svc.UploadAndInstall(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), installReq)
	if err != nil {
		return nil, err
	}

	// Only a non-immediate schedule actually defers the deployment. An immediate install
	// can also report "scheduled" while it is queued, and that must be waited through.
	deferred := !strings.EqualFold(NormalizeDeploymentSchedule(data.DeploymentSchedule.ValueString()), DeploymentScheduleImmediate)

	if deferred && operation.IsScheduled() {
		tflog.Info(ctx, "Per-tenant extension install deferred to a future deployment window", map[string]interface{}{
			"operation_id":  operation.ID,
			"app_id":        operation.AppID,
			"schedule_kind": operation.ScheduleKind,
		})
		return operation, nil
	}

	return svc.WaitForOperation(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), operation.AppID, operation.ID, timeout, deferred)
}

// applyOperation records the outcome of an install or update operation on the model and
// then refreshes the installed-app details when the operation actually ran.
func (r *PerTenantExtensionResource) applyOperation(ctx context.Context, data *PerTenantExtensionResourceModel, svc *Service, operation *AppOperation) error {
	if operation == nil {
		return fmt.Errorf("no operation was returned for the per-tenant extension upload")
	}

	if operation.AppID != "" {
		data.AppID = types.StringValue(operation.AppID)
	}
	data.LastOperationID = types.StringValue(operation.ID)
	data.FileName = types.StringValue(resolveFileName(data))

	if operation.IsScheduled() {
		// The package is staged but not installed. Record what is pending and read back
		// whatever version (if any) is currently installed.
		data.PendingTargetVersion = types.StringValue(operation.TargetVersionValue())
		data.PendingScheduleKind = types.StringValue(operation.ScheduleKind)
		return r.readAppInto(ctx, data, svc, false)
	}

	data.PendingTargetVersion = types.StringValue("")
	data.PendingScheduleKind = types.StringValue("")

	return r.readAppInto(ctx, data, svc, true)
}

// readAppInto populates the computed app attributes from the `apps` endpoint.
// When required is true, a missing app is an error; otherwise the computed attributes are
// left empty, which is the expected shape for an install still awaiting its window.
func (r *PerTenantExtensionResource) readAppInto(ctx context.Context, data *PerTenantExtensionResourceModel, svc *Service, required bool) error {
	app, err := svc.GetApp(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString())
	if err != nil {
		return fmt.Errorf("failed to read extension details: %w", err)
	}

	if app == nil {
		if required {
			return fmt.Errorf("extension %q was not found on environment %q after the install operation succeeded",
				data.AppID.ValueString(), data.EnvironmentName.ValueString())
		}
		data.DisplayName = types.StringValue("")
		data.Publisher = types.StringValue("")
		data.Version = types.StringValue("")
		data.State = types.StringValue("")
		data.AppType = types.StringValue("")
		return nil
	}

	data.DisplayName = types.StringValue(app.Name)
	data.Publisher = types.StringValue(app.Publisher)
	data.Version = types.StringValue(app.Version)
	data.State = types.StringValue(app.State)
	data.AppType = types.StringValue(app.AppType)

	return nil
}

// Create uploads the package and installs (or schedules) the extension.
func (r *PerTenantExtensionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PerTenantExtensionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateFileInputs(&data); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}

	// Use provider tenant ID if not explicitly set.
	if !hasValue(data.AADTenantID) {
		data.AADTenantID = types.StringValue(r.client.GetTenantID())
	}

	svc := NewService(r.client.ForTenant(data.AADTenantID.ValueString()))

	operation, err := r.uploadAndWait(ctx, &data, svc, operationTimeout(ctx, data.Timeouts, "create"))
	if err != nil {
		resp.Diagnostics.AddError("Failed to install per-tenant extension", err.Error())
		return
	}

	if err := r.applyOperation(ctx, &data, svc, operation); err != nil {
		resp.Diagnostics.AddError("Failed to read extension details after install", err.Error())
		return
	}

	data.ID = types.StringValue(BuildPerTenantExtensionID(
		data.AADTenantID.ValueString(),
		data.ApplicationFamily.ValueString(),
		data.EnvironmentName.ValueString(),
		data.AppID.ValueString(),
	))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the state from the Admin Center API.
func (r *PerTenantExtensionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PerTenantExtensionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !hasValue(data.AppID) {
		// Nothing identifies the extension — treat it as gone.
		resp.State.RemoveResource(ctx)
		return
	}

	svc := NewService(r.client.ForTenant(data.AADTenantID.ValueString()))

	app, err := svc.GetApp(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read per-tenant extension", err.Error())
		return
	}

	// Refresh what is still scheduled for a future deployment window so a pending
	// install that has since run stops being reported as pending.
	scheduled, err := svc.GetScheduledPteOperationsForApp(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString())
	if err != nil {
		// A tenant on an API version without the endpoint should not break refresh of an
		// otherwise healthy resource.
		tflog.Warn(ctx, "Failed to list scheduled per-tenant extension operations", map[string]interface{}{
			"app_id": data.AppID.ValueString(),
			"error":  err.Error(),
		})
		scheduled = nil
	}

	if app == nil && len(scheduled) == 0 {
		// Neither installed nor pending — the extension is gone.
		resp.State.RemoveResource(ctx)
		return
	}

	if app != nil {
		data.DisplayName = types.StringValue(app.Name)
		data.Publisher = types.StringValue(app.Publisher)
		data.Version = types.StringValue(app.Version)
		data.State = types.StringValue(app.State)
		data.AppType = types.StringValue(app.AppType)
	} else {
		data.DisplayName = types.StringValue("")
		data.Publisher = types.StringValue("")
		data.Version = types.StringValue("")
		data.State = types.StringValue("")
		data.AppType = types.StringValue("")
	}

	if len(scheduled) > 0 {
		data.PendingTargetVersion = types.StringValue(scheduled[0].TargetVersionValue())
		data.PendingScheduleKind = types.StringValue(scheduled[0].ScheduleKind)
	} else {
		data.PendingTargetVersion = types.StringValue("")
		data.PendingScheduleKind = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update uploads a new version of the PTE.
func (r *PerTenantExtensionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PerTenantExtensionResourceModel
	var state PerTenantExtensionResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateFileInputs(&data); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}

	// aad_tenant_id forces replacement, so the state value always applies here.
	data.AADTenantID = state.AADTenantID
	data.AppID = state.AppID

	svc := NewService(r.client.ForTenant(data.AADTenantID.ValueString()))

	operation, err := r.uploadAndWait(ctx, &data, svc, operationTimeout(ctx, data.Timeouts, "update"))
	if err != nil {
		resp.Diagnostics.AddError("Failed to update per-tenant extension", err.Error())
		return
	}

	if err := r.applyOperation(ctx, &data, svc, operation); err != nil {
		resp.Diagnostics.AddError("Failed to read extension details after update", err.Error())
		return
	}

	data.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// cancelScheduledVersions removes every version of the extension still staged for a
// future deployment window. Uninstalling does not clear these, so leaving them in place
// would reinstall the extension after destroy.
func (r *PerTenantExtensionResource) cancelScheduledVersions(ctx context.Context, data *PerTenantExtensionResourceModel, svc *Service) {
	scheduled, err := svc.GetScheduledPteOperationsForApp(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString())
	if err != nil {
		tflog.Warn(ctx, "Failed to list scheduled per-tenant extension versions before destroy", map[string]interface{}{
			"app_id": data.AppID.ValueString(),
			"error":  err.Error(),
		})
		return
	}

	for i := range scheduled {
		op := &scheduled[i]
		targetVersion := op.TargetVersionValue()
		scheduleKind := op.ScheduleKind
		if op.Parameters != nil {
			if targetVersion == "" {
				targetVersion = op.Parameters.TargetAppVersion
			}
			if scheduleKind == "" {
				scheduleKind = op.Parameters.ScheduleKind
			}
		}

		if targetVersion == "" || scheduleKind == "" || strings.EqualFold(scheduleKind, DeploymentScheduleImmediate) {
			// An immediate operation is already running and has no staged package to remove.
			continue
		}

		if _, err := svc.RemoveScheduledPteVersion(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString(),
			&RemoveScheduledPteVersionRequest{TargetVersion: targetVersion, ScheduleKind: scheduleKind}); err != nil {
			if IsNotFoundError(err) {
				continue
			}
			tflog.Warn(ctx, "Failed to cancel scheduled per-tenant extension version", map[string]interface{}{
				"app_id":         data.AppID.ValueString(),
				"target_version": targetVersion,
				"schedule_kind":  scheduleKind,
				"error":          err.Error(),
			})
			continue
		}

		tflog.Info(ctx, "Cancelled scheduled per-tenant extension version", map[string]interface{}{
			"app_id":         data.AppID.ValueString(),
			"target_version": targetVersion,
			"schedule_kind":  scheduleKind,
		})
	}
}

// Delete cancels any scheduled versions and uninstalls the PTE.
func (r *PerTenantExtensionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PerTenantExtensionResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !hasValue(data.AppID) {
		// Nothing was ever installed.
		return
	}

	svc := NewService(r.client.ForTenant(data.AADTenantID.ValueString()))

	if data.CancelScheduledOnDestroy.ValueBool() {
		r.cancelScheduledVersions(ctx, &data, svc)
	}

	app, err := svc.GetApp(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read per-tenant extension before uninstall", err.Error())
		return
	}

	if app == nil {
		// Already uninstalled; the scheduled versions above were the only thing left.
		return
	}

	operation, err := svc.Uninstall(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString(),
		&UninstallAppRequest{
			UseEnvironmentUpdateWindow: data.UninstallInUpdateWindow.ValueBool(),
			UninstallDependents:        data.UninstallDependents.ValueBool(),
			DeleteData:                 data.DeleteData.ValueBool(),
		})
	if err != nil {
		if IsNotFoundError(err) {
			return
		}
		resp.Diagnostics.AddError("Failed to uninstall per-tenant extension", err.Error())
		return
	}

	// An uninstall reports "scheduled" both while it is merely queued and when it has been
	// deferred to the update window. Only the latter must not be waited on, and only the
	// configuration says which one it is.
	deferred := data.UninstallInUpdateWindow.ValueBool()

	if deferred && operation.IsScheduled() {
		tflog.Info(ctx, "Per-tenant extension uninstall deferred to the environment update window", map[string]interface{}{
			"operation_id": operation.ID,
			"app_id":       data.AppID.ValueString(),
		})
		return
	}

	timeout := operationTimeout(ctx, data.Timeouts, "delete")

	// A missing operation ID would make the operation lookup fall back to the app's whole
	// history, so only poll the operation when the API actually returned one. The removal
	// wait below is authoritative either way.
	if operation.ID != "" {
		if _, err := svc.WaitForOperation(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString(), operation.ID,
			timeout, deferred); err != nil {
			resp.Diagnostics.AddError("Failed waiting for per-tenant extension uninstall", err.Error())
			return
		}
	} else {
		tflog.Warn(ctx, "Uninstall response contained no operation id; falling back to polling for removal", map[string]interface{}{
			"app_id": data.AppID.ValueString(),
		})
	}

	// The uninstall operation reports success before the app actually disappears, so wait
	// for it to leave the apps list. Otherwise a destroy immediately followed by an apply
	// fails with "already installed".
	if err := svc.WaitForAppRemoval(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString(), timeout); err != nil {
		resp.Diagnostics.AddError("Failed waiting for per-tenant extension removal", err.Error())
		return
	}
}

// ImportState imports a per-tenant extension by its ARM-like resource ID.
func (r *PerTenantExtensionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tenantID, appFamily, envName, appID, err := ParsePerTenantExtensionID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("aad_tenant_id"), tenantID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("application_family"), appFamily)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("environment_name"), envName)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("app_id"), appID)...)
}
