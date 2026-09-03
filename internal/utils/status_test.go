// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package utils

import "testing"

func TestStatusIs(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		candidates []string
		want       bool
	}{
		{"exact match", "succeeded", []string{"succeeded"}, true},
		// The live API returns statuses TitleCase on some endpoints and lower-cased on
		// others; an == comparison silently fails to match the TitleCase form.
		{"differing case matches", "Succeeded", []string{"succeeded"}, true},
		{"upper case matches", "SUCCEEDED", []string{"succeeded"}, true},
		{"first of several candidates", "canceled", []string{"canceled", "cancelled"}, true},
		{"second of several candidates", "cancelled", []string{"canceled", "cancelled"}, true},
		{"no match", "running", []string{"succeeded", "failed"}, false},
		{"empty status", "", []string{"succeeded"}, false},
		{"no candidates", "succeeded", nil, false},
		{"empty candidate does not match a set status", "succeeded", []string{""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StatusIs(tt.status, tt.candidates...); got != tt.want {
				t.Errorf("StatusIs(%q, %q) = %v, want %v", tt.status, tt.candidates, got, tt.want)
			}
		})
	}
}
