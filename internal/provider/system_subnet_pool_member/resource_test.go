// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolmember_test

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
					resource.TestCheckResourceAttrPair(
						memberResourceName,
						"subnet_pool_id",
						"oxide_system_subnet_pool.test",
						"id",
					),
					resource.TestCheckResourceAttr(memberResourceName, "subnet", subnet1),
					resource.TestCheckResourceAttr(
						memberResourceName,
						"min_prefix_length",
						"26",
					),
					resource.TestCheckResourceAttr(
						memberResourceName,
						"max_prefix_length",
						"28",
					),
					resource.TestCheckResourceAttrSet(memberResourceName, "id"),
					resource.TestCheckResourceAttrSet(memberResourceName, "time_created"),
					testAccResourceIDMatchesAPI(memberResourceName),
				),
			},
			{
				ResourceName:            memberResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"pool", "timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					member, ok := s.RootModule().Resources[memberResourceName]
					if !ok {
						return "", fmt.Errorf(
							"resource not found: %s",
							memberResourceName,
						)
					}
					pool, ok := s.RootModule().Resources["oxide_system_subnet_pool.test"]
					if !ok {
						return "", fmt.Errorf(
							"resource not found: oxide_system_subnet_pool.test",
						)
					}
					return fmt.Sprintf(
						"%s/%s",
						pool.Primary.ID,
						member.Primary.ID,
					), nil
				},
			},
			{
				Config: testResourceConfigWithDefaults(poolName, subnet2),
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
						"min_prefix_length",
						"24",
					),
					resource.TestCheckResourceAttr(
						memberResourceName,
						"max_prefix_length",
						"32",
					),
					resource.TestCheckResourceAttrSet(memberResourceName, "id"),
					testAccResourceIDMatchesAPI(memberResourceName),
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
					"oxide_system_subnet_pool.test",
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

func TestAccResourceSystemSubnetPoolMember_poolRenamed(t *testing.T) {
	poolName := sharedtest.NewResourceName()
	renamedPoolName := poolName + "-renamed"
	poolResourceName := "oxide_system_subnet_pool.test"
	memberResourceName := "oxide_system_subnet_pool_member.test"
	subnet := sharedtest.NextSubnetCIDR(t)
	var memberID string

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig(poolName, subnet),
				Check: resource.ComposeAggregateTestCheckFunc(
					func(s *terraform.State) error {
						member, ok := s.RootModule().Resources[memberResourceName]
						if !ok {
							return fmt.Errorf("resource not found: %s", memberResourceName)
						}
						memberID = member.Primary.ID
						return nil
					},
					testAccPoolRename(poolResourceName, renamedPoolName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
				Check: func(s *terraform.State) error {
					member, ok := s.RootModule().Resources[memberResourceName]
					if !ok {
						return fmt.Errorf(
							"resource not found after pool rename: %s",
							memberResourceName,
						)
					}
					got := member.Primary.ID
					if got != memberID {
						return fmt.Errorf(
							"expected member ID %q after pool rename, got %q",
							memberID,
							got,
						)
					}
					return nil
				},
			},
			{
				Config: testResourceConfig(renamedPoolName, subnet),
			},
		},
	})
}

func testResourceConfig(poolName, subnet string) string {
	return fmt.Sprintf(`
resource "oxide_system_subnet_pool" "test" {
	name        = %[1]q
	description = "a test subnet pool for system member tests"
	ip_version  = "v4"
}

resource "oxide_system_subnet_pool_member" "test" {
	pool             = oxide_system_subnet_pool.test.name
	subnet           = %[2]q
	min_prefix_length = 26
	max_prefix_length = 28
}
`, poolName, subnet)
}

func testResourceConfigWithDefaults(poolName, subnet string) string {
	return fmt.Sprintf(`
resource "oxide_system_subnet_pool" "test" {
	name        = %[1]q
	description = "a test subnet pool for system member tests"
	ip_version  = "v4"
}

resource "oxide_system_subnet_pool_member" "test" {
	pool   = oxide_system_subnet_pool.test.name
	subnet = %[2]q
}
`, poolName, subnet)
}

func testResourceConfigWithPoolID(poolName, subnet string) string {
	return fmt.Sprintf(`
resource "oxide_system_subnet_pool" "test" {
	name        = %[1]q
	description = "a test subnet pool for system member tests"
	ip_version  = "v4"
}

resource "oxide_system_subnet_pool_member" "test" {
	pool             = oxide_system_subnet_pool.test.id
	subnet           = %[2]q
	min_prefix_length = 26
	max_prefix_length = 28
}
`, poolName, subnet)
}

func testAccPoolRename(resourceName, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}

		_, err = client.SystemSubnetPoolUpdate(
			context.Background(),
			oxide.SystemSubnetPoolUpdateParams{
				Pool: oxide.NameOrId(rs.Primary.ID),
				Body: &oxide.SubnetPoolUpdate{
					Name:        oxide.Name(name),
					Description: rs.Primary.Attributes["description"],
				},
			},
		)
		return err
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
		subnet, err := oxide.NewIpNet(rs.Primary.Attributes["subnet"])
		if err != nil {
			return fmt.Errorf("error parsing subnet: %w", err)
		}

		return client.SystemSubnetPoolMemberRemove(
			context.Background(),
			oxide.SystemSubnetPoolMemberRemoveParams{
				Pool: oxide.NameOrId(rs.Primary.Attributes["subnet_pool_id"]),
				Body: &oxide.SubnetPoolMemberRemove{Subnet: subnet},
			},
		)
	}
}

func testAccResourceIDMatchesAPI(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}
		members, err := client.SystemSubnetPoolMemberListAllPages(
			context.Background(),
			oxide.SystemSubnetPoolMemberListParams{
				Pool: oxide.NameOrId(rs.Primary.Attributes["subnet_pool_id"]),
			},
		)
		if err != nil {
			return err
		}

		for _, member := range members {
			if member.Subnet.String() != rs.Primary.Attributes["subnet"] {
				continue
			}
			if rs.Primary.Attributes["id"] != member.Id {
				return fmt.Errorf(
					"expected ID %q, got %q",
					member.Id,
					rs.Primary.Attributes["id"],
				)
			}
			return nil
		}

		return fmt.Errorf("subnet pool member not found")
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

		members, err := client.SystemSubnetPoolMemberListAllPages(
			context.Background(),
			oxide.SystemSubnetPoolMemberListParams{
				Pool: oxide.NameOrId(rs.Primary.Attributes["subnet_pool_id"]),
			},
		)
		if errors.Is(err, oxide.ErrHTTP404) {
			continue
		}
		if err != nil {
			return err
		}

		for _, member := range members {
			if member.Subnet.String() == rs.Primary.Attributes["subnet"] {
				return fmt.Errorf(
					"system subnet pool member (%s) still exists",
					member.Subnet.String(),
				)
			}
		}
	}

	return nil
}
