// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpc_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

type resourceConfig struct {
	BlockName        string
	SupportBlockName string
	VPCName          string
}

var resourceConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "{{.BlockName}}" {
	project_id        = data.oxide_project.{{.SupportBlockName}}.id
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

var resourceUpdateConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "{{.BlockName}}" {
	project_id        = data.oxide_project.{{.SupportBlockName}}.id
	description       = "a test vopac"
	name              = "{{.VPCName}}"
	dns_name          = "my-vpc-donas"
  }
`

var resourceIPv6ConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "{{.BlockName}}" {
	project_id        = data.oxide_project.{{.SupportBlockName}}.id
	description       = "a test vpc"
	name              = "{{.VPCName}}"
	dns_name          = "my-vpc-dns"
	ipv6_prefix       = "fd1e:4947:d4a1::/48"
  }
`

var resourceSkipDefaultSubnetConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "{{.BlockName}}" {
	project_id          = data.oxide_project.{{.SupportBlockName}}.id
	description         = "a test VPC without a default subnet"
	name                = "{{.VPCName}}"
	dns_name            = "no-default-subnet"
	skip_default_subnet = true
}
`

func TestAccCloudResourceVPC_full(t *testing.T) {
	vpcName := sharedtest.NewResourceName()
	blockName := sharedtest.NewBlockName("vpc")
	resourceName := fmt.Sprintf("oxide_vpc.%s", blockName)
	supportBlockName := sharedtest.NewBlockName("support")
	config := sharedtest.ParsedAccConfig(t,
		resourceConfig{
			BlockName:        blockName,
			VPCName:          vpcName,
			SupportBlockName: supportBlockName,
		},
		resourceConfigTpl,
	)

	vpcNameUpdated := vpcName + "-updated"
	configUpdate := sharedtest.ParsedAccConfig(t,
		resourceConfig{
			BlockName:        blockName,
			VPCName:          vpcNameUpdated,
			SupportBlockName: supportBlockName,
		},
		resourceUpdateConfigTpl,
	)

	blockName2 := sharedtest.NewBlockName("vpc")
	resourceName2 := fmt.Sprintf("oxide_vpc.%s", blockName2)
	vpcName2 := vpcName + "-2"
	config2 := sharedtest.ParsedAccConfig(t,
		resourceConfig{
			BlockName:        blockName2,
			VPCName:          vpcName2,
			SupportBlockName: supportBlockName,
		},
		resourceIPv6ConfigTpl,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResource(resourceName, vpcName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceUpdate(resourceName, vpcNameUpdated),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config2,
				Check:  checkResourceIPv6(resourceName2, vpcName2),
			},
			{
				ResourceName:      resourceName2,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCloudResourceVPC_skipDefaultSubnet(t *testing.T) {
	vpcName := sharedtest.NewResourceName()
	blockName := sharedtest.NewBlockName("vpc")
	resourceName := fmt.Sprintf("oxide_vpc.%s", blockName)
	supportBlockName := sharedtest.NewBlockName("support")
	config := sharedtest.ParsedAccConfig(t,
		resourceConfig{
			BlockName:        blockName,
			VPCName:          vpcName,
			SupportBlockName: supportBlockName,
		},
		resourceSkipDefaultSubnetConfigTpl,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						resourceName,
						"skip_default_subnet",
						"true",
					),
					testAccCheckNoSubnets(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"skip_default_subnet",
				},
			},
		},
	})
}

func testAccCheckNoSubnets(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}

		subnets, err := client.VpcSubnetListAllPages(
			context.Background(),
			oxide.VpcSubnetListParams{Vpc: oxide.NameOrId(rs.Primary.ID)},
		)
		if err != nil {
			return err
		}
		if len(subnets) != 0 {
			return fmt.Errorf("expected VPC to have no subnets, got %d", len(subnets))
		}

		return nil
	}
}

func checkResource(resourceName, vpcName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckNoResourceAttr(resourceName, "skip_default_subnet"),
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

func checkResourceUpdate(resourceName, vpcName string) resource.TestCheckFunc {
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

func checkResourceIPv6(resourceName, vpcName string) resource.TestCheckFunc {
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

func testAccResourceDestroy(s *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_vpc" {
			continue
		}

		params := oxide.VpcViewParams{
			Vpc: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		res, err := client.VpcView(ctx, params)
		if err != nil && errors.Is(err, oxide.ErrHTTP404) {
			continue
		}

		return fmt.Errorf("vpc (%v) still exists", &res.Name)
	}

	return nil
}
