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
	"github.com/stretchr/testify/require"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

const examplesRoot = "../../examples"

// knownFailingExample lists example .tf files known to fail acceptance tests. Try not to add more
// files to this list.
//
// TODO: Fix failing examples and deprecate this escape hatch.
var knownFailingExample = map[string]bool{
	"data-sources/oxide_floating_ip/data-source.tf":           true,
	"data-sources/oxide_image/data-source.tf":                 true,
	"data-sources/oxide_instance_external_ips/data-source.tf": true,
	"data-sources/oxide_silo/data-source.tf":                  true,
	"functions/credentials/function.tf":                       true,
	"provider/provider.tf":                                    true,
	"provider/provider-auth-config.tf":                        true,
	"resources/oxide_disk/resource.tf":                        true,
	"resources/oxide_image/resource.tf":                       true,
	"resources/oxide_instance/resource-external-ips.tf":       true,
	"resources/oxide_ip_pool_silo_link/resource.tf":           true,
	"resources/oxide_silo_saml_identity_provider/resource.tf": true,
	"resources/oxide_silo/resource.tf":                        true,
	"resources/oxide_subnet_pool_silo_link/resource.tf":       true,
	"resources/oxide_vpc_firewall_rules/resource.tf":          true,
	"resources/oxide_vpc_router_route/resource.tf":            true,
}

// TestAcc_Examples lists .tf files within examples/ and runs each as an acceptance test. Because
// this is a generic test function, we only assert that each example runs without error, since we
// don't currently have per-example context to write more detailed expectations. Run
// `scripts/example-test-setup.sh` before running to ensure that the shared resources assumed by the
// examples exist.
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
			if os.Getenv("SKIP_KNOWN_FAILING") != "" && knownFailingExample[relPath] {
				t.Skip()
			}

			// Run subtests in series. Multiple tests may consume the same shared resources from
			// `scripts/example-test-setup.sh`, and we don't want them to fight.
			hcl, err := os.ReadFile(path)
			require.NoError(t, err)

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
	require.NoError(t, err)
}
