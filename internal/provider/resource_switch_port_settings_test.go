// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// TestAccCloudResourceSwitchPortSettings_full tests whether Terraform
// can create `oxide_switch_port_settings` resources.
//
// This test overwrites Oxide switch port settings which can cause result in
// network connectivity issues if not careful. Read on to learn how to run this
// test safely.
//
// This test assumes it's running against a simulated Omicron deployment
// started by `cargo xtask omicron-dev run-all`. A simulated deployment
// contains an address lot named `initial-infra` with an address of `0.0.0.0`.
// The ID of this address lot is needed to run this test and is expected to
// be passed to this test via the `OXIDE_TEST_ADDRESS_LOT_ID` environment
// variable. You can query the address lots in your simulated deployment at
// `/v1/system/networking/address-lot` to retrieve the address lot ID.
//
// With all that in mind, you can run this test like so.
//
// TEST_ACC_NAME=TestAccCloudResourceSwitchPortSettings_full OXIDE_TEST_ADDRESS_LOT_ID=<CHANGE_ME> make testacc
//
// A future iteration of this test can make use of the currently
// non-existent `oxide_address_lot` data source to remove the need for
// `OXIDE_TEST_ADDRESS_LOT_ID`. However, since the `oxide_switch_port_settings`
// resource completely overwrites all switch port settings it is still
// considered unsafe to run this test against anything but a simulated Omicron
// deployment.
func TestAccCloudResourceSwitchPortSettings_full(t *testing.T) {
	type resourceSwitchPortSettingsConfig struct {
		BlockName              string
		SwitchPortSettingsName string
		AddressLotID           string
	}

	initialConfigTmpl := `
resource "oxide_switch_port_settings" "{{.BlockName}}" {
  name        = "{{.SwitchPortSettingsName}}"
  description = "Terraform acceptance testing."

  port_config = {
    geometry = "qsfp28x1"
  }

  addresses = [
    {
      link_name = "phy0"
      addresses = [
        {
          address        = "0.0.0.0/0"
          address_lot_id = "{{.AddressLotID}}"
        },
      ]
    },
  ]

  links = [
    {
      link_name = "phy0"
      autoneg   = false
      mtu       = 1500
      speed     = "speed1_g"
      lldp = {
        enabled = false
      }
    },
  ]

  routes = [
    {
      link_name = "phy0"
      routes = [
        {
          dst = "0.0.0.0/0"
          gw  = "0.0.0.0"
        },
      ]
    },
  ]
}
`
	updateConfigTmpl := `
resource "oxide_switch_port_settings" "{{.BlockName}}" {
  name        = "{{.SwitchPortSettingsName}}"
  description = "Terraform acceptance testing (updated)."

  port_config = {
    geometry = "qsfp28x1"
  }

  addresses = [
    {
      link_name = "phy0"
      addresses = [
        {
          address        = "0.0.0.0/0"
          address_lot_id = "{{.AddressLotID}}"
        },
      ]
    },
  ]

  links = [
    {
      link_name = "phy0"
      autoneg   = false
      mtu       = 1500
      speed     = "speed1_g"
      lldp = {
        enabled = true
      }
    },
  ]

  routes = [
    {
      link_name = "phy0"
      routes = [
        {
          dst = "0.0.0.0/0"
          gw  = "0.0.0.0"
        },
      ]
    },
  ]
}
`
	// See the doc comment at [TestAccCloudResourceSwitchPortSettings_full].
	addressLotID := os.Getenv("OXIDE_TEST_ADDRESS_LOT_ID")

	switchPortSettingsName := newResourceName()
	blockName := newBlockName("switch-port-settings")
	resourceName := fmt.Sprintf("oxide_switch_port_settings.%s", blockName)

	initialConfig, err := parsedAccConfig(
		resourceSwitchPortSettingsConfig{
			BlockName:              blockName,
			SwitchPortSettingsName: switchPortSettingsName,
			AddressLotID:           addressLotID,
		},
		initialConfigTmpl,
	)
	if err != nil {
		t.Errorf("error parsing initial config template data: %e", err)
	}

	updateConfig, err := parsedAccConfig(
		resourceSwitchPortSettingsConfig{
			BlockName:              blockName,
			SwitchPortSettingsName: switchPortSettingsName,
			AddressLotID:           addressLotID,
		},
		updateConfigTmpl,
	)
	if err != nil {
		t.Errorf("error parsing update config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			// See the doc comment at [TestAccCloudResourceSwitchPortSettings_full].
			if addressLotID == "" {
				t.Skip("Skipping test. Export OXIDE_TEST_ADDRESS_LOT_ID to run.")
			}
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSwitchPortSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: initialConfig,
				Check:  checkResourceSwitchPortSettings(resourceName, switchPortSettingsName),
			},
			{
				Config: updateConfig,
				Check:  checkResourceSwitchPortSettingsUpdate(resourceName, switchPortSettingsName),
			},
			{
				Config: initialConfig,
				Check:  checkResourceSwitchPortSettings(resourceName, switchPortSettingsName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceSwitchPortSettings(resourceName string, name string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", "Terraform acceptance testing."),
		resource.TestCheckResourceAttr(resourceName, "port_config.geometry", "qsfp28x1"),
		resource.TestCheckNoResourceAttr(resourceName, "bgp_peers"),
		resource.TestCheckNoResourceAttr(resourceName, "groups"),
		resource.TestCheckNoResourceAttr(resourceName, "interfaces"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),

		resource.TestCheckResourceAttr(resourceName, "addresses.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.0.address", "0.0.0.0/0"),
		resource.TestCheckResourceAttrSet(resourceName, "addresses.0.addresses.0.address_lot_id"),
		resource.TestCheckNoResourceAttr(resourceName, "addresses.0.addresses.0.vlan_id"),

		resource.TestCheckResourceAttr(resourceName, "links.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "links.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "links.0.autoneg", "false"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.fec"),
		resource.TestCheckResourceAttr(resourceName, "links.0.mtu", "1500"),
		resource.TestCheckResourceAttr(resourceName, "links.0.speed", "speed1_g"),
		resource.TestCheckResourceAttr(resourceName, "links.0.lldp.%", "7"),
		resource.TestCheckResourceAttr(resourceName, "links.0.lldp.enabled", "false"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.chassis_id"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.link_description"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.link_name"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.management_ip"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.system_description"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.system_name"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.tx_eq"),

		resource.TestCheckResourceAttr(resourceName, "routes.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.dst", "0.0.0.0/0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.gw", "0.0.0.0"),
		resource.TestCheckNoResourceAttr(resourceName, "routes.0.routes.0.rib_priority"),
	}...)
}

func checkResourceSwitchPortSettingsUpdate(resourceName string, name string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(resourceName, "description", "Terraform acceptance testing (updated)."),
		resource.TestCheckResourceAttr(resourceName, "port_config.geometry", "qsfp28x1"),
		resource.TestCheckNoResourceAttr(resourceName, "bgp_peers"),
		resource.TestCheckNoResourceAttr(resourceName, "groups"),
		resource.TestCheckNoResourceAttr(resourceName, "interfaces"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),

		resource.TestCheckResourceAttr(resourceName, "addresses.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.0.address", "0.0.0.0/0"),
		resource.TestCheckResourceAttrSet(resourceName, "addresses.0.addresses.0.address_lot_id"),
		resource.TestCheckNoResourceAttr(resourceName, "addresses.0.addresses.0.vlan_id"),

		resource.TestCheckResourceAttr(resourceName, "links.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "links.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "links.0.autoneg", "false"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.fec"),
		resource.TestCheckResourceAttr(resourceName, "links.0.mtu", "1500"),
		resource.TestCheckResourceAttr(resourceName, "links.0.speed", "speed1_g"),
		resource.TestCheckResourceAttr(resourceName, "links.0.lldp.%", "7"),
		resource.TestCheckResourceAttr(resourceName, "links.0.lldp.enabled", "true"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.chassis_id"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.link_description"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.link_name"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.management_ip"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.system_description"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.lldp.system_name"),
		resource.TestCheckNoResourceAttr(resourceName, "links.0.tx_eq"),

		resource.TestCheckResourceAttr(resourceName, "routes.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.dst", "0.0.0.0/0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.gw", "0.0.0.0"),
		resource.TestCheckNoResourceAttr(resourceName, "routes.0.routes.0.rib_priority"),
	}...)
}

func testAccSwitchPortSettingsDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_switch_port_settings" {
			continue
		}

		ctx := context.Background()

		res, err := client.NetworkingSwitchPortSettingsView(
			ctx,
			oxide.NetworkingSwitchPortSettingsViewParams{
				Port: oxide.NameOrId(rs.Primary.Attributes["id"]),
			},
		)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("switch_port_settings (%v) still exists", &res.Name)
	}

	return nil
}
