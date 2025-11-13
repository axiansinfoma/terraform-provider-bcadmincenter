// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package environmentsupportcontact

// SupportContact represents the support contact information for an environment.
type SupportContact struct {
	Name  string `json:"name"`  // The name of the support contact
	Email string `json:"email"` // The email address of the support contact
	URL   string `json:"url"`   // A freeform URL for additional support contact information such as a support website
}
