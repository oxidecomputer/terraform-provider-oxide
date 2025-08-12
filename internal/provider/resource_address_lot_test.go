package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAddressLot_full(t *testing.T) {
	resourceName := "oxide_address_lot.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testResourceAddressLotConfig,
				Check:  checkResourceAddressLot(resourceName),
			},
			{
				Config: testResourceAddressLotUpdateConfig,
				Check:  checkResourceAddressLotUpdate(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

var testResourceAddressLotConfig = `
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
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

var testResourceAddressLotUpdateConfig = `
resource "oxide_address_lot" "test" {
	description       = "a test address lot"
	name              = "terraform-acc-my-address-lot"
	kind              = "infra"
	blocks = [
		{
			first_address = "172.0.1.1"
			last_address  = "172.0.1.10"
		},
		{
			first_address = "172.0.10.1"
			last_address  = "172.0.10.10"
		},
	]
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

func checkResourceAddressLot(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test address lot"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-my-address-lot"),
		resource.TestCheckResourceAttrSet(resourceName, "blocks.0.first_address"),
		resource.TestCheckResourceAttrSet(resourceName, "blocks.0.last_address"),
		resource.TestCheckResourceAttr(resourceName, "blocks.0.first_address", "172.0.1.1"),
		resource.TestCheckResourceAttr(resourceName, "blocks.0.last_address", "172.0.1.10"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func checkResourceAddressLotUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test address lot"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-my-address-lot"),
		resource.TestCheckResourceAttrSet(resourceName, "blocks.0.first_address"),
		resource.TestCheckResourceAttrSet(resourceName, "blocks.0.last_address"),
		resource.TestCheckResourceAttrSet(resourceName, "blocks.1.first_address"),
		resource.TestCheckResourceAttrSet(resourceName, "blocks.1.last_address"),
		resource.TestCheckResourceAttr(resourceName, "blocks.0.first_address", "172.0.1.1"),
		resource.TestCheckResourceAttr(resourceName, "blocks.0.last_address", "172.0.1.10"),
		resource.TestCheckResourceAttr(resourceName, "blocks.1.first_address", "172.0.10.1"),
		resource.TestCheckResourceAttr(resourceName, "blocks.1.last_address", "172.0.10.10"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}
