// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceVPCConfig struct {
	BlockName        string
	SupportBlockName string
}

var dataSourceVPCConfigTpl = `
data "oxide_vpc" "{{.BlockName}}" {
  project_name = "tf-acc-test"
  name         = "default"
  timeouts = {
    read = "1m"
  }
}
`

func TestAccDataSourceVPC_full(t *testing.T) {
	blockName := newBlockName("datasource-vpc")
	config, err := parsedAccConfig(
		dataSourceVPCConfig{
			BlockName:        blockName,
			SupportBlockName: newBlockName("support"),
		},
		dataSourceVPCConfigTpl,
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
				Check: checkDataSourceVPC(
					fmt.Sprintf("data.oxide_vpc.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceVPC(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "description", "Default VPC"),
		resource.TestCheckResourceAttr(dataName, "name", "default"),
		resource.TestCheckResourceAttr(dataName, "dns_name", "default"),
		resource.TestCheckResourceAttrSet(dataName, "ipv6_prefix"),
		resource.TestCheckResourceAttrSet(dataName, "project_id"),
		resource.TestCheckResourceAttrSet(dataName, "system_router_id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}
