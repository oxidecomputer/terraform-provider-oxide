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

type resourceVPCInternetGatewayConfig struct {
	BlockName              string
	VPCName                string
	SupportBlockName       string
	VPCBlockName           string
	VPCInternetGatewayName string
}

var resourceVPCInternetGatewayConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "{{.VPCBlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	description = "a test vpc"
	name        = "{{.VPCName}}"
	dns_name    = "my-vpc"
}

resource "oxide_vpc_internet_gateway" "{{.BlockName}}" {
	description = "a test internet gateway"
	name        = "{{.VPCInternetGatewayName}}"
	vpc_id      = oxide_vpc.{{.VPCBlockName}}.id
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
  }
`

var resourceVPCInternetGatewayUpdateConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "{{.VPCBlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	description = "a test vpc"
	name        = "{{.VPCName}}"
	dns_name    = "my-vpc"
}

resource "oxide_vpc_internet_gateway" "{{.BlockName}}" {
	cascade_delete = true
	description    = "a test internet gateway"
	name           = "{{.VPCInternetGatewayName}}"
	vpc_id         = oxide_vpc.{{.VPCBlockName}}.id
  }
`

func TestAccCloudResourceVPCInternetGateway_full(t *testing.T) {
	vpcName := newResourceName()
	internetGatewayName := newResourceName()
	blockName := newBlockName("internet_gateway")
	supportBlockName := newBlockName("support")
	vpcBlockName := newBlockName("vpc")
	resourceName := fmt.Sprintf("oxide_vpc_internet_gateway.%s", blockName)
	config, err := parsedAccConfig(
		resourceVPCInternetGatewayConfig{
			VPCName:                vpcName,
			SupportBlockName:       supportBlockName,
			VPCBlockName:           vpcBlockName,
			BlockName:              blockName,
			VPCInternetGatewayName: internetGatewayName,
		},
		resourceVPCInternetGatewayConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configUpdate, err := parsedAccConfig(
		resourceVPCInternetGatewayConfig{
			VPCName:                vpcName,
			SupportBlockName:       supportBlockName,
			VPCBlockName:           vpcBlockName,
			BlockName:              blockName,
			VPCInternetGatewayName: internetGatewayName,
		},
		resourceVPCInternetGatewayUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccVPCInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceVPCInternetGateway(resourceName, internetGatewayName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceVPCInternetGatewayUpdate(resourceName, internetGatewayName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// Value for cascade_delete cannot be imported as it is only a query parameter that
				// can be passed during a delete
				ImportStateVerifyIgnore: []string{"cascade_delete"},
			},
		},
	})
}

func checkResourceVPCInternetGateway(
	resourceName, internetGatewayName string,
) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "cascade_delete", "false"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test internet gateway"),
		resource.TestCheckResourceAttr(resourceName, "name", internetGatewayName),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func checkResourceVPCInternetGatewayUpdate(
	resourceName, internetGatewayName string,
) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttr(resourceName, "cascade_delete", "true"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test internet gateway"),
		resource.TestCheckResourceAttr(resourceName, "name", internetGatewayName),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccVPCInternetGatewayDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_vpc_internet_gateway" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.InternetGatewayViewParams{
			Gateway: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.InternetGatewayView(ctx, params)
		if err != nil && is404(err) {
			continue
		}
		return fmt.Errorf("internet gateway (%v) still exists", &res.Name)
	}

	return nil
}
