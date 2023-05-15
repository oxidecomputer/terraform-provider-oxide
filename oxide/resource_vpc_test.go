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

type resourceVPCConfig struct {
	BlockName        string
	SupportBlockName string
	VPCName          string
}

var resourceVPCConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_vpc" "{{.BlockName}}" {
	project_id        = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
	description       = "a test vpc"
	name              = "{{.VPCName}}"
	dns_name          = "my-vpc-dns"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
  }
`

var resourceVPCUpdateConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_vpc" "{{.BlockName}}" {
	project_id        = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
	description       = "a test vopac"
	name              = "{{.VPCName}}"
	dns_name          = "my-vpc-donas"
  }
`

var resourceVPCIPv6ConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_vpc" "{{.BlockName}}" {
	project_id        = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
	description       = "a test vpc"
	name              = "{{.VPCName}}"
	dns_name          = "my-vpc-dns"
	ipv6_prefix       = "fd1e:4947:d4a1::/48"
  }
`

func TestAccResourceVPC_full(t *testing.T) {
	vpcName := newResourceName()
	blockName := newBlockName("vpc")
	resourceName := fmt.Sprintf("oxide_vpc.%s", blockName)
	supportBlockName := newBlockName("support")
	config, err := parsedAccConfig(
		resourceVPCConfig{
			BlockName:        blockName,
			VPCName:          vpcName,
			SupportBlockName: supportBlockName,
		},
		resourceVPCConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	vpcNameUpdated := vpcName + "-updated"
	configUpdate, err := parsedAccConfig(
		resourceVPCConfig{
			BlockName:        blockName,
			VPCName:          vpcNameUpdated,
			SupportBlockName: supportBlockName,
		},
		resourceVPCUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	blockName2 := newBlockName("vpc")
	resourceName2 := fmt.Sprintf("oxide_vpc.%s", blockName2)
	vpcName2 := vpcName + "-2"
	config2, err := parsedAccConfig(
		resourceVPCConfig{
			BlockName:        blockName2,
			VPCName:          vpcName2,
			SupportBlockName: supportBlockName,
		},
		resourceVPCIPv6ConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccVPCDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceVPC(resourceName, vpcName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceVPCUpdate(resourceName, vpcNameUpdated),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config2,
				Check:  checkResourceVPCIPv6(resourceName2, vpcName2),
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceVPC(resourceName, vpcName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vpc"),
		resource.TestCheckResourceAttr(resourceName, "name", vpcName),
		resource.TestCheckResourceAttr(resourceName, "dns_name", "my-vpc-dns"),
		resource.TestCheckResourceAttrSet(resourceName, "ipv6_prefix"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "system_router_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func checkResourceVPCUpdate(resourceName, vpcName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vopac"),
		resource.TestCheckResourceAttr(resourceName, "name", vpcName),
		resource.TestCheckResourceAttr(resourceName, "dns_name", "my-vpc-donas"),
		resource.TestCheckResourceAttrSet(resourceName, "ipv6_prefix"),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrSet(resourceName, "system_router_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceVPCIPv6(resourceName, vpcName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vpc"),
		resource.TestCheckResourceAttr(resourceName, "name", vpcName),
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

		params := oxideSDK.VpcViewParams{
			Vpc: oxideSDK.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.VpcView(params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("vpc (%v) still exists", &res.Name)
	}

	return nil
}
