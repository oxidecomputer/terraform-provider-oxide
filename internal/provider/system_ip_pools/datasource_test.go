// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippools_test

import (
	"fmt"
	"testing"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

type dataSourceConfig struct {
	BlockName        string
	SupportBlockName string
}

var dataSourceConfigTpl = `
data "oxide_system_ip_pools" "{{.BlockName}}" {
  timeouts = {
    read = "1m"
  }
}
`

func TestAccSiloDataSourceSystemIPPools_full(t *testing.T) {
	blockName := sharedtest.NewBlockName("datasource-ip-pool")
	config := sharedtest.ParsedAccConfig(t,
		dataSourceConfig{
			BlockName:        blockName,
			SupportBlockName: sharedtest.NewBlockName("support"),
		},
		dataSourceConfigTpl,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: checkDataSource(
					fmt.Sprintf("data.oxide_system_ip_pools.%s", blockName),
				),
				// The `oxide_system_ip_pools` data source returns results sorted by UUID which
				// means, depending on the Oxide environment the test runs against, the 0th
				// system IP pool may not have a description. We'll check the state directly for
				// this attribute.
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						fmt.Sprintf("data.oxide_system_ip_pools.%s", blockName),
						tfjsonpath.New("ip_pools").AtSliceIndex(0).
							AtMapKey("description"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func checkDataSource(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "ip_pools.0.id"),
		resource.TestCheckResourceAttrSet(dataName, "ip_pools.0.name"),
		resource.TestCheckResourceAttrSet(dataName, "ip_pools.0.time_created"),
		resource.TestCheckResourceAttrSet(dataName, "ip_pools.0.time_modified"),
	}...)
}
