// Copyright (c) 2025 Michael Villani
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"fmt"
	"io"
)

// ReadResponseBody reads an HTTP response body and returns a safe, printable string.
func ReadResponseBody(body io.Reader) string {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return fmt.Sprintf("failed to read response body: %v", err)
	}
	return string(bodyBytes)
}
