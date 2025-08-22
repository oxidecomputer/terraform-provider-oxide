// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testDataSourceAddressLotConfig = `
resource "oxide_address_lot" "test" {
	description       = "a test address lot"
	name              = "terraform-acc-my-address-lot"
	kind              = "infra"
	blocks = [
		{
			first_address = "172.0.1.1"
			last_address  = "172.0.1.10"
		},
	]
}

data "oxide_address_lot" "test" {
  name = oxide_address_lot.test.name
}
`

func TestAccDataSourceAddressLot_full(t *testing.T) {
	resourceName := "oxide_address_lot.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceAddressLotConfig,
				Check:  checkDataSourceAddressLot(resourceName),
			},
		},
	})
}

func checkDataSourceAddressLot(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "description", "a test address lot"),
		resource.TestCheckResourceAttr(dataName, "name", "terraform-acc-my-address-lot"),
		resource.TestCheckResourceAttrSet(dataName, "blocks.0.first_address"),
		resource.TestCheckResourceAttrSet(dataName, "blocks.0.last_address"),
		resource.TestCheckResourceAttr(dataName, "blocks.0.first_address", "172.0.1.1"),
		resource.TestCheckResourceAttr(dataName, "blocks.0.last_address", "172.0.1.10"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
	}...)
}
