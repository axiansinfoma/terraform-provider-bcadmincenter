// Copyright (c) 2025 Axians Infoma GmbH
// SPDX-License-Identifier: MPL-2.0

package services_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	authorizedentraapps "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/authorized_entra_apps"
	environmentapps "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/environment_apps"
	environmentsupportcontact "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/environment_support_contact"
	"github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/environments"
	notificationrecipients "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/notification_recipients"
	pertenantextensions "github.com/axiansinfoma/terraform-provider-bcadmincenter/internal/services/per_tenant_extensions"
)

// TestImportExamplesParse checks that every ID shipped in examples/resources/*/import.sh
// is actually accepted by the matching ImportState parser. Documentation that shows an
// unparseable import ID is worse than none: the user only finds out after the command
// fails.
func TestImportExamplesParse(t *testing.T) {
	parsers := map[string]func(string) error{
		"bcadmincenter_authorized_entra_app": func(id string) error {
			_, _, err := authorizedentraapps.ParseAuthorizedEntraAppID(id)
			return err
		},
		"bcadmincenter_environment": func(id string) error {
			_, _, _, err := environments.ParseEnvironmentID(id)
			return err
		},
		"bcadmincenter_environment_update_schedule": func(id string) error {
			_, _, _, err := environments.ParseUpdateScheduleID(id)
			return err
		},
		"bcadmincenter_environment_app": func(id string) error {
			_, _, _, _, err := environmentapps.ParseEnvironmentAppID(id)
			return err
		},
		"bcadmincenter_environment_support_contact": func(id string) error {
			_, _, _, err := environmentsupportcontact.ParseEnvironmentSupportContactID(id)
			return err
		},
		"bcadmincenter_notification_recipient": func(id string) error {
			_, _, err := notificationrecipients.ParseNotificationRecipientID(id)
			return err
		},
		"bcadmincenter_per_tenant_extension": func(id string) error {
			_, _, _, _, err := pertenantextensions.ParsePerTenantExtensionID(id)
			return err
		},
	}

	idRE := regexp.MustCompile(`terraform import\s+\S+\s+"([^"]+)"`)

	for name, parse := range parsers {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("..", "..", "examples", "resources", name, "import.sh")
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("every resource should ship an import.sh: %v", err)
			}
			match := idRE.FindSubmatch(content)
			if match == nil {
				t.Fatalf("%s does not contain a `terraform import <addr> \"<id>\"` line:\n%s", path, content)
			}
			id := string(match[1])
			if err := parse(id); err != nil {
				t.Errorf("the documented import ID does not parse: %v\n  id: %s", err, id)
			}
			if !strings.HasPrefix(id, "/tenants/") {
				t.Errorf("import ID should start with /tenants/, got %s", id)
			}
		})
	}
}
