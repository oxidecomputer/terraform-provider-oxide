// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/oxidecomputer/oxide.go/oxide"
)

type dataSourceVPCRouterConfig struct {
	BlockName        string
	SupportBlockName string
}

var dataSourceVPCRouterConfigTpl = `
data "oxide_vpc_router" "{{.BlockName}}" {
  project_name = "tf-acc-test"
  vpc_name     = "default"
  name         = "system"
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceVPCRouter_full(t *testing.T) {
	blockName := newBlockName("datasource-vpc-router")
	config, err := parsedAccConfig(
		dataSourceVPCRouterConfig{
			BlockName:        blockName,
			SupportBlockName: newBlockName("support"),
		},
		dataSourceVPCRouterConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: checkDataSourceVPCRouter(
					fmt.Sprintf("data.oxide_vpc_router.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceVPCRouter(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(
			dataName,
			"description",
			"Routes are automatically added to this router as vpc subnets are created",
		),
		resource.TestCheckResourceAttr(dataName, "name", "system"),
		resource.TestCheckResourceAttr(dataName, "kind", string(oxide.VpcRouterKindSystem)),
		resource.TestCheckResourceAttrSet(dataName, "vpc_id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}
