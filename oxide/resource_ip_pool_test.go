// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceIpPool(t *testing.T) {
	resourceName := "oxide_ip_pool.test"
	resourceName2 := "oxide_ip_pool.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccIpPoolDestroy,
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
				Config: testResourceIpPoolOrgProjectConfig,
				Check:  checkResourceIpPoolOrgProject(resourceName2),
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
	name              = "terraform-acc-myippool"
  }
`

func checkResourceIpPoolUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a new description for ip_pool"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myippool"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

var testResourceIpPoolOrgProjectConfig = `
resource "oxide_ip_pool" "test2" {
	organization_name = "corp"
	project_name      = "test"
	description       = "a test ip_pool"
	name              = "terraform-acc-myippool2"
  }
`

func checkResourceIpPoolOrgProject(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "organization_name", "corp"),
		resource.TestCheckResourceAttr(resourceName, "project_name", "test"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test ip_pool"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myippool2"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
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

		res, err := client.IpPoolView("terraform-acc-myippool")
		if err == nil || !is404(err) {
			return fmt.Errorf("ip_pool (%v) still exists", &res.Name)
		}

		res2, err := client.IpPoolView("terraform-acc-myippool2")
		if err != nil && is404(err) {
			continue
		}
		return fmt.Errorf("ip_pool (%v) still exists", &res2.Name)
	}

	return nil
}
