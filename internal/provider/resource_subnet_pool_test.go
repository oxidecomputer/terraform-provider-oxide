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

func TestAccResourceSubnetPool_full(t *testing.T) {
	resourceName := "oxide_subnet_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSubnetPoolDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testResourceSubnetPoolConfig,
				Check:  checkResourceSubnetPool(resourceName),
			},
			// Update: change name/description
			{
				Config: testResourceSubnetPoolUpdateConfig,
				Check:  checkResourceSubnetPoolUpdate(resourceName),
			},
			// Import
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})
}

var testResourceSubnetPoolConfig = `
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool"
	description = "a test subnet pool"
	ip_version  = "v4"
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

func checkResourceSubnetPool(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-subnet-pool"),
		resource.TestCheckResourceAttr(resourceName, "description", "a test subnet pool"),
		resource.TestCheckResourceAttr(resourceName, "ip_version", "v4"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

var testResourceSubnetPoolUpdateConfig = `
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-new"
	description = "an updated subnet pool"
	ip_version  = "v4"
}
`

func checkResourceSubnetPoolUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", "terraform-acc-subnet-pool-new"),
		resource.TestCheckResourceAttr(resourceName, "description", "an updated subnet pool"),
		resource.TestCheckResourceAttr(resourceName, "ip_version", "v4"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func TestAccResourceSubnetPool_disappears(t *testing.T) {
	resourceName := "oxide_subnet_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccSubnetPoolDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceSubnetPoolDisappearsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					// Delete the pool outside of Terraform
					testAccSubnetPoolDisappears(resourceName),
				),
				// Expect Terraform to detect the pool is gone and plan to recreate
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccSubnetPoolDisappears deletes the pool via the API to simulate
// out-of-band deletion, testing that the Read function properly removes it from state.
func testAccSubnetPoolDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := newTestClient()
		if err != nil {
			return err
		}

		return client.SystemSubnetPoolDelete(
			context.Background(),
			oxide.SystemSubnetPoolDeleteParams{Pool: oxide.NameOrId(rs.Primary.ID)},
		)
	}
}

var testResourceSubnetPoolDisappearsConfig = `
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-disappears"
	description = "a test subnet pool for disappears test"
	ip_version  = "v4"
}
`

func testAccSubnetPoolDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_subnet_pool" {
			continue
		}

		ctx := context.Background()

		res, err := client.SubnetPoolView(
			ctx,
			oxide.SubnetPoolViewParams{Pool: oxide.NameOrId(rs.Primary.ID)},
		)
		if err != nil && is404(err) {
			continue
		}
		if err == nil {
			return fmt.Errorf("subnet_pool (%v) still exists", res.Name)
		}
		return err
	}

	return nil
}
