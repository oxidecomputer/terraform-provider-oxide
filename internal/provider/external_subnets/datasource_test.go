// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package externalsubnets_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccDataSourceExternalSubnets_full(t *testing.T) {
	const dataSourceName = "data.oxide_external_subnets.test"
	externalSubnetName := sharedtest.NewResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceConfig(t, externalSubnetName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "project", "tf-acc-test"),
					resource.TestCheckResourceAttr(dataSourceName, "timeouts.read", "1m"),
					checkExternalSubnetInCollection(dataSourceName, externalSubnetName),
				),
			},
		},
	})
}

func testDataSourceConfig(t *testing.T, externalSubnetName string) string {
	t.Helper()
	return fmt.Sprintf(`
data "oxide_project" "test" {
  name = "tf-acc-test"
}

data "oxide_silo" "test" {
  name = "test-suite-silo"
}

resource "oxide_subnet_pool" "test" {
  name        = %q
  description = "a subnet pool for the external subnets data source"
  ip_version  = "v4"
}

resource "oxide_subnet_pool_member" "test" {
  subnet_pool_id    = oxide_subnet_pool.test.id
  subnet            = %q
  max_prefix_length = 30
}

resource "oxide_subnet_pool_silo_link" "test" {
  subnet_pool_id = oxide_subnet_pool.test.id
  silo_id        = data.oxide_silo.test.id
  is_default     = false
}

resource "oxide_external_subnet" "test" {
  project_id     = data.oxide_project.test.id
  name           = %q
  description    = "an external subnet returned by the collection data source"
  prefix_len     = 28
  subnet_pool_id = oxide_subnet_pool.test.id
  depends_on     = [oxide_subnet_pool_silo_link.test, oxide_subnet_pool_member.test]
}

data "oxide_external_subnets" "test" {
  project    = data.oxide_project.test.name
  depends_on = [oxide_external_subnet.test]
  timeouts = {
    read = "1m"
  }
}
`, sharedtest.NewResourceName(), sharedtest.NextSubnetCIDR(t), externalSubnetName)
}

func checkExternalSubnetInCollection(
	dataSourceName, externalSubnetName string,
) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source %s not found in state", dataSourceName)
		}

		attributes := resourceState.Primary.Attributes
		count, err := strconv.Atoi(attributes["external_subnets.#"])
		if err != nil {
			return fmt.Errorf("invalid external subnet count: %w", err)
		}

		for i := range count {
			prefix := fmt.Sprintf("external_subnets.%d.", i)
			if attributes[prefix+"name"] != externalSubnetName {
				continue
			}

			for _, attribute := range []string{
				"id",
				"project_id",
				"subnet",
				"subnet_pool_id",
				"subnet_pool_member_id",
				"time_created",
				"time_modified",
			} {
				if attributes[prefix+attribute] == "" {
					return fmt.Errorf(
						"external subnet %q has empty %s",
						externalSubnetName,
						attribute,
					)
				}
			}
			if got := attributes[prefix+"description"]; got != "an external subnet returned by the collection data source" {
				return fmt.Errorf("external subnet %q has description %q", externalSubnetName, got)
			}

			return nil
		}

		return fmt.Errorf("external subnet %q not found in collection", externalSubnetName)
	}
}
