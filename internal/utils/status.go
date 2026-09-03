// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package utils

import "strings"

// StatusIs reports whether an Admin Center status value matches any of the candidates,
// ignoring case.
//
// Status values must never be compared with ==. The live API returns operation statuses
// and app states capitalised on some endpoints and lower-cased on others, so an exact
// comparison silently fails to match — and where the caller treats an unrecognised
// status as fatal, that turns a successful long-running operation into a failed apply.
//
// Passing several candidates also covers spelling drift: the same operations API is
// documented as returning both "canceled" and "cancelled", so callers accept both.
func StatusIs(status string, candidates ...string) bool {
	for _, candidate := range candidates {
		if strings.EqualFold(status, candidate) {
			return true
		}
	}
	return false
}
