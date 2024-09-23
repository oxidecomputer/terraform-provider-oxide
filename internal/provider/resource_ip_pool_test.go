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

func TestAccSiloResourceIpPool_full(t *testing.T) {
	resourceName := "oxide_ip_pool.test"
	resourceName2 := "oxide_ip_pool.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccIPPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceIPPoolConfig,
				Check:  checkResourceIPPool(resourceName),
			},
			{
				Config: testResourceIPPoolUpdateConfig,
				Check:  checkResourceIPPoolUpdate(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceIPPoolRemoveUpdateConfig,
				Check:  checkResourceIPPoolRemoveUpdate(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceIPPoolRangesConfig,
				Check:  checkResourceIPPoolRanges(resourceName2),
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

var testResourceIPPoolConfig = `
resource "oxide_ip_pool" "test" {
	description       = "a test ip_pool"
	name              = "terraform-acc-myippool"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

func checkResourceIPPool(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test ip_pool"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myippool"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

var testResourceIPPoolUpdateConfig = `
resource "oxide_ip_pool" "test" {
	description       = "a new description for ip_pool"
	name              = "terraform-acc-myippool-new"
	ranges = [
    {
		first_address = "172.20.15.227"
		last_address  = "172.20.15.230"
	},
	{
		first_address = "172.20.15.231"
		last_address  = "172.20.15.233"
	}
  ]
}
`

func checkResourceIPPoolUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a new description for ip_pool"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myippool-new"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttrSet(resourceName, "ranges.0.first_address"),
		resource.TestCheckResourceAttrSet(resourceName, "ranges.0.last_address"),
		resource.TestCheckResourceAttrSet(resourceName, "ranges.1.first_address"),
		resource.TestCheckResourceAttrSet(resourceName, "ranges.1.last_address"),
	}...)
}

var testResourceIPPoolRemoveUpdateConfig = `
resource "oxide_ip_pool" "test" {
	description       = "a new description for ip_pool"
	name              = "terraform-acc-myippool-new"
	ranges = [
    {
		first_address = "172.20.15.227"
		last_address  = "172.20.15.230"
	}
  ]
}
`

func checkResourceIPPoolRemoveUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a new description for ip_pool"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myippool-new"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "ranges.0.first_address", "172.20.15.227"),
		resource.TestCheckResourceAttr(resourceName, "ranges.0.last_address", "172.20.15.230"),
	}...)
}

var testResourceIPPoolRangesConfig = `
resource "oxide_ip_pool" "test2" {
	description       = "a test ip_pool"
	name              = "terraform-acc-myippool2"
	ranges = [
    {
		first_address = "172.20.15.240"
		last_address  = "172.20.15.249"
	}
  ]
}
`

func checkResourceIPPoolRanges(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test ip_pool"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myippool2"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "ranges.0.first_address", "172.20.15.240"),
		resource.TestCheckResourceAttr(resourceName, "ranges.0.last_address", "172.20.15.249"),
	}...)
}

func testAccIPPoolDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_ip_pool" {
			continue
		}

		ctx := context.Background()

		res, err := client.IpPoolView(
			ctx,
			oxide.IpPoolViewParams{Pool: "terraform-acc-myippool"},
		)
		if err == nil || !is404(err) {
			return fmt.Errorf("ip_pool (%v) still exists", &res.Name)
		}

		res2, err := client.IpPoolView(
			ctx,
			oxide.IpPoolViewParams{Pool: "terraform-acc-myippool2"},
		)
		if err != nil && is404(err) {
			continue
		}
		return fmt.Errorf("ip_pool (%v) still exists", &res2.Name)
	}

	return nil
}
