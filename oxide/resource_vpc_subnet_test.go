// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

type resourceVPCSubnetConfig struct {
	BlockName        string
	SubnetName       string
	SupportBlockName string
	VPCBlockName     string
	VPCName          string
}

var resourceVPCSubnetConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_vpc" "{{.VPCBlockName}}" {
	project_id  = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
	description = "a test vpc"
	name        = "{{.VPCName}}"
	dns_name    = "my-vpc-dns"
	ipv6_prefix = "fdfe:f6a5:5f06::/48"
}

resource "oxide_vpc_subnet" "{{.BlockName}}" {
	vpc_id      = oxide_vpc.{{.VPCBlockName}}.id
	description = "a test vpc subnet"
	name        = "{{.SubnetName}}"
	ipv4_block  = "192.168.1.0/24"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

var resourceVPCSubnetUpdateConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_vpc" "{{.VPCBlockName}}" {
	project_id  = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
	description = "a test vpc"
	name        = "{{.VPCName}}"
	dns_name    = "my-vpc-dns"
	ipv6_prefix = "fdfe:f6a5:5f06::/48"
}

resource "oxide_vpc_subnet" "{{.BlockName}}" {
	vpc_id      = oxide_vpc.{{.VPCBlockName}}.id
	description = "a test vpc subnety"
	name        = "{{.SubnetName}}"
	ipv4_block  = "192.168.1.0/24"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

var resourceVPCSubnetIPv6ConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

resource "oxide_vpc" "{{.VPCBlockName}}" {
	project_id  = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
	description = "a test vpc"
	name        = "{{.VPCName}}"
	dns_name    = "my-vpc-dns"
	ipv6_prefix = "fdfe:f6a5:5f06::/48"
}

resource "oxide_vpc_subnet" "{{.BlockName}}" {
	vpc_id      = oxide_vpc.{{.VPCBlockName}}.id
	description = "a test vpc subnet"
	name        = "{{.SubnetName}}"
	ipv4_block  = "192.168.0.0/16"
	ipv6_block  = "fdfe:f6a5:5f06:a643::/64"
}
`

func TestAccResourceVPCSubnet_full(t *testing.T) {
	subnetName := fmt.Sprintf("acc-terraform-%s", uuid.New())
	vpcName := fmt.Sprintf("acc-terraform-%s", uuid.New())
	blockName := fmt.Sprintf("acc-resource-subnet-%s", uuid.New())
	vpcBlockName := fmt.Sprintf("acc-resource-subnet-%s", uuid.New())
	supportBlockName := fmt.Sprintf("acc-support-%s", uuid.New())
	resourceName := fmt.Sprintf("oxide_vpc_subnet.%s", blockName)
	config, err := parsedAccConfig(
		resourceVPCSubnetConfig{
			BlockName:        blockName,
			SubnetName:       subnetName,
			VPCName:          vpcName,
			VPCBlockName:     vpcBlockName,
			SupportBlockName: supportBlockName,
		},
		resourceVPCSubnetConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	subnetNameUpdated := subnetName + "-updated"
	configUpdate, err := parsedAccConfig(
		resourceVPCSubnetConfig{
			BlockName:        blockName,
			SubnetName:       subnetNameUpdated,
			VPCName:          vpcName,
			VPCBlockName:     vpcBlockName,
			SupportBlockName: supportBlockName,
		},
		resourceVPCSubnetUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	blockName2 := fmt.Sprintf("acc-resource-subnet-%s", uuid.New())
	resourceName2 := fmt.Sprintf("oxide_vpc_subnet.%s", blockName2)
	subnetName2 := subnetName + "-2"
	config2, err := parsedAccConfig(
		resourceVPCSubnetConfig{
			BlockName:        blockName2,
			SubnetName:       subnetName2,
			VPCName:          vpcName,
			VPCBlockName:     vpcBlockName,
			SupportBlockName: supportBlockName,
		},
		resourceVPCSubnetIPv6ConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccVPCSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceVPCSubnet(resourceName, subnetName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceVPCSubnetUpdate(resourceName, subnetNameUpdated),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config2,
				Check:  checkResourceVPCSubnetIPv6(resourceName2, subnetName2),
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceVPCSubnet(resourceName, subnetName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vpc subnet"),
		resource.TestCheckResourceAttr(resourceName, "name", subnetName),
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

func checkResourceVPCSubnetUpdate(resourceName, subnetName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vpc subnety"),
		resource.TestCheckResourceAttr(resourceName, "name", subnetName),
		resource.TestCheckResourceAttr(resourceName, "ipv4_block", "192.168.1.0/24"),
		resource.TestCheckResourceAttrSet(resourceName, "ipv6_block"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceVPCSubnetIPv6(resourceName, subnetName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test vpc subnet"),
		resource.TestCheckResourceAttr(resourceName, "name", subnetName),
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

		params := oxideSDK.VpcSubnetViewParams{
			Subnet: oxideSDK.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.VpcSubnetView(params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("vpc subnet (%v) still exists", &res.Name)
	}

	return nil
}
