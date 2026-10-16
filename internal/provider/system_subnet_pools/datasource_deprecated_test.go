// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpools_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccDataSourceSubnetPools_full(t *testing.T) {
	const dataSourceName = "data.oxide_subnet_pools.test"
	poolName := sharedtest.NewResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDeprecatedDataSourceConfig(poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "timeouts.read", "1m"),
					checkSubnetPoolInCollection(dataSourceName, poolName),
				),
			},
		},
	})
}

func testDeprecatedDataSourceConfig(poolName string) string {
	return fmt.Sprintf(`
resource "oxide_subnet_pool" "test" {
  name        = %q
  description = "a test subnet pool for the collection data source"
  ip_version  = "v4"
}

data "oxide_subnet_pools" "test" {
  depends_on = [oxide_subnet_pool.test]
  timeouts = {
    read = "1m"
  }
}
`, poolName)
}
