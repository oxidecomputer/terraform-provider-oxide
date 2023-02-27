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

func TestAccResourceVPC(t *testing.T) {
	resourceName := "oxide_vpc.test"
	resourceName2 := "oxide_vpc.test2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviderFactory,
		CheckDestroy:      testAccVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceVPCConfig,
				Check:  checkResourceVPC(resourceName),
			},
			{
				Config: testResourceVPCUpdateConfig,
				Check:  checkResourceVPCUpdate(resourceName),
			},
			{
				Config: testResourceVPCIPv6Config,
				Check:  checkResourceVPCIPv6(resourceName2),
			},
		},
	})
}

var testResourceVPCConfig = `
data "oxide_organizations" "org_list" {}

data "oxide_projects" "project_list" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
}

resource "oxide_vpc" "test" {
	project_id        = data.oxide_projects.project_list.projects.0.id
	description       = "a test vpc"
	name              = "terraform-acc-myvpc"
	dns_name          = "my-vpc-dns"
  }
`

func checkResourceVPC(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vpc"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myvpc"),
		resource.TestCheckResourceAttr(resourceName, "dns_name", "my-vpc-dns"),
		resource.TestCheckResourceAttrSet(resourceName, "ipv6_prefix"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "system_router_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

var testResourceVPCUpdateConfig = `
data "oxide_organizations" "org_list" {}

data "oxide_projects" "project_list" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
}

resource "oxide_vpc" "test" {
	project_id        = data.oxide_projects.project_list.projects.0.id
	description       = "a test vopac"
	name              = "terraform-acc-myvpc-new"
	dns_name          = "my-vpc-donas"
  }
`

func checkResourceVPCUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vopac"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myvpc-new"),
		resource.TestCheckResourceAttr(resourceName, "dns_name", "my-vpc-donas"),
		resource.TestCheckResourceAttrSet(resourceName, "ipv6_prefix"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "system_router_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

var testResourceVPCIPv6Config = `
data "oxide_organizations" "org_list" {}

data "oxide_projects" "project_list" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
}

resource "oxide_vpc" "test2" {
	project_id        = data.oxide_projects.project_list.projects.0.id
	description       = "a test vpc"
	name              = "terraform-acc-myvpc2"
	dns_name          = "my-vpc-dns"
	ipv6_prefix       = "fd1e:4947:d4a1::/48"
  }
`

func checkResourceVPCIPv6(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vpc"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-myvpc2"),
		resource.TestCheckResourceAttr(resourceName, "dns_name", "my-vpc-dns"),
		resource.TestCheckResourceAttr(resourceName, "ipv6_prefix", "fd1e:4947:d4a1::/48"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "system_router_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccVPCDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_vpc" {
			continue
		}

		res, err := client.VpcView("corp", "test", "terraform-acc-myvpc")
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("vpc (%v) still exists", &res.Name)
	}

	return nil
}
