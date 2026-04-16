// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package floatingip_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

type resourceConfig struct {
	BlockName        string
	SupportBlockName string
	FloatingIPName   string
}

var resourceConfigTpl = `
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

var resourceUpdateConfigTpl = `
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

var resourceV6ConfigTpl = `
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

var resourceNonDefaultPoolConfigTpl = `
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
	floatingIPName := sharedtest.NewResourceName()
	blockName := sharedtest.NewBlockName("floating_ip")
	resourceName := fmt.Sprintf("oxide_floating_ip.%s", blockName)
	supportBlockName := sharedtest.NewBlockName("support")
	config, err := sharedtest.ParsedAccConfig(
		resourceConfig{
			BlockName:        blockName,
			FloatingIPName:   floatingIPName,
			SupportBlockName: supportBlockName,
		},
		resourceConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	floatingIPNameUpdated := floatingIPName + "-updated"
	configUpdate, err := sharedtest.ParsedAccConfig(
		resourceConfig{
			BlockName:        blockName,
			FloatingIPName:   floatingIPNameUpdated,
			SupportBlockName: supportBlockName,
		},
		resourceUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configV6, err := sharedtest.ParsedAccConfig(
		resourceConfig{
			BlockName:        blockName,
			FloatingIPName:   floatingIPName,
			SupportBlockName: supportBlockName,
		},
		resourceV6ConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configNonDefaultPool, err := sharedtest.ParsedAccConfig(
		resourceConfig{
			BlockName:        blockName,
			FloatingIPName:   floatingIPName,
			SupportBlockName: supportBlockName,
		},
		resourceNonDefaultPoolConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResource(resourceName, floatingIPName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceUpdate(resourceName, floatingIPNameUpdated),
			},
			{
				Config: configV6,
				Check:  checkResourceV6(resourceName, floatingIPName),
			},
			{
				Config: configNonDefaultPool,
				Check:  checkResourceNonDefaultPool(resourceName, floatingIPName),
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

func checkResource(resourceName, floatingIPName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", floatingIPName),
		resource.TestCheckResourceAttr(resourceName, "description", "Floating IP."),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrWith(resourceName, "ip", testAccResourceIsV4),
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

func checkResourceUpdate(resourceName, floatingIPName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", floatingIPName),
		resource.TestCheckResourceAttr(resourceName, "description", "Floating IP (updated)."),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrWith(resourceName, "ip", testAccResourceIsV4),
		resource.TestCheckResourceAttrSet(resourceName, "ip_pool_id"),
		resource.TestCheckResourceAttr(resourceName, "ip_version", "v4"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceV6(resourceName, floatingIPName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", floatingIPName),
		resource.TestCheckResourceAttr(resourceName, "description", "Floating IP (v6)."),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrWith(resourceName, "ip", testAccResourceIsV6),
		resource.TestCheckResourceAttrSet(resourceName, "ip_pool_id"),
		resource.TestCheckResourceAttr(resourceName, "ip_version", "v6"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func checkResourceNonDefaultPool(
	resourceName, floatingIPName string,
) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", floatingIPName),
		resource.TestCheckResourceAttr(resourceName, "description", "Floating IP (non-default)."),
		resource.TestCheckResourceAttrSet(resourceName, "project_id"),
		resource.TestCheckResourceAttrWith(resourceName, "ip", testAccResourceIsV4),
		resource.TestCheckResourceAttrSet(resourceName, "ip_pool_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccResourceDestroy(s *terraform.State) error {
	client, err := sharedtest.NewTestClient()
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
		if err != nil && shared.Is404(err) {
			continue
		}

		return fmt.Errorf("floating ip (%v) still exists", &res.Name)
	}

	return nil
}

func testAccResourceIsV4(value string) error {
	if !shared.IsIPv4(value) {
		return fmt.Errorf("expected %s to be IPv4", value)
	}
	return nil
}

func testAccResourceIsV6(value string) error {
	if !shared.IsIPv6(value) {
		return fmt.Errorf("expected %s to be IPv6", value)
	}
	return nil
}
