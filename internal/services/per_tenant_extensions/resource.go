// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package pertenantextensions

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/client"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                     = &PerTenantExtensionResource{}
	_ resource.ResourceWithConfigure        = &PerTenantExtensionResource{}
	_ resource.ResourceWithImportState      = &PerTenantExtensionResource{}
	_ resource.ResourceWithConfigValidators = &PerTenantExtensionResource{}
)

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
		return readExtensionFile(data.FilePath.ValueString())
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

// readExtensionFile loads a .app package, checking what it is and how big it is before
// reading it into memory.
//
// os.ReadFile alone would slurp the whole path first and only then hit the 50 MB check in
// buildPteInstallForm, so pointing file_path at a huge file, /dev/zero or a FIFO would
// exhaust memory or hang instead of returning the limit error.
func readExtensionFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("reading extension package %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("extension package %q is not a regular file", path)
	}
	if info.Size() > MaxExtensionFileSize {
		return nil, fmt.Errorf("extension package %q is %d bytes, which exceeds the %d byte limit enforced by the pteInstall endpoint",
			path, info.Size(), MaxExtensionFileSize)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading extension package %q: %w", path, err)
	}
	return data, nil
}

// verifyFileChecksum checks the uploaded bytes against the configured file_sha256.
//
// Nothing used to compute or compare this value: it existed purely so that editing it
// produced a diff. Two consequences, both silent. Rebuilding a package in place without
// touching the hash meant `terraform plan` reported no changes while the environment
// stayed on the old version. And filesha256() is evaluated at plan time while the file is
// read at apply time, so a CI job that rebuilds the artifact in between — or a saved plan
// applied later — uploaded bytes that did not match the recorded hash.
func verifyFileChecksum(data *PerTenantExtensionResourceModel, content []byte) error {
	if !hasValue(data.FileSHA256) {
		return nil
	}

	want := strings.ToLower(strings.TrimSpace(data.FileSHA256.ValueString()))
	got := fmt.Sprintf("%x", sha256.Sum256(content))
	if got == want {
		return nil
	}

	return fmt.Errorf("file_sha256 does not match the extension package: configured %q, actual %q. "+
		"Set file_sha256 = filesha256(<path>) so it always tracks the package, and make sure the "+
		"package is not rebuilt between plan and apply", want, got)
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

// clearComputedFileName drops a file_name that was derived on a previous apply rather than
// set by the user.
//
// file_name is Optional + Computed, so the plan carries the previously computed value even
// when the config never mentioned it. resolveFileName prefers file_name, so without this
// an update that renames the package (MyExt_1.0.app -> MyExt_2.0.app) kept uploading under
// the stale name forever. Only the config distinguishes the two cases.
func clearComputedFileName(ctx context.Context, config tfsdk.Config, data *PerTenantExtensionResourceModel, diags *diag.Diagnostics) {
	var configured types.String
	diags.Append(config.GetAttribute(ctx, path.Root("file_name"), &configured)...)
	if diags.HasError() {
		return
	}
	if configured.IsNull() && hasValue(data.FilePath) {
		data.FileName = types.StringNull()
	}
}

// ConfigValidators enforces the documented file_path XOR file_content rule at plan time.
//
// It was only checked inside Create and Update, so a config setting both or neither
// planned cleanly and failed during apply — potentially after other resources in the
// graph had already been changed.
func (r *PerTenantExtensionResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.ExactlyOneOf(
			path.MatchRoot("file_path"),
			path.MatchRoot("file_content"),
		),
	}
}

// operationTimeout reads one timeout from this resource's optional `timeouts` block.
func operationTimeout(ctx context.Context, timeouts types.Object, key string) time.Duration {
	return utils.OperationTimeout(ctx, timeouts, key)
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

	// Check the bytes actually being uploaded against the declared hash before anything
	// reaches the API.
	if err := verifyFileChecksum(data, fileBytes); err != nil {
		return nil, err
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

	// The uploaded package must be the same extension this resource already manages.
	// Pointing file_path at a different .app installed the new extension while leaving the
	// old one behind with no Terraform record, and rewrote app_id — a Computed attribute
	// whose planned value is the known previous GUID — failing the apply anyway.
	if operation.AppID != "" && hasValue(data.AppID) && !strings.EqualFold(operation.AppID, data.AppID.ValueString()) {
		return fmt.Errorf("the uploaded package is extension %s, but this resource manages %s. "+
			"Changing which extension a resource manages is not an in-place update: remove this "+
			"resource and declare the other extension separately, or use `terraform state rm` and "+
			"re-import if you intend to take over the new one",
			operation.AppID, data.AppID.ValueString())
	}

	if operation.AppID != "" {
		data.AppID = types.StringValue(operation.AppID)
	} else if !hasValue(data.AppID) {
		// app_id is Computed with UseStateForUnknown, so on Create its planned value is
		// unknown. Leaving it unknown here propagated an unknown value all the way into
		// resp.State.Set, which Terraform rejects — after the package had been uploaded.
		return fmt.Errorf("the pteInstall response did not include an appId, so the installed " +
			"extension cannot be identified; re-run to reconcile, or import the extension once " +
			"the install completes")
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

	// The package has been uploaded and the extension installed, so a failure to read the
	// details back must not lose the resource. Returning without writing state meant a
	// transient 502 on the read-back left an installed extension untracked: the next apply
	// was rejected as already installed, and the user had to import it by hand.
	if err := r.applyOperation(ctx, &data, svc, operation); err != nil {
		resp.Diagnostics.AddError(
			"Failed to read extension details after install",
			err.Error()+"\n\nThe extension has been installed and recorded in state; run "+
				"`terraform plan` to reconcile its details.",
		)
		r.savePartialCreateState(ctx, resp, &data, operation)
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

// savePartialCreateState records an extension that was installed but whose details could
// not be read back. Terraform rejects a state object containing unknown values, so every
// computed attribute is resolved before the write.
func (r *PerTenantExtensionResource) savePartialCreateState(ctx context.Context, resp *resource.CreateResponse, data *PerTenantExtensionResourceModel, operation *AppOperation) {
	if operation != nil {
		if !hasValue(data.AppID) && operation.AppID != "" {
			data.AppID = types.StringValue(operation.AppID)
		}
		if data.LastOperationID.IsUnknown() || data.LastOperationID.IsNull() {
			data.LastOperationID = types.StringValue(operation.ID)
		}
	}

	// Without an app id there is nothing to key the resource on, so state would be
	// meaningless; the diagnostic already tells the user to import.
	if !hasValue(data.AppID) {
		return
	}

	data.ID = types.StringValue(BuildPerTenantExtensionID(
		data.AADTenantID.ValueString(),
		data.ApplicationFamily.ValueString(),
		data.EnvironmentName.ValueString(),
		data.AppID.ValueString(),
	))

	for _, attr := range []*types.String{
		&data.FileName,
		&data.DeploymentSchedule,
		&data.SyncMode,
		&data.LanguageID,
		&data.DisplayName,
		&data.Publisher,
		&data.Version,
		&data.State,
		&data.AppType,
		&data.LastOperationID,
		&data.PendingTargetVersion,
		&data.PendingScheduleKind,
	} {
		if attr.IsUnknown() {
			*attr = types.StringNull()
		}
	}
	for _, attr := range []*types.Bool{
		&data.AcceptIsvEula,
		&data.InstallOrUpdateNeededDependencies,
		&data.DeleteData,
		&data.UninstallDependents,
		&data.UninstallInUpdateWindow,
		&data.CancelScheduledOnDestroy,
	} {
		if attr.IsUnknown() {
			*attr = types.BoolNull()
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
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
		// The environment itself is gone (deleted or renamed out of band), so the extension
		// is too. Erroring here made every plan, apply and destroy fail permanently, with
		// `terraform state rm` the only way out. IsNotFoundError already existed and was
		// used in Delete but not here.
		if IsNotFoundError(err) {
			tflog.Warn(ctx, "Per-tenant extension or its environment no longer exists; removing from state", map[string]interface{}{
				"app_id":           data.AppID.ValueString(),
				"environment_name": data.EnvironmentName.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
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

	// Re-derive file_name from file_path unless the user set it explicitly.
	clearComputedFileName(ctx, req.Config, &data, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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
//
// It returns the versions it could not cancel. These used to be logged with tflog and
// otherwise discarded, so a 403 or 500 on the removal left the destroy reporting success
// while a staged package remained: at the next update window the extension reinstalled
// itself, live and unmanaged — precisely what cancel_scheduled_on_destroy exists to
// prevent. The caller surfaces them as a warning the user can actually see.
func (r *PerTenantExtensionResource) cancelScheduledVersions(ctx context.Context, data *PerTenantExtensionResourceModel, svc *Service) (uncancelled []string) {
	scheduled, err := svc.GetScheduledPteOperationsForApp(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString())
	if err != nil {
		tflog.Warn(ctx, "Failed to list scheduled per-tenant extension versions before destroy", map[string]interface{}{
			"app_id": data.AppID.ValueString(),
			"error":  err.Error(),
		})
		return []string{fmt.Sprintf("could not list scheduled versions: %v", err)}
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
			uncancelled = append(uncancelled, fmt.Sprintf("%s (%s): %v", targetVersion, scheduleKind, err))
			continue
		}

		tflog.Info(ctx, "Cancelled scheduled per-tenant extension version", map[string]interface{}{
			"app_id":         data.AppID.ValueString(),
			"target_version": targetVersion,
			"schedule_kind":  scheduleKind,
		})
	}

	return uncancelled
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
		if uncancelled := r.cancelScheduledVersions(ctx, &data, svc); len(uncancelled) > 0 {
			resp.Diagnostics.AddWarning(
				"Scheduled extension versions could not be cancelled",
				fmt.Sprintf("The extension will be uninstalled, but %d staged version(s) could not be "+
					"removed and may reinstall it at the next deployment window. Remove them in the "+
					"Admin Center:\n  %s", len(uncancelled), strings.Join(uncancelled, "\n  ")),
			)
		}
	}

	app, err := svc.GetApp(ctx, data.ApplicationFamily.ValueString(), data.EnvironmentName.ValueString(), data.AppID.ValueString())
	if err != nil {
		// Already gone: the desired end state is reached, so let the destroy succeed rather
		// than stranding the resource in state.
		if IsNotFoundError(err) {
			return
		}
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

	// Seed the behavioural attributes with their schema defaults. They cannot be read back
	// from the API and Read does not populate them, so a null cancel_scheduled_on_destroy
	// read as false — meaning a destroy straight after an import silently skipped
	// cancelling staged versions and let the extension reinstall itself at the next update
	// window, which is exactly what that flag exists to prevent.
	// accept_isv_eula and file_sha256 are Required, so the user must supply them in config
	// and they are deliberately not seeded here.
	for attr, value := range map[string]bool{
		"install_or_update_needed_dependencies": false,
		"delete_data":                           false,
		"uninstall_dependents":                  false,
		"uninstall_in_update_window":            false,
		"cancel_scheduled_on_destroy":           true,
	} {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(attr), value)...)
	}
}
