// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

func TestAccResourceIpPool(t *testing.T) {
	resourceName := "oxide_ip_pool.test"
	resourceName2 := "oxide_ip_pool.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccIpPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceIpPoolConfig,
				Check:  checkResourceIpPool(resourceName),
			},
			{
				Config: testResourceIpPoolUpdateConfig,
				Check:  checkResourceIpPoolUpdate(resourceName),
			},
			{
				Config: testResourceIpPoolRangesConfig,
				Check:  checkResourceIpPoolRanges(resourceName2),
			},
		},
	})
}

var testResourceIpPoolConfig = `
resource "oxide_ip_pool" "test" {
	description       = "a test ip_pool"
	name              = "terraform-acc-myippool"
}
`

func checkResourceIpPool(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test ip_pool"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myippool"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

var testResourceIpPoolUpdateConfig = `
resource "oxide_ip_pool" "test" {
	description       = "a new description for ip_pool"
	name              = "terraform-acc-myippool-new"
}
`

func checkResourceIpPoolUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a new description for ip_pool"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myippool-new"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

var testResourceIpPoolRangesConfig = `
resource "oxide_ip_pool" "test2" {
	description       = "a test ip_pool"
	name              = "terraform-acc-myippool2"
	ranges = [
    {
		first_address = "172.20.15.227"
		last_address  = "172.20.15.239"
	}
  ]
}
`

func checkResourceIpPoolRanges(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test ip_pool"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myippool2"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "ranges.0.first_address", "172.20.15.227"),
		resource.TestCheckResourceAttr(resourceName, "ranges.0.last_address", "172.20.15.239"),
		resource.TestCheckResourceAttrSet(resourceName, "ranges.0.id"),
		resource.TestCheckResourceAttrSet(resourceName, "ranges.0.time_created"),
	}...)
}

func testAccIpPoolDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_ip_pool" {
			continue
		}

		res, err := client.IpPoolView(oxide.IpPoolViewParams{Pool: "terraform-acc-myippool"})
		if err == nil || !is404(err) {
			return fmt.Errorf("ip_pool (%v) still exists", &res.Name)
		}

		res2, err := client.IpPoolView(oxide.IpPoolViewParams{Pool: "terraform-acc-myippool2"})
		if err != nil && is404(err) {
			continue
		}
		return fmt.Errorf("ip_pool (%v) still exists", &res2.Name)
	}

	return nil
}
