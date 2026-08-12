// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippool_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccDataSourceSystemIPPool_full(t *testing.T) {
	dataSourceName := "data.oxide_system_ip_pool.test"
	poolName := sharedtest.NewResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceConfig(poolName),
				Check:  checkDataSource(dataSourceName, poolName),
			},
		},
	})
}

func testDataSourceConfig(poolName string) string {
	return fmt.Sprintf(`
resource "oxide_ip_pool" "test" {
  name        = %q
  description = "a test system IP pool for the data source"
}

data "oxide_system_ip_pool" "test" {
  pool = oxide_ip_pool.test.name
  timeouts = {
    read = "1m"
  }
}
`, poolName)
}

func checkDataSource(dataName, poolName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttr(dataName, "pool", poolName),
		resource.TestCheckResourceAttr(dataName, "name", poolName),
		resource.TestCheckResourceAttr(
			dataName,
			"description",
			"a test system IP pool for the data source",
		),
		resource.TestCheckResourceAttr(dataName, "assignment", "silos"),
		resource.TestCheckResourceAttr(dataName, "ip_version", "v4"),
		resource.TestCheckResourceAttr(dataName, "pool_type", "unicast"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
	}...)
}
