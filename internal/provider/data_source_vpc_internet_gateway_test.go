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
data "oxide_vpc" "test" {
  project_name = "tf-acc-test"
  name         = "default"
}

resource "oxide_vpc_internet_gateway" "{{.BlockName}}" {
  vpc_id      = data.oxide_vpc.test.id
  name        = "test"
  description = "test description"
}

data "oxide_vpc_internet_gateway" "{{.BlockName}}" {
  project_name = "tf-acc-test"
  vpc_name     = "default"
  name         = oxide_vpc_internet_gateway.{{.BlockName}}.name
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
		resource.TestCheckResourceAttr(dataName, "description", "test description"),
		resource.TestCheckResourceAttr(dataName, "name", "test"),
		resource.TestCheckResourceAttrSet(dataName, "vpc_id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}
