// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ippool_test

import (
	"fmt"
	"testing"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceIPPoolConfig struct {
	BlockName        string
	SupportBlockName string
}

var dataSourceIPPoolConfigTpl = `
data "oxide_ip_pool" "{{.BlockName}}" {
  name = "default"
  timeouts = {
    read = "1m"
  }
}
`

func TestAccSiloDataSourceIPPool_full(t *testing.T) {
	blockName := sharedtest.NewBlockName("datasource-ip-pool")
	config, err := sharedtest.ParsedAccConfig(
		dataSourceIPPoolConfig{
			BlockName:        blockName,
			SupportBlockName: sharedtest.NewBlockName("support"),
		},
		dataSourceIPPoolConfigTpl,
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
				Check: checkDataSourceIPPool(
					fmt.Sprintf("data.oxide_ip_pool.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceIPPool(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "name"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "is_default"),
		resource.TestCheckResourceAttrSet(dataName, "description"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
	}...)
}
