package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAddressLot_full(t *testing.T) {
	resourceName := "oxide_address_lot.test"
	addressLotName := newResourceName()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testResourceAddressLotConfig(addressLotName),
				Check:  checkResourceAddressLot(resourceName, addressLotName),
			},
			{
				Config: testResourceAddressLotUpdateConfig(addressLotName),
				Check:  checkResourceAddressLotUpdate(resourceName, addressLotName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testResourceAddressLotConfig(name string) string {
	return fmt.Sprintf(`
resource "oxide_address_lot" "test" {
	description       = "a test address lot"
	name              = "%[1]s"
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
`, name)
}

func testResourceAddressLotUpdateConfig(name string) string {
	return fmt.Sprintf(`
resource "oxide_address_lot" "test" {
	description       = "a test address lot"
	name              = "%[1]s"
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
`, name)
}

func checkResourceAddressLot(resourceName string, addressLotName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test address lot"),
		resource.TestCheckResourceAttr(resourceName, "name", addressLotName),
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

func checkResourceAddressLotUpdate(
	resourceName string,
	addressLotName string,
) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test address lot"),
		resource.TestCheckResourceAttr(resourceName, "name", addressLotName),
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
