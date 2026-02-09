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

// TestAccSiloResourceSwitchPortSettings_full tests whether Terraform
// can create `oxide_switch_port_settings` resources.
//
// This test overwrites Oxide switch port settings which can cause result in
// network connectivity issues if not careful. Read on to learn how to run this
// test safely.
//
// This test assumes it's running against a simulated Omicron deployment
// started by `cargo xtask omicron-dev run-all`. A simulated deployment
// contains an address lot named `initial-infra` with an address of `0.0.0.0`.
//
// Since the `oxide_switch_port_settings` resource completely overwrites all
// switch port settings it is still considered unsafe to run this test against
// anything but a simulated Omicron deployment. Set TF_ACC_SIM=1 to run.
func TestAccSiloResourceSwitchPortSettings_full(t *testing.T) {
	if os.Getenv("TF_ACC_SIM") == "" {
		t.Skip("Skipping destructive test; set TF_ACC_SIM=1 to run.")
	}

	type resourceSwitchPortSettingsConfig struct {
		BlockName              string
		SwitchPortSettingsName string
	}

	initialConfigTmpl := `
data "oxide_address_lot" "initial_infra" {
  name = "initial-infra"
}

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
          address_lot_id = data.oxide_address_lot.initial_infra.id
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
    {
      link_name = "phy1"
      autoneg   = false
      mtu       = 1500
      speed     = "speed10_g"
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
	updateConfigTmpl := `
data "oxide_address_lot" "initial_infra" {
  name = "initial-infra"
}

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
          address_lot_id = data.oxide_address_lot.initial_infra.id
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
    {
      link_name = "phy1"
      autoneg   = false
      mtu       = 1500
      speed     = "speed10_g"
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
	switchPortSettingsName := newResourceName()
	blockName := newBlockName("switch-port-settings")
	resourceName := fmt.Sprintf("oxide_switch_port_settings.%s", blockName)

	initialConfig, err := parsedAccConfig(
		resourceSwitchPortSettingsConfig{
			BlockName:              blockName,
			SwitchPortSettingsName: switchPortSettingsName,
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
		},
		updateConfigTmpl,
	)
	if err != nil {
		t.Errorf("error parsing update config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
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
		resource.TestCheckResourceAttr(
			resourceName,
			"description",
			"Terraform acceptance testing.",
		),
		resource.TestCheckResourceAttr(resourceName, "port_config.geometry", "qsfp28x1"),
		resource.TestCheckNoResourceAttr(resourceName, "bgp_peers"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),

		resource.TestCheckResourceAttr(resourceName, "addresses.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.#", "1"),
		resource.TestCheckResourceAttr(
			resourceName,
			"addresses.0.addresses.0.address",
			"0.0.0.0/0",
		),
		resource.TestCheckResourceAttrSet(resourceName, "addresses.0.addresses.0.address_lot_id"),
		resource.TestCheckNoResourceAttr(resourceName, "addresses.0.addresses.0.vlan_id"),

		resource.TestCheckResourceAttr(resourceName, "links.#", "2"),
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
		resource.TestCheckResourceAttr(resourceName, "links.1.link_name", "phy1"),
		resource.TestCheckResourceAttr(resourceName, "links.1.autoneg", "false"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.fec"),
		resource.TestCheckResourceAttr(resourceName, "links.1.mtu", "1500"),
		resource.TestCheckResourceAttr(resourceName, "links.1.speed", "speed10_g"),
		resource.TestCheckResourceAttr(resourceName, "links.1.lldp.%", "7"),
		resource.TestCheckResourceAttr(resourceName, "links.1.lldp.enabled", "true"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.chassis_id"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.link_description"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.link_name"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.management_ip"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.system_description"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.system_name"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.tx_eq"),

		resource.TestCheckResourceAttr(resourceName, "routes.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.dst", "0.0.0.0/0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.gw", "0.0.0.0"),
		resource.TestCheckNoResourceAttr(resourceName, "routes.0.routes.0.rib_priority"),
	}...)
}

func checkResourceSwitchPortSettingsUpdate(
	resourceName string,
	name string,
) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", name),
		resource.TestCheckResourceAttr(
			resourceName,
			"description",
			"Terraform acceptance testing (updated).",
		),
		resource.TestCheckResourceAttr(resourceName, "port_config.geometry", "qsfp28x1"),
		resource.TestCheckNoResourceAttr(resourceName, "bgp_peers"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),

		resource.TestCheckResourceAttr(resourceName, "addresses.#", "1"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.#", "1"),
		resource.TestCheckResourceAttr(
			resourceName,
			"addresses.0.addresses.0.address",
			"0.0.0.0/0",
		),
		resource.TestCheckResourceAttrSet(resourceName, "addresses.0.addresses.0.address_lot_id"),
		resource.TestCheckNoResourceAttr(resourceName, "addresses.0.addresses.0.vlan_id"),

		resource.TestCheckResourceAttr(resourceName, "links.#", "2"),
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
		resource.TestCheckResourceAttr(resourceName, "links.1.link_name", "phy1"),
		resource.TestCheckResourceAttr(resourceName, "links.1.autoneg", "false"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.fec"),
		resource.TestCheckResourceAttr(resourceName, "links.1.mtu", "1500"),
		resource.TestCheckResourceAttr(resourceName, "links.1.speed", "speed10_g"),
		resource.TestCheckResourceAttr(resourceName, "links.1.lldp.%", "7"),
		resource.TestCheckResourceAttr(resourceName, "links.1.lldp.enabled", "false"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.chassis_id"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.link_description"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.link_name"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.management_ip"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.system_description"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.lldp.system_name"),
		resource.TestCheckNoResourceAttr(resourceName, "links.1.tx_eq"),

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
