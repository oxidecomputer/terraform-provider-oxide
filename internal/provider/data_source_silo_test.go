// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceSiloConfig struct {
	BlockName   string
	SiloDNSName string
}

var dataSourceSiloConfigTpl = `
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
  name          = "tf-acc-test"
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

data "oxide_silo" "{{.BlockName}}" {
  name = oxide_silo.{{.BlockName}}.name
  timeouts = {
    read = "1m"
  }
}
`

func TestAccSiloDataSourceSilo_full(t *testing.T) {
	blockName := newBlockName("datasource-silo")

	dnsName := testAccSiloDNSName()

	config, err := parsedAccConfig(
		dataSourceSiloConfig{
			BlockName:   blockName,
			SiloDNSName: dnsName,
		},
		dataSourceSiloConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	// Silo creation and deletion can cause database contention in nexus,
	// so run all related tests in series:
	// https://github.com/oxidecomputer/omicron/issues/9851
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: checkDataSourceSilo(
					fmt.Sprintf("data.oxide_silo.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceSilo(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "name"),
		resource.TestCheckResourceAttr(dataName, "name", "tf-acc-test"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "discoverable"),
		resource.TestCheckResourceAttrSet(dataName, "identity_mode"),
		resource.TestCheckResourceAttrSet(dataName, "description"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
	}...)
}
