// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceSystemIPPoolsConfig struct {
	BlockName        string
	SupportBlockName string
}

var dataSourceSystemIPPoolsConfigTpl = `
data "oxide_system_ip_pools" "{{.BlockName}}" {
  timeouts = {
    read = "1m"
  }
}
`

func TestAccSiloDataSourceSystemIPPools_full(t *testing.T) {
	blockName := newBlockName("datasource-ip-pool")
	config, err := parsedAccConfig(
		dataSourceSystemIPPoolsConfig{
			BlockName:        blockName,
			SupportBlockName: newBlockName("support"),
		},
		dataSourceSystemIPPoolsConfigTpl,
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
				Check: checkDataSourceSystemIPPools(
					fmt.Sprintf("data.oxide_system_ip_pools.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceSystemIPPools(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "ip_pools.0.id"),
		resource.TestCheckResourceAttrSet(dataName, "ip_pools.0.name"),
		resource.TestCheckResourceAttrSet(dataName, "ip_pools.0.description"),
		resource.TestCheckResourceAttrSet(dataName, "ip_pools.0.time_created"),
		resource.TestCheckResourceAttrSet(dataName, "ip_pools.0.time_modified"),
	}...)
}
