// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceSubnetPool_full(t *testing.T) {
	dataSourceName := "data.oxide_subnet_pool.test"

	subnet := nextSubnetCIDR(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSubnetPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testDataSourceSubnetPoolConfig(subnet),
				Check:  checkDataSourceSubnetPool(dataSourceName, subnet),
			},
		},
	})
}

func testDataSourceSubnetPoolConfig(subnet string) string {
	return fmt.Sprintf(`
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-ds-subnet-pool"
	description = "a test subnet pool for data source"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_member" "test" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = %q
	min_prefix_length = 26
	max_prefix_length = 28
}

data "oxide_subnet_pool" "test" {
	name       = oxide_subnet_pool.test.name
	depends_on = [oxide_subnet_pool_member.test]
	timeouts = {
		read = "1m"
	}
}
`, subnet)
}

func checkDataSourceSubnetPool(dataName, subnet string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttr(dataName, "name", "terraform-acc-ds-subnet-pool"),
		resource.TestCheckResourceAttr(
			dataName,
			"description",
			"a test subnet pool for data source",
		),
		resource.TestCheckResourceAttr(dataName, "ip_version", "v4"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "members.#", "1"),
		resource.TestCheckResourceAttr(dataName, "members.0.subnet", subnet),
		resource.TestCheckResourceAttr(dataName, "members.0.min_prefix_length", "26"),
		resource.TestCheckResourceAttr(dataName, "members.0.max_prefix_length", "28"),
	}...)
}
