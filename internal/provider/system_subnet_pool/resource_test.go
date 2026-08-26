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

func TestAccResourceSystemSubnetPool_moveState(t *testing.T) {
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "oxidecomputer")

	poolName := sharedtest.NewResourceName()
	subnet := sharedtest.NextSubnetCIDR(t)
	var poolID, memberID, linkID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccSystemResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLegacyResourcesConfig(poolName, subnet),
				Check: resource.ComposeAggregateTestCheckFunc(
					sharedtest.CaptureResourceID("oxide_subnet_pool.test", &poolID),
					sharedtest.CaptureResourceID(
						"oxide_subnet_pool_member.test",
						&memberID,
					),
					sharedtest.CaptureResourceID(
						"oxide_subnet_pool_silo_link.test",
						&linkID,
					),
				),
			},
			{
				Config: testMovedResourcesConfig(poolName, subnet),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"oxide_system_subnet_pool.test",
							plancheck.ResourceActionNoop,
						),
						plancheck.ExpectResourceAction(
							"oxide_system_subnet_pool_member.test",
							plancheck.ResourceActionNoop,
						),
						plancheck.ExpectResourceAction(
							"oxide_system_subnet_pool_silo_link.test",
							plancheck.ResourceActionNoop,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(
						"oxide_system_subnet_pool.test",
						"id",
						&poolID,
					),
					resource.TestCheckResourceAttrPtr(
						"oxide_system_subnet_pool_member.test",
						"id",
						&memberID,
					),
					resource.TestCheckResourceAttrPtr(
						"oxide_system_subnet_pool_silo_link.test",
						"id",
						&linkID,
					),
				),
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

func testLegacyResourcesConfig(poolName, subnet string) string {
	return fmt.Sprintf(`
data "oxide_silo" "test" {
  name = "test-suite-silo"
}

resource "oxide_subnet_pool" "test" {
  name        = %[1]q
  description = "a subnet pool state move test"
  ip_version  = "v4"
}

resource "oxide_subnet_pool_member" "test" {
  subnet_pool_id    = oxide_subnet_pool.test.id
  subnet            = %[2]q
  min_prefix_length = 24
  max_prefix_length = 28
}

resource "oxide_subnet_pool_silo_link" "test" {
  subnet_pool_id = oxide_subnet_pool.test.id
  silo_id        = data.oxide_silo.test.id
  is_default     = false
}
`, poolName, subnet)
}

func testMovedResourcesConfig(poolName, subnet string) string {
	return fmt.Sprintf(`
data "oxide_silo" "test" {
  name = "test-suite-silo"
}

resource "oxide_system_subnet_pool" "test" {
  name        = %[1]q
  description = "a subnet pool state move test"
  ip_version  = "v4"
}

resource "oxide_system_subnet_pool_member" "test" {
  subnet_pool_id    = oxide_system_subnet_pool.test.id
  subnet            = %[2]q
  min_prefix_length = 24
  max_prefix_length = 28
}

resource "oxide_system_subnet_pool_silo_link" "test" {
  subnet_pool_id = oxide_system_subnet_pool.test.id
  silo_id        = data.oxide_silo.test.id
  is_default     = false
}

moved {
  from = oxide_subnet_pool.test
  to   = oxide_system_subnet_pool.test
}

moved {
  from = oxide_subnet_pool_member.test
  to   = oxide_system_subnet_pool_member.test
}

moved {
  from = oxide_subnet_pool_silo_link.test
  to   = oxide_system_subnet_pool_silo_link.test
}
`, poolName, subnet)
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
