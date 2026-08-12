// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolmember_test

import (
	"context"
	"fmt"
	"net/netip"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccResourceSystemSubnetPoolMember_full(t *testing.T) {
	poolName := sharedtest.NewResourceName()
	memberResourceName := "oxide_system_subnet_pool_member.test"
	subnet1 := sharedtest.NextSubnetCIDR(t)
	subnet2 := sharedtest.NextSubnetCIDR(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig(poolName, subnet1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(memberResourceName, "pool", poolName),
					resource.TestCheckResourceAttr(memberResourceName, "subnet", subnet1),
					resource.TestCheckResourceAttr(
						memberResourceName,
						"id",
						fmt.Sprintf("%s/%s", poolName, subnet1),
					),
				),
			},
			{
				ResourceName:            memberResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
			{
				Config: testResourceConfig(poolName, subnet2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							memberResourceName,
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(memberResourceName, "subnet", subnet2),
					resource.TestCheckResourceAttr(
						memberResourceName,
						"id",
						fmt.Sprintf("%s/%s", poolName, subnet2),
					),
				),
			},
			{
				Config: testResourceConfigWithPoolID(poolName, subnet2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							memberResourceName,
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.TestCheckResourceAttrPair(
					memberResourceName,
					"pool",
					"oxide_subnet_pool.test",
					"id",
				),
			},
		},
	})
}

func TestAccResourceSystemSubnetPoolMember_disappears(t *testing.T) {
	poolName := sharedtest.NewResourceName()
	memberResourceName := "oxide_system_subnet_pool_member.test"
	subnet := sharedtest.NextSubnetCIDR(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig(poolName, subnet),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(memberResourceName, "id"),
					testAccResourceDisappears(memberResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testResourceConfig(poolName, subnet string) string {
	return fmt.Sprintf(`
resource "oxide_subnet_pool" "test" {
	name        = %[1]q
	description = "a test subnet pool for system member tests"
	ip_version  = "v4"
}

resource "oxide_system_subnet_pool_member" "test" {
	pool   = oxide_subnet_pool.test.name
	subnet = %[2]q
}
`, poolName, subnet)
}

func testResourceConfigWithPoolID(poolName, subnet string) string {
	return fmt.Sprintf(`
resource "oxide_subnet_pool" "test" {
	name        = %[1]q
	description = "a test subnet pool for system member tests"
	ip_version  = "v4"
}

resource "oxide_system_subnet_pool_member" "test" {
	pool   = oxide_subnet_pool.test.id
	subnet = %[2]q
}
`, poolName, subnet)
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
		subnet, err := oxide.NewIpNet(rs.Primary.Attributes["subnet"])
		if err != nil {
			return fmt.Errorf("error parsing subnet: %w", err)
		}

		return client.SystemSubnetPoolMemberRemove(
			context.Background(),
			oxide.SystemSubnetPoolMemberRemoveParams{
				Pool: oxide.NameOrId(rs.Primary.Attributes["pool"]),
				Body: &oxide.SubnetPoolMemberRemove{Subnet: subnet},
			},
		)
	}
}

func testAccResourceDestroy(s *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_system_subnet_pool_member" {
			continue
		}

		stateSubnet, err := netip.ParsePrefix(rs.Primary.Attributes["subnet"])
		if err != nil {
			return err
		}
		members, err := client.SystemSubnetPoolMemberListAllPages(
			context.Background(),
			oxide.SystemSubnetPoolMemberListParams{
				Pool: oxide.NameOrId(rs.Primary.Attributes["pool"]),
			},
		)
		if err != nil && shared.Is404(err) {
			continue
		}
		if err != nil {
			return err
		}

		for _, member := range members {
			memberSubnet, err := netip.ParsePrefix(member.Subnet.String())
			if err != nil {
				return err
			}
			if memberSubnet == stateSubnet {
				return fmt.Errorf(
					"system subnet pool member (%s) still exists",
					member.Subnet.String(),
				)
			}
		}
	}

	return nil
}
