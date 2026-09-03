// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

//go:build !bcadmincenter_testing

package client

// testingBuild is false in release builds. See buildmode_testing.go for what it gates.
const testingBuild = false
