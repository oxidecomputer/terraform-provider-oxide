// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

type vpcSubnetTestConfig struct {
	SubnetName  string
	Description string
	IPv4Block   string
	IPv6Block   string // optional, rendered conditionally
}

var vpcSubnetConfigTpl = `
data "oxide_project" "test" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "test" {
	project_id  = data.oxide_project.test.id
	description = "a test vpc"
	name        = "terraform-acc-vpc-subnet-vpc"
	dns_name    = "my-vpc-dns"
	ipv6_prefix = "fdfe:f6a5:5f06::/48"
}

resource "oxide_vpc_subnet" "test" {
	vpc_id      = oxide_vpc.test.id
	description = "{{.Description}}"
	name        = "{{.SubnetName}}"
	ipv4_block  = "{{.IPv4Block}}"
{{- if .IPv6Block}}
	ipv6_block  = "{{.IPv6Block}}"
{{- end}}
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

func buildVPCSubnetConfig(t *testing.T, cfg vpcSubnetTestConfig) string {
	t.Helper()
	config, err := parsedAccConfig(cfg, vpcSubnetConfigTpl)
	if err != nil {
		t.Fatalf("error parsing config template: %v", err)
	}
	return config
}

func TestAccCloudResourceVPCSubnet_full(t *testing.T) {
	resourceName := "oxide_vpc_subnet.test"

	// Initial creation config.
	baseConfig := vpcSubnetTestConfig{
		SubnetName:  "terraform-acc-vpc-subnet",
		Description: "a test vpc subnet",
		IPv4Block:   "192.168.1.0/24",
	}

	// In-place update config.
	updateConfig := baseConfig
	updateConfig.SubnetName = "terraform-acc-vpc-subnet-updated"
	updateConfig.Description = "a test vpc subnety"

	// Recreate config with v6 cidr.
	ipv6Config := baseConfig
	ipv6Config.SubnetName = "terraform-acc-vpc-subnet-v6"
	ipv6Config.IPv6Block = "fdfe:f6a5:5f06:a643::/64"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccVPCSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildVPCSubnetConfig(t, baseConfig),
				Check:  checkResourceVPCSubnet("terraform-acc-vpc-subnet"),
			},
			{
				Config: buildVPCSubnetConfig(t, updateConfig),
				Check:  checkResourceVPCSubnetUpdate("terraform-acc-vpc-subnet-updated"),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: buildVPCSubnetConfig(t, ipv6Config),
				Check:  checkResourceVPCSubnetIPv6("terraform-acc-vpc-subnet-v6"),
			},
		},
	})
}

func checkResourceVPCSubnet(subnetName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "id"),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "description", "a test vpc subnet"),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "name", subnetName),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "ipv4_block", "192.168.1.0/24"),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "ipv6_block"),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "vpc_id"),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "time_created"),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "time_modified"),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "timeouts.read", "1m"),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "timeouts.create", "3m"),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "timeouts.update", "4m"),
	}...)
}

func checkResourceVPCSubnetUpdate(subnetName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "id"),
		resource.TestCheckResourceAttr(
			"oxide_vpc_subnet.test",
			"description",
			"a test vpc subnety",
		),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "name", subnetName),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "ipv4_block", "192.168.1.0/24"),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "ipv6_block"),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "vpc_id"),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "time_created"),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "time_modified"),
	}...)
}

func checkResourceVPCSubnetIPv6(subnetName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "id"),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "description", "a test vpc subnet"),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "name", subnetName),
		resource.TestCheckResourceAttr("oxide_vpc_subnet.test", "ipv4_block", "192.168.1.0/24"),
		resource.TestCheckResourceAttr(
			"oxide_vpc_subnet.test",
			"ipv6_block",
			"fdfe:f6a5:5f06:a643::/64",
		),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "vpc_id"),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "time_created"),
		resource.TestCheckResourceAttrSet("oxide_vpc_subnet.test", "time_modified"),
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

		params := oxide.VpcSubnetViewParams{
			Subnet: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		res, err := client.VpcSubnetView(ctx, params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("vpc subnet (%v) still exists", &res.Name)
	}

	return nil
}
