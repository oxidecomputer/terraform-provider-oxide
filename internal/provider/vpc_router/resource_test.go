// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcrouter_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

type resourceConfig struct {
	BlockName        string
	VPCName          string
	SupportBlockName string
	VPCBlockName     string
	VPCRouterName    string
}

var resourceConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "{{.VPCBlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	description = "a test vpc"
	name        = "{{.VPCName}}"
	dns_name    = "my-vpc"
}

resource "oxide_vpc_router" "{{.BlockName}}" {
	description = "a test router"
	name        = "{{.VPCRouterName}}"
	vpc_id      = oxide_vpc.{{.VPCBlockName}}.id
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

resource "oxide_vpc" "{{.VPCBlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	description = "a test vpc"
	name        = "{{.VPCName}}"
	dns_name    = "my-vpc"
}

resource "oxide_vpc_router" "{{.BlockName}}" {
	description = "a new description for router"
	name        = "{{.VPCRouterName}}"
	vpc_id      = oxide_vpc.{{.VPCBlockName}}.id
  }
`

func TestAccCloudResourceVPCRouter_full(t *testing.T) {
	vpcName := sharedtest.NewResourceName()
	routerName := sharedtest.NewResourceName()
	blockName := sharedtest.NewBlockName("router")
	supportBlockName := sharedtest.NewBlockName("support")
	vpcBlockName := sharedtest.NewBlockName("vpc")
	resourceName := fmt.Sprintf("oxide_vpc_router.%s", blockName)
	config, err := sharedtest.ParsedAccConfig(
		resourceConfig{
			VPCName:          vpcName,
			SupportBlockName: supportBlockName,
			VPCBlockName:     vpcBlockName,
			BlockName:        blockName,
			VPCRouterName:    routerName,
		},
		resourceConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	routerNameUpdated := routerName + "-updated"
	configUpdate, err := sharedtest.ParsedAccConfig(
		resourceConfig{
			VPCName:          vpcName,
			SupportBlockName: supportBlockName,
			VPCBlockName:     vpcBlockName,
			BlockName:        blockName,
			VPCRouterName:    routerNameUpdated,
		},
		resourceUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResource(resourceName, routerName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceUpdate(resourceName, routerNameUpdated),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResource(resourceName, routerName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test router"),
		resource.TestCheckResourceAttr(resourceName, "kind", string(oxide.RouterRouteKindCustom)),
		resource.TestCheckResourceAttr(resourceName, "name", routerName),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func checkResourceUpdate(resourceName, routerName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a new description for router"),
		resource.TestCheckResourceAttr(resourceName, "kind", string(oxide.RouterRouteKindCustom)),
		resource.TestCheckResourceAttr(resourceName, "name", routerName),
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
		if rs.Type != "oxide_vpc_router" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.VpcRouterViewParams{
			Router: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.VpcRouterView(ctx, params)
		if err != nil && shared.Is404(err) {
			continue
		}
		return fmt.Errorf("router (%v) still exists", &res.Name)
	}

	return nil
}
