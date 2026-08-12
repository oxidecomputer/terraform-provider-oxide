// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippool_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/shared"
)

func TestAccSystemResourceIPPool_full(t *testing.T) {
	resourceName := "oxide_system_ip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig,
				Check:  checkResource(resourceName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testResourceUpdateConfig,
				Check:  checkResourceUpdate(resourceName),
			},
		},
	})
}

var testResourceConfig = `
resource "oxide_system_ip_pool" "test" {
	name = "terraform-acc-system-ippool"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

func checkResource(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "description", ""),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-system-ippool"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

var testResourceUpdateConfig = `
resource "oxide_system_ip_pool" "test" {
	description = "a new description for ip_pool"
	name        = "terraform-acc-system-ippool-new"
}
`

func checkResourceUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(
			resourceName,
			"description",
			"a new description for ip_pool",
		),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-system-ippool-new"),
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
		if rs.Type != "oxide_system_ip_pool" {
			continue
		}

		_, err := client.SystemIpPoolView(
			context.Background(),
			oxide.SystemIpPoolViewParams{Pool: oxide.NameOrId(rs.Primary.ID)},
		)
		if err == nil || !shared.Is404(err) {
			return fmt.Errorf("IP pool (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
