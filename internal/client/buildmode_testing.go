// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

//go:build bcadmincenter_testing

package client

// testingBuild reports whether this binary was built with the `bcadmincenter_testing`
// build tag.
//
// It gates two capabilities that must never exist in a released provider:
//
//   - accepting a static, pre-obtained bearer token in place of Azure AD authentication
//     (Config.AccessToken / BCADMINCENTER_TEST_TOKEN), and
//   - talking to a plaintext http:// base URL, which would put a live Azure AD access
//     token on the wire in the clear.
//
// Because this is a constant, the compiler removes the guarded branches entirely from
// release builds rather than leaving them reachable at runtime.
const testingBuild = true
