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

type resourceInstanceNICConfig struct {
	BlockName         string
	NICName           string
	SupportBlockName  string
	InstanceBlockName string
	InstanceName      string
	SubnetBlockName   string
	SubnetName        string
	VPCBlockName      string
	VPCName           string
}

var resourceInstanceNICConfigTpl = `
 data "oxide_projects" "{{.SupportBlockName}}" {}

 resource "oxide_vpc" "{{.VPCBlockName}}" {
 	project_id  = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
 	description = "a test vpc"
 	name        = "{{.VPCName}}"
 	dns_name    = "my-vpc-dns"
 }

 resource "oxide_vpc_subnet" "{{.SubnetBlockName}}" {
 	vpc_id      = oxide_vpc.{{.VPCBlockName}}.id
 	description = "a test vpc subnet"
 	name        = "{{.SubnetName}}"
 	ipv4_block  = "192.168.1.0/24"
 }

 resource "oxide_instance" "{{.InstanceBlockName}}" {
   project_id      = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].id), 0)
   description     = "a test instance"
   name            = "{{.InstanceName}}"
   host_name       = "terraform-acc-myhost"
   memory          = 1073741824
   ncpus           = 1
   start_on_create = false
 }

 resource "oxide_instance_network_interface" "test" {
   instance_id = oxide_instance.{{.InstanceBlockName}}.id
   subnet_id   = oxide_vpc_subnet.{{.SubnetBlockName}}.id
   vpc_id      = oxide_vpc.{{.VPCBlockName}}.id
   description = "a test nic"
   name        = "{{.NICName}}"
 }
 `

func TestAccResourceInstanceNIC_full(t *testing.T) {
	nicName := newResourceName()
	subnetName := newResourceName()
	vpcName := newResourceName()
	instanceName := newResourceName()
	blockName := newBlockName("nic")
	vpcBlockName := newBlockName("nic")
	subnetBlockName := newBlockName("nic")
	instanceBlockName := newBlockName("nic")
	supportBlockName := newBlockName("support")
	resourceName := fmt.Sprintf("oxide_instance_network_interface.%s", blockName)
	config, err := parsedAccConfig(
		resourceInstanceNICConfig{
			BlockName:         blockName,
			NICName:           nicName,
			VPCBlockName:      vpcBlockName,
			VPCName:           vpcName,
			SubnetBlockName:   subnetBlockName,
			SubnetName:        subnetName,
			InstanceBlockName: instanceBlockName,
			InstanceName:      instanceName,
			SupportBlockName:  supportBlockName,
		},
		resourceInstanceNICConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccInstanceNICDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceInstanceNIC(resourceName, nicName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceInstanceNIC(resourceName, nicName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
		resource.TestCheckResourceAttrSet(resourceName, "subnet_id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test nic"),
		resource.TestCheckResourceAttr(resourceName, "name", nicName),
		resource.TestCheckResourceAttrSet(resourceName, "ip_address"),
		resource.TestCheckResourceAttrSet(resourceName, "mac_address"),
		resource.TestCheckResourceAttrSet(resourceName, "primary"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
	}...)
}

func testAccInstanceNICDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_instance_network_interface" {
			continue
		}

		params := oxideSDK.InstanceNetworkInterfaceViewParams{
			Interface: oxideSDK.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.InstanceNetworkInterfaceView(params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("instance NIC (%v) still exists", &res.Name)
	}

	return nil
}
