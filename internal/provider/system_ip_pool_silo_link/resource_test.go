// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippoolsilolink_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccResourceSystemIPPoolSiloLink_full(t *testing.T) {
	firstPoolName := sharedtest.NewResourceName()
	secondPoolName := sharedtest.NewResourceName()
	linkResourceName := "oxide_system_ip_pool_silo_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig(firstPoolName, secondPoolName, "first", false, "1m"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(linkResourceName, "id"),
					resource.TestCheckResourceAttr(linkResourceName, "is_default", "false"),
					resource.TestCheckResourceAttr(linkResourceName, "pool", firstPoolName),
					resource.TestCheckResourceAttr(linkResourceName, "silo", "test-suite-silo"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.read", "1m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.create", "3m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.update", "4m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.delete", "2m"),
					checkLinkIDComposite(linkResourceName),
				),
			},
			{
				Config: testResourceConfig(firstPoolName, secondPoolName, "first", false, "2m"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							linkResourceName,
							plancheck.ResourceActionUpdate,
						),
					},
				},
				Check: resource.TestCheckResourceAttr(
					linkResourceName,
					"timeouts.read",
					"2m",
				),
			},
			{
				Config: testResourceConfig(
					firstPoolName,
					secondPoolName,
					"first",
					true,
					"2m",
				),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							linkResourceName,
							plancheck.ResourceActionUpdate,
						),
					},
				},
			},
			{
				Config: testResourceConfig(firstPoolName, secondPoolName, "second", false, "2m"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							linkResourceName,
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(linkResourceName, "pool", secondPoolName),
					resource.TestCheckResourceAttr(linkResourceName, "silo", "test-suite-silo"),
					checkLinkIDComposite(linkResourceName),
				),
			},
			{
				ResourceName:            linkResourceName,
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s/test-suite-silo", secondPoolName),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
			{
				ResourceName:            linkResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"pool", "silo", "timeouts"},
			},
			{
				Config: testResourceConfig(firstPoolName, secondPoolName, "second", false, "2m"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							linkResourceName,
							plancheck.ResourceActionUpdate,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceDisappears(linkResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testResourceConfig(
	firstPoolName, secondPoolName, selectedPool string,
	isDefault bool,
	readTimeout string,
) string {
	return fmt.Sprintf(`
resource "oxide_ip_pool" "first" {
	name        = %[1]q
	description = "first system IP pool silo link test pool"
}

resource "oxide_ip_pool" "second" {
	name        = %[2]q
	description = "second system IP pool silo link test pool"
}

resource "oxide_system_ip_pool_silo_link" "test" {
	pool       = oxide_ip_pool.%[3]s.name
	silo       = "test-suite-silo"
	is_default = %[4]t
	timeouts = {
		read   = %[5]q
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`, firstPoolName, secondPoolName, selectedPool, isDefault, readTimeout)
}

func checkLinkIDComposite(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		parts := strings.Split(rs.Primary.ID, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("expected id in IP_POOL_ID/SILO_ID format, got %q", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourceDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}

		return client.SystemIpPoolSiloUnlink(
			context.Background(),
			oxide.SystemIpPoolSiloUnlinkParams{
				Pool: oxide.NameOrId(rs.Primary.Attributes["pool"]),
				Silo: oxide.NameOrId(rs.Primary.Attributes["silo"]),
			},
		)
	}
}
