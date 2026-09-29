// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ippools_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccSiloDataSourceIPPools_full(t *testing.T) {
	const dataSourceName = "data.oxide_ip_pools.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
data "oxide_ip_pools" "test" {
  timeouts = {
    read = "1m"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "timeouts.read", "1m"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ip_pools.0.id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ip_pools.0.ip_version"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ip_pools.0.is_default"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ip_pools.0.name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ip_pools.0.pool_type"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ip_pools.0.time_created"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ip_pools.0.time_modified"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						dataSourceName,
						tfjsonpath.New("ip_pools").AtSliceIndex(0).AtMapKey("description"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}
