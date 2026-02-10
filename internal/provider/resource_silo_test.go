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

type resourceSiloConfig struct {
	BlockName   string
	SiloName    string
	SiloDNSName string
}

var resourceSiloConfigTpl = `
resource "tls_private_key" "self-signed" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "self-signed" {
  private_key_pem       = tls_private_key.self-signed.private_key_pem
  validity_period_hours = 8760

  subject {
    common_name  = "{{.SiloDNSName}}"
    organization = "Oxide Computer Company"
  }

  dns_names = ["{{.SiloDNSName}}"]

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "oxide_silo" "{{.BlockName}}" {
  name          = "{{.SiloName}}"
  description   = "Managed by Terraform."
  discoverable  = true
  identity_mode = "local_only"

  quotas = {
    cpus    = 2
    memory  = 8589934592
    storage = 8589934592
  }

  mapped_fleet_roles = {
    admin  = ["admin", "collaborator"]
    viewer = ["viewer"]
  }

  tls_certificates = [
    {
      name        = "self-signed-wildcard"
      description = "Self-signed wildcard certificate for *.sys.r3.oxide-preview.com."
      cert        = tls_self_signed_cert.self-signed.cert_pem
      key         = tls_private_key.self-signed.private_key_pem
      service     = "external_api"
    },
  ]

  timeouts = {
    create = "1m"
    read   = "2m"
    update = "3m"
    delete = "4m"
  }
}
`

var resourceSiloUpdateConfigTpl = `
resource "tls_private_key" "self-signed" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "self-signed" {
  private_key_pem       = tls_private_key.self-signed.private_key_pem
  validity_period_hours = 8760

  subject {
    common_name  = "{{.SiloDNSName}}"
    organization = "Oxide Computer Company"
  }

  dns_names = ["{{.SiloDNSName}}"]

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "oxide_silo" "{{.BlockName}}" {
  name          = "{{.SiloName}}"
  description   = "Managed by Terraform."
  discoverable  = true
  identity_mode = "local_only"

  quotas = {
    cpus    = 4           # 2 -> 4
    memory  = 17179869184 # 8 GiB -> 16 GiB
    storage = 17179869184 # 8 GiB -> 16 GiB
  }

  mapped_fleet_roles = {
    admin  = ["admin", "collaborator"]
    viewer = ["viewer"]
  }

  tls_certificates = [
    {
      name        = "self-signed-wildcard"
      description = "Self-signed wildcard certificate for *.sys.r3.oxide-preview.com."
      cert        = tls_self_signed_cert.self-signed.cert_pem
      key         = tls_private_key.self-signed.private_key_pem
      service     = "external_api"
    },
  ]
}
`

func TestAccSiloResourceSilo_full(t *testing.T) {
	siloName := newResourceName()
	blockName := newBlockName("silo")
	resourceName := fmt.Sprintf("oxide_silo.%s", blockName)

	dnsName := testAccSiloDNSName()

	config, err := parsedAccConfig(
		resourceSiloConfig{
			BlockName:   blockName,
			SiloName:    siloName,
			SiloDNSName: dnsName,
		},
		resourceSiloConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configUpdate, err := parsedAccConfig(
		resourceSiloConfig{
			BlockName:   blockName,
			SiloName:    siloName,
			SiloDNSName: dnsName,
		},
		resourceSiloUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		CheckDestroy: testAccSiloDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceSilo(resourceName, siloName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceSiloUpdate(resourceName, siloName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func checkResourceSilo(resourceName string, siloName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", siloName),
		resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform."),
		resource.TestCheckResourceAttr(resourceName, "quotas.cpus", "2"),
		resource.TestCheckResourceAttr(resourceName, "quotas.memory", "8589934592"),
		resource.TestCheckResourceAttr(resourceName, "quotas.storage", "8589934592"),
		resource.TestCheckResourceAttr(resourceName, "discoverable", "true"),
		resource.TestCheckResourceAttr(resourceName, "identity_mode", "local_only"),
		resource.TestCheckResourceAttrSet(resourceName, "mapped_fleet_roles.admin.0"),
		resource.TestCheckResourceAttrSet(resourceName, "mapped_fleet_roles.viewer.0"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "4m"),
	}...)
}

func checkResourceSiloUpdate(resourceName string, siloName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", siloName),
		resource.TestCheckResourceAttr(resourceName, "description", "Managed by Terraform."),
		resource.TestCheckResourceAttr(resourceName, "quotas.cpus", "4"),
		resource.TestCheckResourceAttr(resourceName, "quotas.memory", "17179869184"),
		resource.TestCheckResourceAttr(resourceName, "quotas.storage", "17179869184"),
		resource.TestCheckResourceAttr(resourceName, "discoverable", "true"),
		resource.TestCheckResourceAttr(resourceName, "identity_mode", "local_only"),
		resource.TestCheckResourceAttrSet(resourceName, "mapped_fleet_roles.admin.0"),
		resource.TestCheckResourceAttrSet(resourceName, "mapped_fleet_roles.viewer.0"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
	}...)
}

func testAccSiloDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_silo" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.SiloViewParams{
			Silo: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}

		res, err := client.SiloView(ctx, params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("silo (%v) still exists", &res.Name)
	}

	return nil
}
