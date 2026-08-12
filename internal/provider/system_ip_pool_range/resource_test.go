// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippoolrange_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccResourceSystemIPPoolRange_full(t *testing.T) {
	poolName := sharedtest.NewResourceName()
	resourceName := "oxide_system_ip_pool_range.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig(poolName, "172.20.30.10", "172.20.30.19"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "pool_id"),
					resource.TestCheckResourceAttr(resourceName, "pool", poolName),
					resource.TestCheckResourceAttr(resourceName, "first", "172.20.30.10"),
					resource.TestCheckResourceAttr(resourceName, "last", "172.20.30.19"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rangeResource, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return fmt.Sprintf(
						"%s/%s",
						rangeResource.Primary.Attributes["pool"],
						rangeResource.Primary.ID,
					), nil
				},
			},
			{
				Config: testResourceWithUpdatedPoolConfig(
					poolName,
					"172.20.30.10",
					"172.20.30.19",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "first", "172.20.30.10"),
					resource.TestCheckResourceAttr(resourceName, "last", "172.20.30.19"),
				),
			},
			{
				Config: testResourceConfig(poolName, "172.20.30.20", "172.20.30.29"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "first", "172.20.30.20"),
					resource.TestCheckResourceAttr(resourceName, "last", "172.20.30.29"),
				),
			},
		},
	})
}

func testResourceWithUpdatedPoolConfig(poolName, first, last string) string {
	return fmt.Sprintf(`
resource "oxide_ip_pool" "test" {
	name        = %q
	description = "updated system IP pool range acceptance test"
}

resource "oxide_system_ip_pool_range" "test" {
	pool  = oxide_ip_pool.test.name
	first = %q
	last  = %q
}
`, poolName, first, last)
}

func testResourceConfig(poolName, first, last string) string {
	return fmt.Sprintf(`
resource "oxide_ip_pool" "test" {
	name        = %q
	description = "system IP pool range acceptance test"
}

resource "oxide_system_ip_pool_range" "test" {
	pool  = oxide_ip_pool.test.name
	first = %q
	last  = %q
}
`, poolName, first, last)
}

func testAccResourceDestroy(state *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, resourceState := range state.RootModule().Resources {
		if resourceState.Type != "oxide_system_ip_pool_range" {
			continue
		}

		ranges, err := client.SystemIpPoolRangeListAllPages(
			context.Background(),
			oxide.SystemIpPoolRangeListParams{
				Pool: oxide.NameOrId(resourceState.Primary.Attributes["pool"]),
			},
		)
		if err != nil && shared.Is404(err) {
			continue
		}
		if err != nil {
			return err
		}

		for _, ipRange := range ranges {
			if ipRange.Id == resourceState.Primary.ID {
				return fmt.Errorf("system_ip_pool_range (%s) still exists", ipRange.Id)
			}
		}
	}

	return nil
}
