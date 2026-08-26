// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolsilolink_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccResourceSystemSubnetPoolSiloLink_full(t *testing.T) {
	poolName := sharedtest.NewResourceName()
	poolResourceName := "oxide_system_subnet_pool.test"
	linkResourceName := "oxide_system_subnet_pool_silo_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig(poolName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(linkResourceName, "id"),
					resource.TestCheckResourceAttr(
						linkResourceName,
						"is_default",
						"false",
					),
					resource.TestCheckResourceAttrPair(
						linkResourceName,
						"subnet_pool_id",
						poolResourceName,
						"id",
					),
					resource.TestCheckResourceAttrPair(
						linkResourceName,
						"silo_id",
						"data.oxide_silo.test",
						"id",
					),
				),
			},
			{
				Config: testResourceConfig(poolName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(linkResourceName, "is_default", "true"),
				),
			},
			{
				ResourceName:            linkResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					link, ok := s.RootModule().Resources[linkResourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", linkResourceName)
					}
					return fmt.Sprintf(
						"%s/%s",
						link.Primary.Attributes["subnet_pool_id"],
						link.Primary.Attributes["silo_id"],
					), nil
				},
			},
			{
				Config: testPoolOnlyConfig(poolName),
				Check:  testAccResourceDestroy(poolResourceName),
			},
		},
	})
}

func testResourceConfig(poolName string, isDefault bool) string {
	return fmt.Sprintf(`
data "oxide_silo" "test" {
  name = "test-suite-silo"
}

resource "oxide_system_subnet_pool" "test" {
  name        = %[1]q
  description = "a test subnet pool for system silo link tests"
  ip_version  = "v6"
}

resource "oxide_system_subnet_pool_silo_link" "test" {
  subnet_pool_id = oxide_system_subnet_pool.test.id
  silo_id        = data.oxide_silo.test.id
  is_default     = %[2]t
}
`, poolName, isDefault)
}

func testPoolOnlyConfig(poolName string) string {
	return fmt.Sprintf(`
resource "oxide_system_subnet_pool" "test" {
  name        = %q
  description = "a test subnet pool for system silo link tests"
  ip_version  = "v6"
}
`, poolName)
}

func testAccResourceDestroy(poolResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[poolResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", poolResourceName)
		}

		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}

		links, err := client.SystemSubnetPoolSiloListAllPages(
			context.Background(),
			oxide.SystemSubnetPoolSiloListParams{Pool: oxide.NameOrId(rs.Primary.ID)},
		)
		if err != nil {
			return err
		}
		if len(links) != 0 {
			return fmt.Errorf(
				"expected no silo links for pool %s, got %d",
				rs.Primary.ID,
				len(links),
			)
		}

		return nil
	}
}
