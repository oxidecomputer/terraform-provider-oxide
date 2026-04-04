// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpc_router_route_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
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

var resourceVPCRouterRouteTargetDropConfigTpl = `
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
		type  = "subnet"
		value = "default"
	}
	target = {
		type  = "drop"
	}
}
`

func TestAccCloudResourceVPCRouterRoute_full(t *testing.T) {
	vpcName := provider.NewResourceName()
	routerName := provider.NewResourceName()
	routerRouteName := provider.NewResourceName()
	routerBlockName := provider.NewBlockName("router")
	blockName := provider.NewBlockName("route")
	supportBlockName := provider.NewBlockName("support")
	vpcBlockName := provider.NewBlockName("vpc")
	resourceName := fmt.Sprintf("oxide_vpc_router_route.%s", blockName)
	config, err := provider.ParsedAccConfig(
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
	configUpdate, err := provider.ParsedAccConfig(
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

	vpcNameDrop := provider.NewResourceName()
	routerNameDrop := provider.NewResourceName()
	routerRouteNameDrop := provider.NewResourceName()
	routerBlockNameDrop := provider.NewBlockName("router")
	supportBlockNameDrop := provider.NewBlockName("support")
	vpcBlockNameDrop := provider.NewBlockName("vpc")
	resourceNameDrop := fmt.Sprintf("oxide_vpc_router_route.%s", blockName)
	configDrop, err := provider.ParsedAccConfig(
		resourceVPCRouterRouteConfig{
			VPCName:            vpcNameDrop,
			SupportBlockName:   supportBlockNameDrop,
			VPCBlockName:       vpcBlockNameDrop,
			BlockName:          blockName,
			VPCRouterName:      routerNameDrop,
			VPCRouterBlockName: routerBlockNameDrop,
			VPCRouterRouteName: routerRouteNameDrop,
		},
		resourceVPCRouterRouteTargetDropConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(t) },
		ProtoV6ProviderFactories: provider.ProviderFactories(),
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
				Config: configDrop,
				Check: checkResourceVPCRouterRouteTargetDrop(
					resourceNameDrop,
					routerRouteNameDrop,
				),
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
		resource.TestCheckResourceAttr(
			resourceName,
			"destination.type",
			string(oxide.RouteDestinationTypeIpNet),
		),
		resource.TestCheckResourceAttr(resourceName, "destination.value", "::/0"),
		resource.TestCheckResourceAttr(
			resourceName,
			"target.type",
			string(oxide.RouteTargetTypeIp),
		),
		resource.TestCheckResourceAttr(resourceName, "target.value", "::1"),
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
		resource.TestCheckResourceAttr(
			resourceName,
			"description",
			"a new description for the route",
		),
		resource.TestCheckResourceAttr(resourceName, "kind", string(oxide.RouterRouteKindCustom)),
		resource.TestCheckResourceAttr(resourceName, "name", routerName),
		resource.TestCheckResourceAttr(
			resourceName,
			"destination.type",
			string(oxide.RouteDestinationTypeIpNet),
		),
		resource.TestCheckResourceAttr(resourceName, "destination.value", "::/0"),
		resource.TestCheckResourceAttr(
			resourceName,
			"target.type",
			string(oxide.RouteTargetTypeIp),
		),
		resource.TestCheckResourceAttr(resourceName, "target.value", "::1"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceVPCRouterRouteTargetDrop(resourceName, routerName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_router_id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test route"),
		resource.TestCheckResourceAttr(resourceName, "kind", string(oxide.RouterRouteKindCustom)),
		resource.TestCheckResourceAttr(resourceName, "name", routerName),
		resource.TestCheckResourceAttr(
			resourceName,
			"destination.type",
			string(oxide.RouteDestinationTypeSubnet),
		),
		resource.TestCheckResourceAttr(resourceName, "destination.value", "default"),
		resource.TestCheckResourceAttr(
			resourceName,
			"target.type",
			string(oxide.RouteTargetTypeDrop),
		),
		resource.TestCheckNoResourceAttr(resourceName, "target.value"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccVPCRouterRouteDestroy(s *terraform.State) error {
	client, err := provider.NewTestClient()
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
		if err != nil && shared.Is404(err) {
			continue
		}
		return fmt.Errorf("router (%v) still exists", &res.Name)
	}

	return nil
}
