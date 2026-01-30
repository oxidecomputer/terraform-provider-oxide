// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

type resourceFloatingIPConfig struct {
	BlockName        string
	SupportBlockName string
	FloatingIPName   string
}

var resourceFloatingIPConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_floating_ip" "{{.BlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	name        = "{{.FloatingIPName}}"
	description = "Floating IP."
	ip_version  = "v4"

	timeouts = {
		create = "1m"
		read   = "2m"
		update = "3m"
		delete = "4m"
	}
}
`

var resourceFloatingIPUpdateConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_floating_ip" "{{.BlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	name        = "{{.FloatingIPName}}"
	description = "Floating IP (updated)."
	ip_version  = "v4"
}
`

var resourceFloatingIPv6ConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_floating_ip" "{{.BlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	name        = "{{.FloatingIPName}}"
	description = "Floating IP (v6)."
	ip_version  = "v6"
}
`

var resourceFloatingIPNonDefaultPoolConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

data "oxide_ip_pool" "{{.SupportBlockName}}" {
	name = "non-default"
}

resource "oxide_floating_ip" "{{.BlockName}}" {
	project_id  = data.oxide_project.{{.SupportBlockName}}.id
	name        = "{{.FloatingIPName}}"
	description = "Floating IP (non-default)."
	ip_pool_id  = data.oxide_ip_pool.{{.SupportBlockName}}.id
}
`

func TestAccCloudResourceFloatingIP_full(t *testing.T) {
	floatingIPName := newResourceName()
	blockName := newBlockName("floating_ip")
	resourceName := fmt.Sprintf("oxide_floating_ip.%s", blockName)
	supportBlockName := newBlockName("support")
	config, err := parsedAccConfig(
		resourceFloatingIPConfig{
			BlockName:        blockName,
			FloatingIPName:   floatingIPName,
			SupportBlockName: supportBlockName,
		},
		resourceFloatingIPConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	floatingIPNameUpdated := floatingIPName + "-updated"
	configUpdate, err := parsedAccConfig(
		resourceFloatingIPConfig{
			BlockName:        blockName,
			FloatingIPName:   floatingIPNameUpdated,
			SupportBlockName: supportBlockName,
		},
		resourceFloatingIPUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configV6, err := parsedAccConfig(
		resourceFloatingIPConfig{
			BlockName:        blockName,
			FloatingIPName:   floatingIPName,
			SupportBlockName: supportBlockName,
		},
		resourceFloatingIPv6ConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configNonDefaultPool, err := parsedAccConfig(
		resourceFloatingIPConfig{
			BlockName:        blockName,
			FloatingIPName:   floatingIPName,
			SupportBlockName: supportBlockName,
		},
		resourceFloatingIPNonDefaultPoolConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccFloatingIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceFloatingIP(resourceName, floatingIPName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceFloatingIPUpdate(resourceName, floatingIPNameUpdated),
			},
			{
				Config: configV6,
				Check:  checkResourceFloatingIPV6(resourceName, floatingIPName),
			},
			{
				Config: configNonDefaultPool,
				Check:  checkResourceFloatingIPNonDefaultPool(resourceName, floatingIPName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					// Expected to be missing.
					"ip_version",
				},
			},
		},
	})
}

func checkResourceFloatingIP(resourceName, floatingIPName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", floatingIPName),
		resource.TestCheckResourceAttr(resourceName, "description", "Floating IP."),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrWith(resourceName, "ip", testAccFloatingIPIsV4),
		resource.TestCheckResourceAttrSet(resourceName, "ip_pool_id"),
		resource.TestCheckResourceAttr(resourceName, "ip_version", "v4"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "4m"),
	}...)
}

func checkResourceFloatingIPUpdate(resourceName, floatingIPName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", floatingIPName),
		resource.TestCheckResourceAttr(resourceName, "description", "Floating IP (updated)."),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrWith(resourceName, "ip", testAccFloatingIPIsV4),
		resource.TestCheckResourceAttrSet(resourceName, "ip_pool_id"),
		resource.TestCheckResourceAttr(resourceName, "ip_version", "v4"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceFloatingIPV6(resourceName, floatingIPName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", floatingIPName),
		resource.TestCheckResourceAttr(resourceName, "description", "Floating IP (v6)."),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrWith(resourceName, "ip", testAccFloatingIPIsV6),
		resource.TestCheckResourceAttrSet(resourceName, "ip_pool_id"),
		resource.TestCheckResourceAttr(resourceName, "ip_version", "v6"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceFloatingIPNonDefaultPool(
	resourceName, floatingIPName string,
) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", floatingIPName),
		resource.TestCheckResourceAttr(resourceName, "description", "Floating IP (non-default)."),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrWith(resourceName, "ip", testAccFloatingIPIsV4),
		resource.TestCheckResourceAttrSet(resourceName, "ip_pool_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccFloatingIPDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_floating_ip" {
			continue
		}

		params := oxide.FloatingIpViewParams{
			FloatingIp: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		res, err := client.FloatingIpView(ctx, params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("floating ip (%v) still exists", &res.Name)
	}

	return nil
}

func testAccFloatingIPIsV4(value string) error {
	if !isIPv4(value) {
		return fmt.Errorf("expected %s to be IPv4", value)
	}
	return nil
}

func testAccFloatingIPIsV6(value string) error {
	if !isIPv6(value) {
		return fmt.Errorf("expected %s to be IPv6", value)
	}
	return nil
}
