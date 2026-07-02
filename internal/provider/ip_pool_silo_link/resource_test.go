// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package ippoolsilolink_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

func TestAccSiloResourceIPPoolSiloLink_full(t *testing.T) {
	// The IPv4 `default` IP pool already holds the default link for
	// test-suite-silo, and only one default IP pool is allowed per silo, so
	// this test links a non-default pool. Run in series to avoid contention
	// over the shared silo's default link.
	ipPoolName := sharedtest.NewResourceName()
	linkResourceName := "oxide_ip_pool_silo_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			// Create pool and link.
			{
				Config: testResourceConfig(ipPoolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(linkResourceName, "id"),
					// The ID is the composite ip_pool_id/silo_id.
					checkLinkIDComposite(linkResourceName),
					resource.TestCheckResourceAttr(linkResourceName, "is_default", "false"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.read", "1m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.create", "3m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.delete", "2m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.update", "4m"),
				),
			},
			// Import using the composite ID (format: ip_pool_id/silo_id).
			{
				ResourceName:            linkResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[linkResourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", linkResourceName)
					}
					return fmt.Sprintf(
						"%s/%s",
						rs.Primary.Attributes["ip_pool_id"],
						rs.Primary.Attributes["silo_id"],
					), nil
				},
			},
		},
	})
}

func TestAccSiloResourceIPPoolSiloLink_disappears(t *testing.T) {
	ipPoolName := sharedtest.NewResourceName()
	linkResourceName := "oxide_ip_pool_silo_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig(ipPoolName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(linkResourceName, "id"),
					// Delete the link out of band.
					testAccResourceDisappears(linkResourceName),
				),
				// Detect the link is gone.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testResourceConfig(ipPoolName string) string {
	return fmt.Sprintf(`
data "oxide_silo" "test" {
	name = "test-suite-silo"
}

resource "oxide_ip_pool" "test" {
	name        = %[1]q
	description = "a test ip_pool for silo link tests"
}

resource "oxide_ip_pool_silo_link" "test" {
	ip_pool_id = oxide_ip_pool.test.id
	silo_id    = data.oxide_silo.test.id
	is_default = false
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`, ipPoolName)
}

// checkLinkIDComposite asserts that the resource ID is the composite
// ip_pool_id/silo_id value.
func checkLinkIDComposite(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		want := fmt.Sprintf(
			"%s/%s",
			rs.Primary.Attributes["ip_pool_id"],
			rs.Primary.Attributes["silo_id"],
		)
		if got := rs.Primary.Attributes["id"]; got != want {
			return fmt.Errorf("expected id %q, got %q", want, got)
		}

		return nil
	}
}

func testAccResourceDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}

		params := oxide.SystemIpPoolSiloUnlinkParams{
			Pool: oxide.NameOrId(rs.Primary.Attributes["ip_pool_id"]),
			Silo: oxide.NameOrId(rs.Primary.Attributes["silo_id"]),
		}

		return client.SystemIpPoolSiloUnlink(context.Background(), params)
	}
}
