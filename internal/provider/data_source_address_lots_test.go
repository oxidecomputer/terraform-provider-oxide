// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceNetworkingAddressLotConfig struct {
	BlockName string
}

var datasourceNetworkingAddressLotsConfigTpl = `
data "oxide_networking_address_lots" "{{.BlockName}}" {
	timeouts = {
	  read = "1m"
	}
}
`

func TestAccSiloDataSourceNetworkingAddressLots_full(t *testing.T) {
	blockName := newBlockName("datasource-networking-address-lots")
	config, err := parsedAccConfig(
		dataSourceNetworkingAddressLotConfig{
			BlockName: blockName,
		},
		datasourceNetworkingAddressLotsConfigTpl,
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
				Check: checkDataSourceNetworkingAddressLots(
					fmt.Sprintf("data.oxide_networking_address_lots.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceNetworkingAddressLots(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "address_lots.0.id"),
		resource.TestCheckResourceAttrSet(dataName, "address_lots.0.description"),
		resource.TestCheckResourceAttrSet(dataName, "address_lots.0.kind"),
		resource.TestCheckResourceAttrSet(dataName, "address_lots.0.name"),
	}...)
}
