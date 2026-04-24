// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

// examplesRoot is the path to examples/ relative to this test file.
const examplesRoot = "../../examples"

// skipExamples lists example .tf files (keyed by path relative to
// examplesRoot) that are known to fail. Remove entries as examples are
// fixed.
var skipExamples = map[string]bool{
	"data-sources/oxide_floating_ip/data-source.tf":           true,
	"data-sources/oxide_image/data-source.tf":                 true,
	"data-sources/oxide_instance_external_ips/data-source.tf": true,
	"data-sources/oxide_silo/data-source.tf":                  true,
	"data-sources/oxide_vpc_internet_gateway/data-source.tf":  true,
	"provider/provider.tf":                                    true,
	"provider/provider-auth-config.tf":                        true,
	"resources/oxide_disk/resource.tf":                        true,
	"resources/oxide_image/resource.tf":                       true,
	"resources/oxide_instance/resource-external-ips.tf":       true,
	"resources/oxide_ip_pool_silo_link/resource.tf":           true,
	"resources/oxide_silo_saml_identity_provider/resource.tf": true,
	"resources/oxide_silo/resource.tf":                        true,
	"resources/oxide_switch_port_settings/resource.tf":        true,
	"resources/oxide_vpc_firewall_rules/resource.tf":          true,
	"resources/oxide_vpc_router_route/resource.tf":            true,
}

// TestAcc_Examples walks every .tf file under examples/ and runs it as an
// acceptance test.
func TestAcc_Examples(t *testing.T) {
	err := filepath.WalkDir(examplesRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".tf") {
			return nil
		}

		relPath, err := filepath.Rel(examplesRoot, path)
		if err != nil {
			return err
		}

		t.Run(relPath, func(t *testing.T) {
			if os.Getenv("SKIP_KNOWN_FAILING") != "" && skipExamples[relPath] {
				t.Skip()
			}

			// Subtests run serially: several examples attach the shared
			// seeded disk, and a disk can only be attached to one
			// instance at a time.
			hcl, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}

			resource.Test(t, resource.TestCase{
				PreCheck:                 func() { sharedtest.PreCheck(t) },
				ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
				Steps: []resource.TestStep{
					{Config: string(hcl)},
				},
			})
		})

		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", examplesRoot, err)
	}
}
