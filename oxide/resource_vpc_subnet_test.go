// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

func TestAccResourceVPCSubnet_full(t *testing.T) {
	resourceName := "oxide_vpc_subnet.test"
	resourceName2 := "oxide_vpc_subnet.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccVPCSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceVPCSubnetConfig,
				Check:  checkResourceVPCSubnet(resourceName),
			},
			{
				Config: testResourceVPCSubnetUpdateConfig,
				Check:  checkResourceVPCSubnetUpdate(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceVPCSubnetIPv6Config,
				Check:  checkResourceVPCSubnetIPv6(resourceName2),
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

var testResourceVPCSubnetConfig = `
data "oxide_projects" "project_list" {}

resource "oxide_vpc" "test" {
	project_id  = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
	description = "a test vpc"
	name        = "terraform-acc-myvpcsubnet"
	dns_name    = "my-vpc-dns"
	ipv6_prefix = "fdfe:f6a5:5f06::/48"
}

resource "oxide_vpc_subnet" "test" {
	vpc_id      = oxide_vpc.test.id
	description = "a test vpc subnet"
	name        = "terraform-acc-mysubnet"
	ipv4_block  = "192.168.1.0/24"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

func checkResourceVPCSubnet(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vpc subnet"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-mysubnet"),
		resource.TestCheckResourceAttr(resourceName, "ipv4_block", "192.168.1.0/24"),
		resource.TestCheckResourceAttrSet(resourceName, "ipv6_block"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

var testResourceVPCSubnetUpdateConfig = `
data "oxide_projects" "project_list" {}

resource "oxide_vpc" "test" {
	project_id  = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
	description = "a test vpc"
	name        = "terraform-acc-myvpcsubnet"
	dns_name    = "my-vpc-dns"
	ipv6_prefix = "fdfe:f6a5:5f06::/48"
}

resource "oxide_vpc_subnet" "test" {
	vpc_id      = oxide_vpc.test.id
	description = "a test vpc subnety"
	name        = "terraform-acc-mysubnety"
	ipv4_block  = "192.168.1.0/24"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

func checkResourceVPCSubnetUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vpc subnety"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-mysubnety"),
		resource.TestCheckResourceAttr(resourceName, "ipv4_block", "192.168.1.0/24"),
		resource.TestCheckResourceAttrSet(resourceName, "ipv6_block"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

var testResourceVPCSubnetIPv6Config = `
data "oxide_projects" "project_list" {}

resource "oxide_vpc" "test" {
	project_id  = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
	description = "a test vpc"
	name        = "terraform-acc-myvpcsubnet"
	dns_name    = "my-vpc-dns"
	ipv6_prefix = "fdfe:f6a5:5f06::/48"
}

resource "oxide_vpc_subnet" "test2" {
	vpc_id      = oxide_vpc.test.id
	description = "a test vpc subnet"
	name        = "terraform-acc-mysubnet-2"
	ipv4_block  = "192.168.0.0/16"
	ipv6_block  = "fdfe:f6a5:5f06:a643::/64"
}
`

func checkResourceVPCSubnetIPv6(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vpc subnet"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-mysubnet-2"),
		resource.TestCheckResourceAttr(resourceName, "ipv4_block", "192.168.0.0/16"),
		resource.TestCheckResourceAttr(resourceName, "ipv6_block", "fdfe:f6a5:5f06:a643::/64"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccVPCSubnetDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_vpc_subnet" {
			continue
		}

		params := oxideSDK.VpcViewParams{
			Project: "test",
			Vpc:     "terraform-acc-myvpcsubnet",
		}
		res, err := client.VpcView(params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("vpc (%v) still exists", &res.Name)
	}

	return nil
}
