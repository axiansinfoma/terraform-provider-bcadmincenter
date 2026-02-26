#!/bin/bash
# Copyright Axians Infoma GmbH 2025, 2026, 0
# SPDX-License-Identifier: MPL-2.0


# Local Provider Testing Helper Script
# This script helps set up local testing for the BC Admin Center Terraform provider

set -e

VERSION="1.0.0"
GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)
PROVIDER_NAME="terraform-provider-bcadmincenter"
TERRAFORMRC="$HOME/.terraformrc"

echo "🔨 Building provider..."
go build -o "$PROVIDER_NAME"

if [ ! -f "$PROVIDER_NAME" ]; then
    echo "❌ Build failed - binary not found"
    exit 1
fi

echo "✅ Provider built successfully"
echo ""

# Check if .terraformrc exists
if [ -f "$TERRAFORMRC" ]; then
    echo "⚠️  Found existing $TERRAFORMRC"
    echo "   Please review and manually add the dev override configuration."
    echo ""
fi

echo "📝 To enable local testing, add this to $TERRAFORMRC:"
echo ""
echo "provider_installation {"
echo "  dev_overrides {"
echo "    \"axiansinfoma/bcadmincenter\" = \"$(pwd)\""
echo "  }"
echo "  direct {}"
echo "}"
echo ""

read -p "Would you like to automatically add this configuration? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    # Backup existing terraformrc if it exists
    if [ -f "$TERRAFORMRC" ]; then
        cp "$TERRAFORMRC" "${TERRAFORMRC}.backup.$(date +%Y%m%d%H%M%S)"
        echo "✅ Backed up existing configuration"
    fi
    
    # Add dev override configuration
    cat >> "$TERRAFORMRC" << EOF

# BC Admin Center Provider - Development Override
provider_installation {
  dev_overrides {
    "axiansinfoma/bcadmincenter" = "$(pwd)"
  }
  direct {}
}
EOF
    
    echo "✅ Configuration added to $TERRAFORMRC"
fi

echo ""
echo "🎯 Next Steps:"
echo "   1. Set your Azure credentials:"
echo "      export AZURE_CLIENT_ID=\"your-client-id\""
echo "      export AZURE_CLIENT_SECRET=\"your-client-secret\""
echo "      export AZURE_TENANT_ID=\"your-tenant-id\""
echo ""
echo "      Or log in via Azure CLI: az login"
echo "   2. Create a test Terraform configuration (see examples/)"
echo ""
echo "   3. Run Terraform commands:"
echo "      terraform init"
echo "      terraform plan"
echo ""
echo "   4. After making changes, just rebuild and test:"
echo "      go build -o $PROVIDER_NAME && terraform plan"
echo ""
echo "✨ Happy testing!"
