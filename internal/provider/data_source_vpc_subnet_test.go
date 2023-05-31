// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceVPCSubnetConfig struct {
	BlockName        string
	SupportBlockName string
}

var dataSourceVPCSubnetConfigTpl = `
data "oxide_projects" "{{.SupportBlockName}}" {}

data "oxide_vpc_subnet" "{{.BlockName}}" {
  project_name = element(tolist(data.oxide_projects.{{.SupportBlockName}}.projects[*].name), 0)
  vpc_name     = "default"
  name         = "default"
  timeouts = {
    read = "1m"
  }
}
`

func TestAccDataSourceVPCSubnet_full(t *testing.T) {
	blockName := newBlockName("datasource-vpc-subnet")
	config, err := parsedAccConfig(
		dataSourceVPCSubnetConfig{
			BlockName:        blockName,
			SupportBlockName: newBlockName("support"),
		},
		dataSourceVPCSubnetConfigTpl,
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
				Check: checkDataSourceVPCSubnet(
					fmt.Sprintf("data.oxide_vpc_subnet.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceVPCSubnet(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "description", "The default subnet for default"),
		resource.TestCheckResourceAttr(dataName, "name", "default"),
		resource.TestCheckResourceAttrSet(dataName, "ipv4_block"),
		resource.TestCheckResourceAttrSet(dataName, "ipv6_block"),
		resource.TestCheckResourceAttrSet(dataName, "vpc_id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}
