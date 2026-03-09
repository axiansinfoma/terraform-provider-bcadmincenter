# environment_apps

Package `environment_apps` manages the install/update/uninstall lifecycle for apps in a Business Central environment via the `bcadmincenter_environment_app` Terraform resource.

## Overview

The package exposes:

- **`NewEnvironmentAppResource()`** – registers the `bcadmincenter_environment_app` Terraform resource.
- **`Service`** – low-level API client wrapping the Business Central Admin Center app management endpoints.
- **`BuildEnvironmentAppID` / `ParseEnvironmentAppID`** – ARM-like resource ID helpers.

## Resource Behaviour

| Operation | API call | Notes |
|-----------|----------|-------|
| Create | `POST .../apps/{appId}/install` | Async; polls until succeeded |
| Read | `GET .../apps` + filter by ID | Removes from state if not found |
| Update | `POST .../apps/{appId}/update` | Triggered when `version` changes |
| Delete | `POST .../apps/{appId}/uninstall` | Async; polls until succeeded |
| Import | Parse ARM ID | Sets identity attributes; Read populates the rest |

## Terminal Failure States

When `status` is `"installFailed"` or `"updateFailed"`, the resource is marked for replacement on the next `terraform plan`. This allows Terraform to recover automatically without manual state manipulation.
