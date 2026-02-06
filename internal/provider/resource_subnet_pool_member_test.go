// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

func TestAccResourceSubnetPoolMember_full(t *testing.T) {
	poolResourceName := "oxide_subnet_pool.test"
	memberResourceName := "oxide_subnet_pool_member.test"
	member2ResourceName := "oxide_subnet_pool_member.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSubnetPoolMemberDestroy,
		Steps: []resource.TestStep{
			// Create pool and one member
			{
				Config: testResourceSubnetPoolMemberConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(poolResourceName, "id"),
					resource.TestCheckResourceAttrSet(memberResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						memberResourceName,
						"subnet_pool_id",
						poolResourceName,
						"id",
					),
					resource.TestCheckResourceAttr(memberResourceName, "subnet", "192.0.2.0/24"),
					resource.TestCheckResourceAttr(memberResourceName, "min_prefix_length", "26"),
					resource.TestCheckResourceAttr(memberResourceName, "max_prefix_length", "28"),
					resource.TestCheckResourceAttrSet(memberResourceName, "time_created"),
				),
			},
			// Import member (format: subnet_pool_id/member_id)
			{
				ResourceName:            memberResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[memberResourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", memberResourceName)
					}
					return fmt.Sprintf(
						"%s/%s",
						rs.Primary.Attributes["subnet_pool_id"],
						rs.Primary.ID,
					), nil
				},
			},
			// Add a second member (with computed min_prefix_length)
			{
				Config: testResourceSubnetPoolMemberAddSecondConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// First member unchanged
					resource.TestCheckResourceAttr(memberResourceName, "subnet", "192.0.2.0/24"),
					resource.TestCheckResourceAttr(memberResourceName, "min_prefix_length", "26"),
					resource.TestCheckResourceAttr(memberResourceName, "max_prefix_length", "28"),
					// Second member has computed min_prefix_length (defaults to subnet prefix = 24)
					resource.TestCheckResourceAttrSet(member2ResourceName, "id"),
					resource.TestCheckResourceAttr(
						member2ResourceName,
						"subnet",
						"198.51.100.0/24",
					),
					resource.TestCheckResourceAttr(member2ResourceName, "min_prefix_length", "24"),
					resource.TestCheckResourceAttr(member2ResourceName, "max_prefix_length", "30"),
				),
			},
		},
	})
}

var testResourceSubnetPoolMemberConfig = `
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-member-test"
	description = "a test subnet pool for member tests"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_member" "test" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = "192.0.2.0/24"
	min_prefix_length = 26
	max_prefix_length = 28
}
`

var testResourceSubnetPoolMemberAddSecondConfig = `
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-member-test"
	description = "a test subnet pool for member tests"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_member" "test" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = "192.0.2.0/24"
	min_prefix_length = 26
	max_prefix_length = 28
}

resource "oxide_subnet_pool_member" "test2" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = "198.51.100.0/24"
	# min_prefix_length omitted - should be computed as 24 (subnet prefix)
	max_prefix_length = 30
}
`

func TestAccResourceSubnetPoolMember_parallel(t *testing.T) {
	poolResourceName := "oxide_subnet_pool.test"
	member1ResourceName := "oxide_subnet_pool_member.m1"
	member2ResourceName := "oxide_subnet_pool_member.m2"
	member3ResourceName := "oxide_subnet_pool_member.m3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSubnetPoolMemberDestroy,
		Steps: []resource.TestStep{
			// Create pool and three members in parallel (no depends_on between members)
			{
				Config: testResourceSubnetPoolMemberParallelConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(poolResourceName, "id"),
					resource.TestCheckResourceAttrSet(member1ResourceName, "id"),
					resource.TestCheckResourceAttrSet(member2ResourceName, "id"),
					resource.TestCheckResourceAttrSet(member3ResourceName, "id"),
					resource.TestCheckResourceAttr(member1ResourceName, "subnet", "203.0.113.0/26"),
					resource.TestCheckResourceAttr(
						member2ResourceName,
						"subnet",
						"203.0.113.64/26",
					),
					resource.TestCheckResourceAttr(
						member3ResourceName,
						"subnet",
						"203.0.113.128/26",
					),
				),
			},
		},
	})
}

var testResourceSubnetPoolMemberParallelConfig = `
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-parallel-test"
	description = "a test subnet pool for parallel member creation"
	ip_version  = "v4"
}

# These three members have no dependencies on each other,
# so Terraform will attempt to create them in parallel.
# Using TEST-NET-3 (203.0.113.0/24) subdivisions to avoid conflicts with _full test.
resource "oxide_subnet_pool_member" "m1" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = "203.0.113.0/26"
	max_prefix_length = 28
}

resource "oxide_subnet_pool_member" "m2" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = "203.0.113.64/26"
	max_prefix_length = 28
}

resource "oxide_subnet_pool_member" "m3" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = "203.0.113.128/26"
	max_prefix_length = 28
}
`

func TestAccResourceSubnetPoolMember_disappears(t *testing.T) {
	poolResourceName := "oxide_subnet_pool.test"
	memberResourceName := "oxide_subnet_pool_member.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSubnetPoolMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceSubnetPoolMemberDisappearsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(poolResourceName, "id"),
					resource.TestCheckResourceAttrSet(memberResourceName, "id"),
					// Delete the member outside of Terraform
					testAccSubnetPoolMemberDisappears(memberResourceName),
				),
				// Expect Terraform to detect the member is gone and plan to recreate
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccSubnetPoolMemberDisappears deletes the member via the API to simulate
// out-of-band deletion, testing that the Read function properly removes it from state.
func testAccSubnetPoolMemberDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := newTestClient()
		if err != nil {
			return err
		}

		subnet, err := oxide.NewIpNet(rs.Primary.Attributes["subnet"])
		if err != nil {
			return fmt.Errorf("error parsing subnet: %w", err)
		}

		params := oxide.SubnetPoolMemberRemoveParams{
			Pool: oxide.NameOrId(rs.Primary.Attributes["subnet_pool_id"]),
			Body: &oxide.SubnetPoolMemberRemove{
				Subnet: subnet,
			},
		}

		return client.SubnetPoolMemberRemove(context.Background(), params)
	}
}

var testResourceSubnetPoolMemberDisappearsConfig = `
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-disappears-test"
	description = "a test subnet pool for disappears test"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_member" "test" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = "198.18.0.0/24"
	max_prefix_length = 28
}
`

func testAccSubnetPoolMemberDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_subnet_pool_member" {
			continue
		}

		ctx := context.Background()
		poolID := rs.Primary.Attributes["subnet_pool_id"]

		// List members and check if ours still exists
		members, err := client.SubnetPoolMemberListAllPages(
			ctx,
			oxide.SubnetPoolMemberListParams{Pool: oxide.NameOrId(poolID)},
		)
		if err != nil && is404(err) {
			// Pool doesn't exist, so member is definitely gone
			continue
		}
		if err != nil {
			return err
		}

		for _, m := range members {
			if m.Id == rs.Primary.ID {
				return fmt.Errorf("subnet_pool_member (%v) still exists", m.Id)
			}
		}
	}

	return nil
}
