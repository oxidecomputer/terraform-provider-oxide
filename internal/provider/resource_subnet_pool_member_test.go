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

	subnet1 := nextSubnetCIDR(t)
	subnet2 := nextSubnetCIDR(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSubnetPoolMemberDestroy,
		Steps: []resource.TestStep{
			// Create pool and one member
			{
				Config: testResourceSubnetPoolMemberConfig(subnet1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(poolResourceName, "id"),
					resource.TestCheckResourceAttrSet(memberResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						memberResourceName,
						"subnet_pool_id",
						poolResourceName,
						"id",
					),
					resource.TestCheckResourceAttr(memberResourceName, "subnet", subnet1),
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
				Config: testResourceSubnetPoolMemberAddSecondConfig(subnet1, subnet2),
				Check: resource.ComposeAggregateTestCheckFunc(
					// First member unchanged
					resource.TestCheckResourceAttr(memberResourceName, "subnet", subnet1),
					resource.TestCheckResourceAttr(memberResourceName, "min_prefix_length", "26"),
					resource.TestCheckResourceAttr(memberResourceName, "max_prefix_length", "28"),
					// Second member has computed min_prefix_length (defaults to subnet prefix = 24)
					resource.TestCheckResourceAttrSet(member2ResourceName, "id"),
					resource.TestCheckResourceAttr(
						member2ResourceName,
						"subnet",
						subnet2,
					),
					resource.TestCheckResourceAttr(member2ResourceName, "min_prefix_length", "24"),
					resource.TestCheckResourceAttr(member2ResourceName, "max_prefix_length", "30"),
				),
			},
		},
	})
}

func testResourceSubnetPoolMemberConfig(subnet string) string {
	return fmt.Sprintf(`
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-member-test"
	description = "a test subnet pool for member tests"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_member" "test" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = %q
	min_prefix_length = 26
	max_prefix_length = 28
}
`, subnet)
}

func testResourceSubnetPoolMemberAddSecondConfig(subnet1, subnet2 string) string {
	return fmt.Sprintf(`
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-member-test"
	description = "a test subnet pool for member tests"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_member" "test" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = %q
	min_prefix_length = 26
	max_prefix_length = 28
}

resource "oxide_subnet_pool_member" "test2" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = %q
	# min_prefix_length omitted - should be computed as 24 (subnet prefix)
	max_prefix_length = 30
}
`, subnet1, subnet2)
}

func TestAccResourceSubnetPoolMember_parallel(t *testing.T) {
	poolResourceName := "oxide_subnet_pool.test"
	member1ResourceName := "oxide_subnet_pool_member.m1"
	member2ResourceName := "oxide_subnet_pool_member.m2"
	member3ResourceName := "oxide_subnet_pool_member.m3"

	sub1 := nextSubnetCIDR(t)
	sub2 := nextSubnetCIDR(t)
	sub3 := nextSubnetCIDR(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSubnetPoolMemberDestroy,
		Steps: []resource.TestStep{
			// Create pool and three members in parallel (no depends_on between members)
			{
				Config: testResourceSubnetPoolMemberParallelConfig(sub1, sub2, sub3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(poolResourceName, "id"),
					resource.TestCheckResourceAttrSet(member1ResourceName, "id"),
					resource.TestCheckResourceAttrSet(member2ResourceName, "id"),
					resource.TestCheckResourceAttrSet(member3ResourceName, "id"),
					resource.TestCheckResourceAttr(member1ResourceName, "subnet", sub1),
					resource.TestCheckResourceAttr(member2ResourceName, "subnet", sub2),
					resource.TestCheckResourceAttr(member3ResourceName, "subnet", sub3),
				),
			},
		},
	})
}

func testResourceSubnetPoolMemberParallelConfig(sub1, sub2, sub3 string) string {
	return fmt.Sprintf(`
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-parallel-test"
	description = "a test subnet pool for parallel member creation"
	ip_version  = "v4"
}

# These three members have no dependencies on each other,
# so Terraform will attempt to create them in parallel.
resource "oxide_subnet_pool_member" "m1" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = %q
	max_prefix_length = 28
}

resource "oxide_subnet_pool_member" "m2" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = %q
	max_prefix_length = 28
}

resource "oxide_subnet_pool_member" "m3" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = %q
	max_prefix_length = 28
}
`, sub1, sub2, sub3)
}

func TestAccResourceSubnetPoolMember_disappears(t *testing.T) {
	poolResourceName := "oxide_subnet_pool.test"
	memberResourceName := "oxide_subnet_pool_member.test"

	subnet := nextSubnetCIDR(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSubnetPoolMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceSubnetPoolMemberDisappearsConfig(subnet),
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

func testResourceSubnetPoolMemberDisappearsConfig(subnet string) string {
	return fmt.Sprintf(`
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-disappears-test"
	description = "a test subnet pool for disappears test"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_member" "test" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = %q
	max_prefix_length = 28
}
`, subnet)
}

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
