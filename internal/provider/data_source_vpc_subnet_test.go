// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

type dataSourceVPCSubnetConfig struct {
	BlockName        string
	SupportBlockName string
}

var dataSourceVPCSubnetConfigTpl = `
data "oxide_vpc_subnet" "{{.BlockName}}" {
  project_name = "tf-acc-test"
  vpc_name     = "default"
  name         = "default"
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceVPCSubnet_full(t *testing.T) {
	blockName := sharedtest.NewBlockName("datasource-vpc-subnet")
	config, err := sharedtest.ParsedAccConfig(
		dataSourceVPCSubnetConfig{
			BlockName:        blockName,
			SupportBlockName: sharedtest.NewBlockName("support"),
		},
		dataSourceVPCSubnetConfigTpl,
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
