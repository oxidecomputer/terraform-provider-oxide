// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolsilolink_test

import (
	"context"
	"fmt"
	"strings"
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
	var linkID string
	captureLinkID := func(s *terraform.State) error {
		linkID = s.RootModule().Resources[linkResourceName].Primary.ID
		return nil
	}
	checkLinkID := func(s *terraform.State) error {
		got := s.RootModule().Resources[linkResourceName].Primary.ID
		if got != linkID {
			return fmt.Errorf("expected link ID %q after update, got %q", linkID, got)
		}
		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig(poolName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(linkResourceName, "id"),
					checkLinkIDFormat(linkResourceName, poolResourceName),
					captureLinkID,
					resource.TestCheckResourceAttr(
						linkResourceName,
						"pool",
						poolName,
					),
					resource.TestCheckResourceAttr(linkResourceName, "silo", "test-suite-silo"),
					resource.TestCheckResourceAttr(linkResourceName, "is_default", "false"),
				),
			},
			{
				Config: testResourceConfig(poolName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(linkResourceName, "is_default", "true"),
					checkLinkID,
				),
			},
			{
				ResourceName:      linkResourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"pool",
					"silo",
					"timeouts",
				},
			},
			{
				Config: testPoolOnlyConfig(poolName),
				Check:  testAccResourceDestroy(poolResourceName),
			},
		},
	})
}

func checkLinkIDFormat(linkResourceName, poolResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		link, ok := s.RootModule().Resources[linkResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", linkResourceName)
		}
		pool, ok := s.RootModule().Resources[poolResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", poolResourceName)
		}

		idParts := strings.Split(link.Primary.ID, "/")
		if len(idParts) != 2 || idParts[0] == "" || idParts[1] != pool.Primary.ID {
			return fmt.Errorf(
				"expected link ID in silo_id/subnet_pool_id format, got %q",
				link.Primary.ID,
			)
		}

		return nil
	}
}

func testResourceConfig(poolName string, isDefault bool) string {
	return fmt.Sprintf(`
resource "oxide_system_subnet_pool" "test" {
  name        = %[1]q
  description = "a test subnet pool for system silo link tests"
  ip_version  = "v6"
}

resource "oxide_system_subnet_pool_silo_link" "test" {
  pool       = oxide_system_subnet_pool.test.name
  silo       = "test-suite-silo"
  is_default = %[2]t
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
