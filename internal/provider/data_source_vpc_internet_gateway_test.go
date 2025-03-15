// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceVPCInternetGatewayConfig struct {
	BlockName        string
	SupportBlockName string
}

var dataSourceVPCInternetGatewayConfigTpl = `
data "oxide_vpc_internet_gateway" "{{.BlockName}}" {
  project_name = "tf-acc-test"
  vpc_name     = "default"
  name         = "default"
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceVPCInternetGateway_full(t *testing.T) {
	blockName := newBlockName("datasource-vpc-router")
	config, err := parsedAccConfig(
		dataSourceVPCInternetGatewayConfig{
			BlockName:        blockName,
			SupportBlockName: newBlockName("support"),
		},
		dataSourceVPCInternetGatewayConfigTpl,
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
				Check: checkDataSourceVPCInternetGateway(
					fmt.Sprintf("data.oxide_vpc_internet_gateway.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceVPCInternetGateway(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "description", "Default VPC gateway"),
		resource.TestCheckResourceAttr(dataName, "name", "default"),
		resource.TestCheckResourceAttrSet(dataName, "vpc_id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}
