// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpools_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccDataSourceSystemSubnetPools_full(t *testing.T) {
	const dataSourceName = "data.oxide_system_subnet_pools.test"
	poolName := sharedtest.NewResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceConfig(poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "timeouts.read", "1m"),
					checkSubnetPoolInCollection(dataSourceName, poolName),
				),
			},
		},
	})
}

func testDataSourceConfig(poolName string) string {
	return fmt.Sprintf(`
resource "oxide_subnet_pool" "test" {
  name        = %q
  description = "a test subnet pool for the collection data source"
  ip_version  = "v4"
}

data "oxide_system_subnet_pools" "test" {
  depends_on = [oxide_subnet_pool.test]
  timeouts = {
    read = "1m"
  }
}
`, poolName)
}

func checkSubnetPoolInCollection(dataSourceName, poolName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source %s not found in state", dataSourceName)
		}

		attributes := resourceState.Primary.Attributes
		count, err := strconv.Atoi(attributes["subnet_pools.#"])
		if err != nil {
			return fmt.Errorf("invalid subnet pool count: %w", err)
		}

		for i := range count {
			prefix := fmt.Sprintf("subnet_pools.%d.", i)
			if attributes[prefix+"name"] != poolName {
				continue
			}

			for _, attribute := range []string{"id", "time_created", "time_modified"} {
				if attributes[prefix+attribute] == "" {
					return fmt.Errorf("subnet pool %q has empty %s", poolName, attribute)
				}
			}
			if got := attributes[prefix+"description"]; got != "a test subnet pool for the collection data source" {
				return fmt.Errorf("subnet pool %q has description %q", poolName, got)
			}
			if got := attributes[prefix+"ip_version"]; got != "v4" {
				return fmt.Errorf("subnet pool %q has IP version %q", poolName, got)
			}

			return nil
		}

		return fmt.Errorf("subnet pool %q not found in collection", poolName)
	}
}
