// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpool_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccResourceSystemSubnetPool_full(t *testing.T) {
	poolName := sharedtest.NewResourceName()
	resourceName := "oxide_system_subnet_pool.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccSystemResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testSystemResourceConfig(poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						poolName,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"description",
						"a test subnet pool",
					),
					resource.TestCheckResourceAttr(resourceName, "ip_version", "v4"),
					resource.TestCheckResourceAttrSet(resourceName, "time_created"),
					resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
				),
			},
			{
				Config: testSystemResourceUpdateConfig(poolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						poolName+"-new",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"description",
						"an updated subnet pool",
					),
					resource.TestCheckResourceAttr(resourceName, "ip_version", "v4"),
				),
			},
			{
				Config: testSystemResourceReplaceConfig(poolName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							resourceName,
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.TestCheckResourceAttr(
					resourceName,
					"ip_version",
					"v6",
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

func testSystemResourceConfig(poolName string) string {
	return fmt.Sprintf(`
resource "oxide_system_subnet_pool" "test" {
  name        = %q
  description = "a test subnet pool"
  ip_version  = "v4"
}
`, poolName)
}

func testSystemResourceUpdateConfig(poolName string) string {
	return fmt.Sprintf(`
resource "oxide_system_subnet_pool" "test" {
  name        = %q
  description = "an updated subnet pool"
  ip_version  = "v4"
}
`, poolName+"-new")
}

func testSystemResourceReplaceConfig(poolName string) string {
	return fmt.Sprintf(`
resource "oxide_system_subnet_pool" "test" {
  name        = %q
  description = "an updated subnet pool"
  ip_version  = "v6"
}
`, poolName+"-new")
}

func testAccSystemResourceDestroy(s *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_system_subnet_pool" {
			continue
		}

		res, err := client.SystemSubnetPoolView(
			context.Background(),
			oxide.SystemSubnetPoolViewParams{
				Pool: oxide.NameOrId(rs.Primary.ID),
			},
		)
		if errors.Is(err, oxide.ErrHTTP404) {
			continue
		}
		if err == nil {
			return fmt.Errorf("system_subnet_pool (%v) still exists", res.Name)
		}
		return err
	}

	return nil
}
