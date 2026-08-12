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
	poolResourceName := "oxide_subnet_pool.test"
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
				Config: testResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(linkResourceName, "id"),
					captureLinkID,
					resource.TestCheckResourceAttr(
						linkResourceName,
						"pool",
						"terraform-acc-system-subnet-pool-silo-link",
					),
					resource.TestCheckResourceAttr(linkResourceName, "silo", "test-suite-silo"),
					resource.TestCheckResourceAttr(linkResourceName, "is_default", "false"),
				),
			},
			{
				Config: testResourceUpdateConfig,
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
				Config: testPoolOnlyConfig,
				Check:  testAccResourceDestroy(poolResourceName),
			},
		},
	})
}

var testResourceConfig = `
resource "oxide_subnet_pool" "test" {
  name        = "terraform-acc-system-subnet-pool-silo-link"
  description = "a test subnet pool for system silo link tests"
  ip_version  = "v6"
}

resource "oxide_system_subnet_pool_silo_link" "test" {
  pool = oxide_subnet_pool.test.name
  silo = "test-suite-silo"
}
`

var testResourceUpdateConfig = testResourceConfig[:len(testResourceConfig)-2] + `
  is_default = true
}
`

var testPoolOnlyConfig = `
resource "oxide_subnet_pool" "test" {
  name        = "terraform-acc-system-subnet-pool-silo-link"
  description = "a test subnet pool for system silo link tests"
  ip_version  = "v6"
}
`

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
