// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"errors"
	"strings"
	"testing"
)

type failingReader struct{}

func (f failingReader) Read(_ []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestReadResponseBody(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		got := ReadResponseBody(strings.NewReader("test body"))
		if got != "test body" {
			t.Errorf("ReadResponseBody() = %q, want %q", got, "test body")
		}
	})

	t.Run("read error", func(t *testing.T) {
		got := ReadResponseBody(failingReader{})
		if !strings.Contains(got, "failed to read response body") {
			t.Errorf("ReadResponseBody() = %q, want message containing %q", got, "failed to read response body")
		}
	})
}
