// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

type resourceIPPoolSiloLinkConfig struct {
	BlockName        string
	SupportBlockName string
	IPPoolName       string
}

type resourceIPPoolSiloLinkConfigUpdate struct {
	BlockName         string
	BlockName2        string
	SupportBlockName  string
	SupportBlockName2 string
	IPPoolName        string
	IPPoolName2       string
}

// TODO: Change the silo ID when we have a silo datasource
var resourceIPPoolSiloLinkConfigTpl = `
resource "oxide_ip_pool" "{{.SupportBlockName}}" {
  description       = "a test ip_pool"
  name              = "{{.IPPoolName}}"
  ranges = [
    {
	  first_address = "172.20.15.234"
	  last_address  = "172.20.15.237"
	}
  ]
}

resource "oxide_ip_pool_silo_link" "{{.BlockName}}" {
  silo_id = "1fec2c21-cf22-40d8-9ebd-e5b57ebec80f"
  ip_pool_id = oxide_ip_pool.{{.SupportBlockName}}.id
  is_default = true
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "4m"
  }
}
`

var resourceIPPoolSiloLinkUpdateConfigTpl = `
resource "oxide_ip_pool" "{{.SupportBlockName}}" {
  description       = "a test ip_pool"
  name              = "{{.IPPoolName}}"
  ranges = [
    {
	  first_address = "172.20.15.234"
	  last_address  = "172.20.15.237"
	}
  ]
}

resource "oxide_ip_pool" "{{.SupportBlockName2}}" {
  description       = "a test ip_pool"
  name              = "{{.IPPoolName2}}"
  ranges = [
    {
	  first_address = "172.20.15.238"
	  last_address  = "172.20.15.240"
	}
  ]
}

resource "oxide_ip_pool_silo_link" "{{.BlockName}}" {
  silo_id = "1fec2c21-cf22-40d8-9ebd-e5b57ebec80f"
  ip_pool_id = oxide_ip_pool.{{.SupportBlockName}}.id
  is_default = false
}

resource "oxide_ip_pool_silo_link" "{{.BlockName2}}" {
  silo_id = "1fec2c21-cf22-40d8-9ebd-e5b57ebec80f"
  ip_pool_id = oxide_ip_pool.{{.SupportBlockName2}}.id
  is_default = true
}
`

func TestAccSiloResourceIPPoolSiloLink_full(t *testing.T) {
	t.Skip("skipping test until there is a silo datasource to retrieve the ID.")

	ipPoolName := newResourceName()
	blockName := newBlockName("ip_pool")
	blockName2 := newBlockName("ip_pool")
	supportBlockName := newBlockName("support")
	resourceName := fmt.Sprintf("oxide_ip_pool_silo_link.%s", blockName)
	resourceName2 := fmt.Sprintf("oxide_ip_pool_silo_link.%s", blockName)
	config, err := parsedAccConfig(
		resourceIPPoolSiloLinkConfig{
			BlockName:        blockName,
			SupportBlockName: supportBlockName,
			IPPoolName:       ipPoolName,
		},
		resourceIPPoolSiloLinkConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configUpdate, err := parsedAccConfig(
		resourceIPPoolSiloLinkConfigUpdate{
			BlockName:         blockName,
			IPPoolName:        ipPoolName,
			BlockName2:        blockName2,
			IPPoolName2:       newResourceName(),
			SupportBlockName:  supportBlockName,
			SupportBlockName2: newBlockName("support"),
		},
		resourceIPPoolSiloLinkUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccIPPoolSiloLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceIPPoolSiloLink(resourceName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceIPPoolSiloLinkUpdate(resourceName, resourceName2),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceIPPoolSiloLink(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "silo_id"),
		resource.TestCheckResourceAttrSet(resourceName, "ip_pool_id"),
		resource.TestCheckResourceAttr(resourceName, "is_default", "true"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func checkResourceIPPoolSiloLinkUpdate(resourceName, resourceName2 string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "silo_id"),
		resource.TestCheckResourceAttrSet(resourceName, "ip_pool_id"),
		resource.TestCheckResourceAttr(resourceName, "is_default", "false"),
		resource.TestCheckResourceAttrSet(resourceName2, "id"),
		resource.TestCheckResourceAttrSet(resourceName2, "silo_id"),
		resource.TestCheckResourceAttrSet(resourceName2, "ip_pool_id"),
		resource.TestCheckResourceAttr(resourceName2, "is_default", "true"),
	}...)
}

func testAccIPPoolSiloLinkDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_ip_pool" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		ipPoolID := rs.Primary.Attributes["ip_pool_id"]
		siloID := rs.Primary.Attributes["silo_id"]
		params := oxide.IpPoolSiloListParams{
			Pool:   oxide.NameOrId(oxide.NameOrId(ipPoolID)),
			Limit:  oxide.NewPointer(1000000000),
			SortBy: oxide.IdSortModeIdAscending,
		}

		links, err := client.IpPoolSiloList(ctx, params)
		if err != nil && is404(err) {
			continue
		}

		idx := slices.IndexFunc(
			links.Items,
			func(l oxide.IpPoolSiloLink) bool { return l.SiloId == siloID },
		)
		if idx >= 0 {
			return fmt.Errorf(
				"link between IP pool: '%v' and silo '%v' still exists",
				ipPoolID,
				siloID,
			)
		}
	}

	return nil
}
