// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippool_test

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccSystemResourceIPPool_full(t *testing.T) {
	resourceName := "oxide_system_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig,
				Check:  checkResource(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceUpdateConfig,
				Check:  checkResourceUpdate(resourceName),
			},
		},
	})
}

func TestAccSystemResourceIPPool_moveStateAndImportRanges(t *testing.T) {
	t.Setenv("TF_ACC_PROVIDER_NAMESPACE", "oxidecomputer")

	poolName := sharedtest.NewResourceName()
	first, last := testRangeAddresses(poolName)
	var poolID, rangeID, linkID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testLegacyResourceConfig(poolName, first, last),
				Check: resource.ComposeAggregateTestCheckFunc(
					sharedtest.CaptureResourceID("oxide_ip_pool.test", &poolID),
					sharedtest.CaptureResourceID(
						"oxide_ip_pool_silo_link.test",
						&linkID,
					),
					captureRangeID(poolName, first, last, &rangeID),
				),
			},
			{
				Config: testMovedResourceConfig(poolName, first, last),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							"oxide_system_ip_pool.test",
							plancheck.ResourceActionNoop,
						),
						plancheck.ExpectResourceAction(
							"oxide_system_ip_pool_range.test",
							plancheck.ResourceActionNoop,
						),
						plancheck.ExpectResourceAction(
							"oxide_system_ip_pool_silo_link.test",
							plancheck.ResourceActionNoop,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(
						"oxide_system_ip_pool.test",
						"id",
						&poolID,
					),
					resource.TestCheckResourceAttrPtr(
						"oxide_system_ip_pool_range.test",
						"id",
						&rangeID,
					),
					resource.TestCheckResourceAttrPtr(
						"oxide_system_ip_pool_silo_link.test",
						"id",
						&linkID,
					),
				),
			},
		},
	})
}

func testRangeAddresses(poolName string) (string, string) {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(poolName))
	value := hash.Sum32()
	prefix := fmt.Sprintf("172.%d.%d", 16+(value>>8)%16, value%256)
	return prefix + ".10", prefix + ".19"
}

func testLegacyResourceConfig(poolName, first, last string) string {
	return fmt.Sprintf(`
data "oxide_silo" "test" {
  name = "test-suite-silo"
}

resource "oxide_ip_pool" "test" {
  name        = %[1]q
  description = "an IP pool state move test"
  ranges = [{
    first_address = %[2]q
    last_address  = %[3]q
  }]
}

resource "oxide_ip_pool_silo_link" "test" {
  ip_pool_id = oxide_ip_pool.test.id
  silo_id    = data.oxide_silo.test.id
  is_default = false
}
`, poolName, first, last)
}

func testMovedResourceConfig(poolName, first, last string) string {
	return fmt.Sprintf(`
data "oxide_silo" "test" {
  name = "test-suite-silo"
}

resource "oxide_system_ip_pool" "test" {
  name        = %[1]q
  description = "an IP pool state move test"
}

resource "oxide_system_ip_pool_range" "test" {
  ip_pool_id    = oxide_system_ip_pool.test.id
  first_address = %[2]q
  last_address  = %[3]q
}

resource "oxide_system_ip_pool_silo_link" "test" {
  ip_pool_id = oxide_system_ip_pool.test.id
  silo_id    = data.oxide_silo.test.id
  is_default = false
}

moved {
  from = oxide_ip_pool.test
  to   = oxide_system_ip_pool.test
}

moved {
  from = oxide_ip_pool_silo_link.test
  to   = oxide_system_ip_pool_silo_link.test
}

import {
  to = oxide_system_ip_pool_range.test
  id = "${oxide_system_ip_pool.test.id}/%[2]s/%[3]s"
}
`, poolName, first, last)
}

func captureRangeID(
	poolName, first, last string,
	rangeID *string,
) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}

		ranges, err := client.SystemIpPoolRangeListAllPages(
			context.Background(),
			oxide.SystemIpPoolRangeListParams{
				Pool: oxide.NameOrId(poolName),
			},
		)
		if err != nil {
			return err
		}
		want := fmt.Sprintf("%s-%s", first, last)
		for _, ipRange := range ranges {
			if ipRange.Range.String() == want {
				*rangeID = ipRange.Id
				return nil
			}
		}

		return fmt.Errorf("range %s not found in system IP pool %s", want, poolName)
	}
}

var testResourceConfig = `
resource "oxide_system_ip_pool" "test" {
	name        = "terraform-acc-system-ippool"
	description = "a system IP pool"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

func checkResource(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a system IP pool"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-system-ippool"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

var testResourceUpdateConfig = `
resource "oxide_system_ip_pool" "test" {
	name        = "terraform-acc-system-ippool-new"
	description = "a new description for ip_pool"
}
`

func checkResourceUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(
			resourceName,
			"description",
			"a new description for ip_pool",
		),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-system-ippool-new"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccResourceDestroy(s *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_system_ip_pool" {
			continue
		}

		_, err := client.SystemIpPoolView(
			context.Background(),
			oxide.SystemIpPoolViewParams{Pool: oxide.NameOrId(rs.Primary.ID)},
		)
		if err == nil || !errors.Is(err, oxide.ErrHTTP404) {
			return fmt.Errorf("IP pool (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
