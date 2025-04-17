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

type resourceVPCRouterRouteConfig struct {
	BlockName          string
	VPCName            string
	SupportBlockName   string
	VPCBlockName       string
	VPCRouterName      string
	VPCRouterBlockName string
	VPCRouterRouteName string
}

var resourceVPCRouterRouteConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "{{.VPCBlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	description = "a test vpc"
	name        = "{{.VPCName}}"
	dns_name    = "my-vpc"
}

resource "oxide_vpc_router" "{{.VPCRouterBlockName}}" {
	description = "a test router"
	name        = "{{.VPCRouterName}}"
	vpc_id      = oxide_vpc.{{.VPCBlockName}}.id
}

resource "oxide_vpc_router_route" "{{.BlockName}}" {
	description    = "a test route"
	name           = "{{.VPCRouterRouteName}}"
	vpc_router_id  = oxide_vpc_router.{{.VPCRouterBlockName}}.id
	destination = {
		type  = "ip_net"
		value = "::/0"
	}
	target = {
		type  = "ip"
		value = "::1" 
	}
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

var resourceVPCRouterRouteUpdateConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "{{.VPCBlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	description = "a test vpc"
	name        = "{{.VPCName}}"
	dns_name    = "my-vpc"
}

resource "oxide_vpc_router" "{{.VPCRouterBlockName}}" {
	description = "a new description for router"
	name        = "{{.VPCRouterName}}"
	vpc_id      = oxide_vpc.{{.VPCBlockName}}.id
}

resource "oxide_vpc_router_route" "{{.BlockName}}" {
	description    = "a new description for the route"
	name           = "{{.VPCRouterRouteName}}"
	vpc_router_id  = oxide_vpc_router.{{.VPCRouterBlockName}}.id
	destination = {
		type  = "ip_net"
		value = "::/0"
	}
	target = {
		type  = "ip"
		value = "::1" 
	}
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

func TestAccCloudResourceVPCRouterRoute_full(t *testing.T) {
	vpcName := newResourceName()
	routerName := newResourceName()
	routerRouteName := newResourceName()
	routerBlockName := newBlockName("router")
	blockName := newBlockName("route")
	supportBlockName := newBlockName("support")
	vpcBlockName := newBlockName("vpc")
	resourceName := fmt.Sprintf("oxide_vpc_router_route.%s", blockName)
	config, err := parsedAccConfig(
		resourceVPCRouterRouteConfig{
			VPCName:            vpcName,
			SupportBlockName:   supportBlockName,
			VPCBlockName:       vpcBlockName,
			BlockName:          blockName,
			VPCRouterName:      routerName,
			VPCRouterBlockName: routerBlockName,
			VPCRouterRouteName: routerRouteName,
		},
		resourceVPCRouterRouteConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	routerRouteNameUpdated := routerRouteName + "-updated"
	configUpdate, err := parsedAccConfig(
		resourceVPCRouterRouteConfig{
			VPCName:            vpcName,
			SupportBlockName:   supportBlockName,
			VPCBlockName:       vpcBlockName,
			BlockName:          blockName,
			VPCRouterName:      routerName,
			VPCRouterRouteName: routerRouteNameUpdated,
			VPCRouterBlockName: routerBlockName,
		},
		resourceVPCRouterRouteUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccVPCRouterRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceVPCRouterRoute(resourceName, routerRouteName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceVPCRouterRouteUpdate(resourceName, routerRouteNameUpdated),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceVPCRouterRoute(resourceName, routerName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test route"),
		resource.TestCheckResourceAttr(resourceName, "kind", string(oxide.RouterRouteKindCustom)),
		resource.TestCheckResourceAttr(resourceName, "name", routerName),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_router_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func checkResourceVPCRouterRouteUpdate(resourceName, routerName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_router_id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a new description for the route"),
		resource.TestCheckResourceAttr(resourceName, "kind", string(oxide.RouterRouteKindCustom)),
		resource.TestCheckResourceAttr(resourceName, "name", routerName),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccVPCRouterRouteDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_vpc_router_route" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.VpcRouterViewParams{
			Router: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}
		res, err := client.VpcRouterView(ctx, params)
		if err != nil && is404(err) {
			continue
		}
		return fmt.Errorf("router (%v) still exists", &res.Name)
	}

	return nil
}
