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

func TestAccResourceSubnetPoolSiloLink_full(t *testing.T) {
	// Only one subnet pool can set `is_default` for a given silo. To ensure that this tests doesn't
	// conflict with other tests that set a default subnet pool, don't run it in parallel.
	poolResourceName := "oxide_subnet_pool.test"
	linkResourceName := "oxide_subnet_pool_silo_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			// Create pool and link.
			{
				Config: testResourceSubnetPoolSiloLinkConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(poolResourceName, "id"),
					resource.TestCheckResourceAttrSet(linkResourceName, "id"),
					resource.TestCheckResourceAttrPair(
						linkResourceName,
						"subnet_pool_id",
						poolResourceName,
						"id",
					),
					resource.TestCheckResourceAttr(linkResourceName, "is_default", "true"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.read", "1m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.create", "3m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.delete", "2m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.update", "4m"),
				),
			},
			// Update in place.
			{
				Config: testResourceSubnetPoolSiloLinkUpdateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(linkResourceName, "id"),
					resource.TestCheckResourceAttr(linkResourceName, "is_default", "false"),
				),
			},
			// Remove the link, keeping the pool so we can verify deletion.
			{
				Config: testResourceSubnetPoolSiloLinkPoolOnlyConfig,
				Check:  testAccLinksDestroyed(poolResourceName),
			},
		},
	})
}

var testResourceSubnetPoolSiloLinkConfig = `
data "oxide_silo" "test" {
	name = "test-suite-silo"
}

resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-silo-link-test"
	description = "a test subnet pool for silo link tests"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_silo_link" "test" {
	subnet_pool_id = oxide_subnet_pool.test.id
	silo_id        = data.oxide_silo.test.id
	is_default     = true
	timeouts = {
		read   = "1m"
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`

var testResourceSubnetPoolSiloLinkUpdateConfig = `
data "oxide_silo" "test" {
	name = "test-suite-silo"
}

resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-silo-link-test"
	description = "a test subnet pool for silo link tests"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_silo_link" "test" {
	subnet_pool_id = oxide_subnet_pool.test.id
	silo_id        = data.oxide_silo.test.id
	is_default     = false
}
`

var testResourceSubnetPoolSiloLinkPoolOnlyConfig = `
resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-silo-link-test"
	description = "a test subnet pool for silo link tests"
	ip_version  = "v4"
}
`

func TestAccResourceSubnetPoolSiloLink_disappears(t *testing.T) {
	poolResourceName := "oxide_subnet_pool.test"
	linkResourceName := "oxide_subnet_pool_silo_link.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testResourceSubnetPoolSiloLinkDisappearsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(poolResourceName, "id"),
					resource.TestCheckResourceAttrSet(linkResourceName, "id"),
					// Delete the link out of band.
					testAccSubnetPoolSiloLinkDisappears(linkResourceName),
				),
				// Detect the link is gone.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccSubnetPoolSiloLinkDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := newTestClient()
		if err != nil {
			return err
		}

		params := oxide.SystemSubnetPoolSiloUnlinkParams{
			Pool: oxide.NameOrId(rs.Primary.Attributes["subnet_pool_id"]),
			Silo: oxide.NameOrId(rs.Primary.Attributes["silo_id"]),
		}

		return client.SystemSubnetPoolSiloUnlink(context.Background(), params)
	}
}

var testResourceSubnetPoolSiloLinkDisappearsConfig = `
data "oxide_silo" "test" {
	name = "test-suite-silo"
}

resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-silo-link-disappears-test"
	description = "a test subnet pool for disappears test"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_silo_link" "test" {
	subnet_pool_id = oxide_subnet_pool.test.id
	silo_id        = data.oxide_silo.test.id
	is_default     = false
}
`

func TestAccResourceSubnetPoolSiloLink_multiSiloImport(t *testing.T) {
	siloDNSName := testAccSiloDNSName()

	link1ResourceName := "oxide_subnet_pool_silo_link.link1"
	link2ResourceName := "oxide_subnet_pool_silo_link.link2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		Steps: []resource.TestStep{
			// Create pool linked to two silos.
			{
				Config: testResourceSubnetPoolSiloLinkMultiSiloConfig(siloDNSName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(link1ResourceName, "id"),
					resource.TestCheckResourceAttrSet(link2ResourceName, "id"),
					resource.TestCheckResourceAttr(link1ResourceName, "is_default", "false"),
					resource.TestCheckResourceAttr(link2ResourceName, "is_default", "false"),
				),
			},
			// Import first link (format: subnet_pool_id/silo_id).
			{
				ResourceName:            link1ResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[link1ResourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", link1ResourceName)
					}
					return fmt.Sprintf(
						"%s/%s",
						rs.Primary.Attributes["subnet_pool_id"],
						rs.Primary.Attributes["silo_id"],
					), nil
				},
			},
			// Import second link.
			{
				ResourceName:            link2ResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[link2ResourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", link2ResourceName)
					}
					return fmt.Sprintf(
						"%s/%s",
						rs.Primary.Attributes["subnet_pool_id"],
						rs.Primary.Attributes["silo_id"],
					), nil
				},
			},
		},
	})
}

func testResourceSubnetPoolSiloLinkMultiSiloConfig(siloDNSName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "self_signed" {
	algorithm = "RSA"
	rsa_bits  = 2048
}

resource "tls_self_signed_cert" "self_signed" {
	private_key_pem       = tls_private_key.self_signed.private_key_pem
	validity_period_hours = 8760

	subject {
		common_name  = "%s"
		organization = "Oxide Computer Company"
	}

	dns_names = ["%s"]

	allowed_uses = [
		"key_encipherment",
		"digital_signature",
		"server_auth",
	]
}

data "oxide_silo" "silo1" {
	name = "test-suite-silo"
}

resource "oxide_silo" "silo2" {
	name          = "terraform-acc-test-silo-2"
	description   = "Second test silo for multi-silo link test"
	discoverable  = false
	identity_mode = "local_only"
	quotas = {
		cpus    = 1
		memory  = 1073741824
		storage = 1073741824
	}
	tls_certificates = [
		{
			name        = "self-signed"
			description = "Self-signed certificate"
			cert        = tls_self_signed_cert.self_signed.cert_pem
			key         = tls_private_key.self_signed.private_key_pem
			service     = "external_api"
		},
	]
}

resource "oxide_subnet_pool" "test" {
	name        = "terraform-acc-subnet-pool-multi-silo-test"
	description = "a test subnet pool for multi-silo link tests"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_silo_link" "link1" {
	subnet_pool_id = oxide_subnet_pool.test.id
	silo_id        = data.oxide_silo.silo1.id
	is_default     = false
}

resource "oxide_subnet_pool_silo_link" "link2" {
	subnet_pool_id = oxide_subnet_pool.test.id
	silo_id        = oxide_silo.silo2.id
	is_default     = false
}
`, siloDNSName, siloDNSName)
}

func testAccLinksDestroyed(poolResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[poolResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", poolResourceName)
		}

		client, err := newTestClient()
		if err != nil {
			return err
		}

		links, err := client.SystemSubnetPoolSiloListAllPages(
			context.Background(),
			oxide.SystemSubnetPoolSiloListParams{Pool: oxide.NameOrId(rs.Primary.ID)},
		)
		if err != nil {
			return err
		}

		if len(links) > 0 {
			return fmt.Errorf(
				"expected no silo links for pool %s, got %d",
				rs.Primary.ID,
				len(links),
			)
		}

		return nil
	}
}
