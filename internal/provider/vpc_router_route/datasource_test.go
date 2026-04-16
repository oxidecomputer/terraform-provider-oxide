// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcrouterroute_test

import (
	"fmt"
	"testing"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/oxidecomputer/oxide.go/oxide"
)

type dataSourceConfig struct {
	BlockName        string
	SupportBlockName string
}

var dataSourceConfigTpl = `
data "oxide_vpc_router_route" "{{.BlockName}}" {
  project_name    = "tf-acc-test"
  vpc_name        = "default"
  vpc_router_name = "system"
  name            = "default-v4"
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceVPCRouterRoute_full(t *testing.T) {
	blockName := sharedtest.NewBlockName("datasource-vpc-router")
	config, err := sharedtest.ParsedAccConfig(
		dataSourceConfig{
			BlockName:        blockName,
			SupportBlockName: sharedtest.NewBlockName("support"),
		},
		dataSourceConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: checkDataSource(
					fmt.Sprintf("data.oxide_vpc_router_route.%s", blockName),
				),
			},
		},
	})
}

func checkDataSource(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(
			dataName,
			"description",
			"The default route of a vpc",
		),
		resource.TestCheckResourceAttr(dataName, "name", "default-v4"),
		resource.TestCheckResourceAttr(
			dataName,
			"destination.type",
			string(oxide.RouteDestinationTypeIpNet),
		),
		resource.TestCheckResourceAttr(dataName, "destination.value", "0.0.0.0/0"),
		resource.TestCheckResourceAttr(
			dataName,
			"target.type",
			string(oxide.RouteTargetTypeInternetGateway),
		),
		resource.TestCheckResourceAttr(dataName, "target.value", "default"),
		resource.TestCheckResourceAttr(dataName, "kind", string(oxide.RouterRouteKindDefault)),
		resource.TestCheckResourceAttrSet(dataName, "vpc_router_id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}
