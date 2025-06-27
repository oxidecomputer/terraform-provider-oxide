// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

func TestAccSwitchPortSettings(t *testing.T) {
	resourceName := "oxide_switch_port_settings.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSwitchPortSettingsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceSwitchPortSettingsCreate,
				Check:  checkResourceSwitchPortSettings(resourceName),
			},
			{
				Config: testResourceSwitchPortSettingsUpdate,
				Check:  checkResourceSwitchPortSettingsUpdate(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceSwitchPortSettingsRemoveUpdate,
				Check:  checkResourceSwitchPortSettingsRemoveUpdate(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

var testResourceSwitchPortSettingsCreate = `
resource "oxide_switch_port_configuration" "test" {
  name        = "test-uplink"
  description = "test uplink configuration"
  port_config = "qsfp28x1"

  addresses = [
    {
      addresses = [
        {
          address = "172.20.250.10/24"
          address_lot = {
            name = "testing"
          }
        },
      ]
      link_name = "phy0"
    },
  ]

  # bgp_peers = []

  links = [
    {
      autoneg = false
      fec     = "none"
      mtu     = 1500
      name    = "phy0"
      speed   = "speed40_g"
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
          destination = "0.0.0.0/0"
          gateway     = "172.20.250.1"
        },
      ]
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

func checkResourceSwitchPortSettings(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", "test-uplink"),
		resource.TestCheckResourceAttr(resourceName, "description", "test uplink configuration"),
		resource.TestCheckResourceAttr(resourceName, "port_config", "qsfp28x1"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.0.address", "172.20.250.10/24"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.0.address_lot.name", "test-address-lot"),
		resource.TestCheckResourceAttr(resourceName, "links.0.autoneg", "false"),
		resource.TestCheckResourceAttr(resourceName, "links.0.fec", "none"),
		resource.TestCheckResourceAttr(resourceName, "links.0.mtu", "1500"),
		resource.TestCheckResourceAttr(resourceName, "links.0.name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "links.0.speed", "speed40_g"),
		resource.TestCheckResourceAttr(resourceName, "links.0.lldp.enabled", "false"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.destination", "0.0.0.0/0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.gateway", "172.20.250.1"),

		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

var testResourceSwitchPortSettingsUpdate = `
resource "oxide_switch_port_configuration" "test" {
  name        = "test-uplink-v2"
  description = "test uplink configuration with additional route"
  port_config = "qsfp28x1"

  addresses = [
    {
      addresses = [
        {
          address = "172.20.250.10/24"
          address_lot = {
            name = "test-address-lot"
          }
        },
      ]
      link_name = "phy0"
    },
  ]

  # bgp_peers = []

  links = [
    {
      autoneg = false
      fec     = "none"
      mtu     = 1500
      name    = "phy0"
      speed   = "speed40_g"
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
          destination = "0.0.0.0/0"
          gateway     = "172.20.250.1"
        },
        {
          destination = "1.1.1.1/32"
          gateway     = "172.20.250.1"
        },
      ]
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

func checkResourceSwitchPortSettingsUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", "test-uplink"),
		resource.TestCheckResourceAttr(resourceName, "description", "test uplink configuration"),
		resource.TestCheckResourceAttr(resourceName, "port_config", "qsfp28x1"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.0.address", "172.20.250.10/24"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.0.address_lot.name", "test-address-lot"),
		resource.TestCheckResourceAttr(resourceName, "links.0.autoneg", "false"),
		resource.TestCheckResourceAttr(resourceName, "links.0.fec", "none"),
		resource.TestCheckResourceAttr(resourceName, "links.0.mtu", "1500"),
		resource.TestCheckResourceAttr(resourceName, "links.0.name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "links.0.speed", "speed40_g"),
		resource.TestCheckResourceAttr(resourceName, "links.0.lldp.enabled", "false"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.destination", "0.0.0.0/0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.gateway", "172.20.250.1"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.1.destination", "1.1.1.1/0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.1.gateway", "172.20.250.1"),

		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

var testResourceSwitchPortSettingsRemoveUpdate = `
resource "oxide_switch_port_configuration" "test" {
  name        = "test-uplink-v3"
  description = "test uplink configuration with route removed"
  port_config = "qsfp28x1"

  addresses = [
    {
      addresses = [
        {
          address = "172.20.250.10/24"
          address_lot = {
            name = "test-address-lot"
          }
        },
      ]
      link_name = "phy0"
    },
  ]

  # bgp_peers = []

  links = [
    {
      autoneg = false
      fec     = "none"
      mtu     = 1500
      name    = "phy0"
      speed   = "speed40_g"
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
          destination = "0.0.0.0/0"
          gateway     = "172.20.250.1"
        },
      ]
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

func checkResourceSwitchPortSettingsRemoveUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", "test-uplink"),
		resource.TestCheckResourceAttr(resourceName, "description", "test uplink configuration"),
		resource.TestCheckResourceAttr(resourceName, "port_config", "qsfp28x1"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.0.address", "172.20.250.10/24"),
		resource.TestCheckResourceAttr(resourceName, "addresses.0.addresses.0.address_lot.name", "test-address-lot"),
		resource.TestCheckResourceAttr(resourceName, "links.0.autoneg", "false"),
		resource.TestCheckResourceAttr(resourceName, "links.0.fec", "none"),
		resource.TestCheckResourceAttr(resourceName, "links.0.mtu", "1500"),
		resource.TestCheckResourceAttr(resourceName, "links.0.name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "links.0.speed", "speed40_g"),
		resource.TestCheckResourceAttr(resourceName, "links.0.lldp.enabled", "false"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.link_name", "phy0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.destination", "0.0.0.0/0"),
		resource.TestCheckResourceAttr(resourceName, "routes.0.routes.0.gateway", "172.20.250.1"),

		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
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

		res, err := client.IpPoolView(
			ctx,
			oxide.IpPoolViewParams{Pool: "terraform-acc-myippool"},
		)
		if err == nil || !is404(err) {
			return fmt.Errorf("ip_pool (%v) still exists", &res.Name)
		}

		res2, err := client.IpPoolView(
			ctx,
			oxide.IpPoolViewParams{Pool: "terraform-acc-myippool2"},
		)
		if err != nil && is404(err) {
			continue
		}
		return fmt.Errorf("ip_pool (%v) still exists", &res2.Name)
	}

	return nil
}
