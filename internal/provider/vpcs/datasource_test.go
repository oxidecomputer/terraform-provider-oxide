// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcs_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccDataSourceVPCs_full(t *testing.T) {
	const dataSourceName = "data.oxide_vpcs.test"
	vpcName := sharedtest.NewResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceConfig(vpcName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "project", "tf-acc-test"),
					resource.TestCheckResourceAttr(dataSourceName, "timeouts.read", "1m"),
					checkVPCInCollection(dataSourceName, vpcName),
				),
			},
		},
	})
}

func testDataSourceConfig(vpcName string) string {
	return fmt.Sprintf(`
data "oxide_project" "test" {
  name = "tf-acc-test"
}

resource "oxide_vpc" "test" {
  project_id  = data.oxide_project.test.id
  name        = %[1]q
  description = "a test VPC for the collection data source"
  dns_name    = %[1]q
}

data "oxide_vpcs" "test" {
  project    = "tf-acc-test"
  depends_on = [oxide_vpc.test]
  timeouts = {
    read = "1m"
  }
}
`, vpcName)
}

func checkVPCInCollection(dataSourceName, vpcName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source %s not found in state", dataSourceName)
		}

		attributes := resourceState.Primary.Attributes
		count, err := strconv.Atoi(attributes["vpcs.#"])
		if err != nil {
			return fmt.Errorf("invalid VPC count: %w", err)
		}

		for i := range count {
			prefix := fmt.Sprintf("vpcs.%d.", i)
			if attributes[prefix+"name"] != vpcName {
				continue
			}

			for _, attribute := range []string{
				"dns_name", "id", "ipv6_prefix", "project_id", "system_router_id", "time_created", "time_modified",
			} {
				if attributes[prefix+attribute] == "" {
					return fmt.Errorf("VPC %q has empty %s", vpcName, attribute)
				}
			}
			if got := attributes[prefix+"description"]; got != "a test VPC for the collection data source" {
				return fmt.Errorf("VPC %q has description %q", vpcName, got)
			}

			return nil
		}

		return fmt.Errorf("VPC %q not found in collection", vpcName)
	}
}
