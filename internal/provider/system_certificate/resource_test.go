// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemcertificate_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

type resourceConfig struct {
	BlockName   string
	SiloName    string
	SiloDNSName string
	CertKey     string
	ReadTimeout string
	OmitCert    bool
}

var resourceConfigTpl = `
resource "tls_private_key" "initial" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "initial" {
  private_key_pem       = tls_private_key.initial.private_key_pem
  validity_period_hours = 24
  dns_names             = ["{{.SiloDNSName}}"]
  allowed_uses          = ["key_encipherment", "digital_signature", "server_auth"]

  subject {
    common_name = "{{.SiloDNSName}}"
  }
}

resource "tls_private_key" "{{.CertKey}}" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

resource "tls_self_signed_cert" "{{.CertKey}}" {
  private_key_pem       = tls_private_key.{{.CertKey}}.private_key_pem
  validity_period_hours = 48
  dns_names             = ["{{.SiloDNSName}}"]
  allowed_uses          = ["key_encipherment", "digital_signature", "server_auth"]

  subject {
    common_name = "{{.SiloDNSName}}"
  }
}

resource "oxide_system_silo" "{{.BlockName}}" {
  name          = "{{.SiloName}}"
  description   = "Managed by Terraform."
  identity_mode = "local_only"

  quotas = {
    cpus    = 1
    memory  = 1073741824
    storage = 1073741824
  }

  tls_certificates = [
    {
      name        = "initial"
      description = "Initial certificate."
      cert        = tls_self_signed_cert.initial.cert_pem
      key         = tls_private_key.initial.private_key_pem
      service     = "external_api"
    },
  ]
}

{{ if not .OmitCert }}
resource "oxide_system_certificate" "{{.BlockName}}" {
  silo_id     = oxide_system_silo.{{.BlockName}}.id
  name        = "managed"
  description = "Managed by Terraform."
  cert        = tls_self_signed_cert.{{.CertKey}}.cert_pem
  key         = tls_private_key.{{.CertKey}}.private_key_pem
  service     = "external_api"

  timeouts = {
    create = "1m"
    read   = "{{.ReadTimeout}}"
    delete = "3m"
  }
}
{{ end }}
`

func TestAccSystemCertificateResource_full(t *testing.T) {
	t.Setenv(resource.EnvTfAccProviderNamespace, "oxidecomputer")

	siloName := sharedtest.NewResourceName()
	blockName := sharedtest.NewBlockName("silo_certificate")
	resourceName := fmt.Sprintf("oxide_system_certificate.%s", blockName)
	siloResourceName := fmt.Sprintf("oxide_system_silo.%s", blockName)

	newConfig := func(certKey, readTimeout string, omitCert bool) string {
		return sharedtest.ParsedAccConfig(t,
			resourceConfig{
				BlockName:   blockName,
				SiloName:    siloName,
				SiloDNSName: sharedtest.SiloDNSName(),
				CertKey:     certKey,
				ReadTimeout: readTimeout,
				OmitCert:    omitCert,
			},
			resourceConfigTpl,
		)
	}

	var certID string

	// Silo creation and deletion can cause database contention in nexus,
	// so don't run in parallel:
	// https://github.com/oxidecomputer/omicron/issues/9851
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"tls": {Source: "hashicorp/tls"},
		},
		CheckDestroy: testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: newConfig("managed", "2m", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(
						resourceName, "silo_id", siloResourceName, "id",
					),
					resource.TestCheckResourceAttr(resourceName, "name", "managed"),
					resource.TestCheckResourceAttr(
						resourceName, "description", "Managed by Terraform.",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "cert",
						"tls_self_signed_cert.managed", "cert_pem",
					),
					resource.TestCheckNoResourceAttr(resourceName, "key"),
					resource.TestCheckResourceAttr(resourceName, "service", "external_api"),
					resource.TestCheckResourceAttrSet(resourceName, "time_created"),
					resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.create", "1m"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.read", "2m"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "3m"),
					sharedtest.CaptureResourceID(resourceName, &certID),
					checkSiloCertificates(siloName, "initial", "managed"),
				),
			},
			{
				// Import by silo and certificate name.
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           siloName + "/managed",
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
			{
				// Import by silo and certificate ID.
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs := s.RootModule().Resources[resourceName]
					return fmt.Sprintf(
						"%s/%s",
						rs.Primary.Attributes["silo_id"],
						rs.Primary.ID,
					), nil
				},
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
			{
				// A certificate that isn't in the silo is rejected.
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: siloName + "/missing",
				ExpectError:   regexp.MustCompile(`does not belong to silo`),
			},
			{
				// A certificate ID from another silo is rejected.
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					return "test-suite-silo/" +
						s.RootModule().Resources[resourceName].Primary.ID, nil
				},
				ExpectError: regexp.MustCompile(`does not belong to silo`),
			},
			{
				// Changing only timeouts updates the resource in place
				// without replacing the certificate.
				Config: newConfig("managed", "5m", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							resourceName,
							plancheck.ResourceActionUpdate,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "timeouts.read", "5m"),
					resource.TestCheckResourceAttrPtr(resourceName, "id", &certID),
				),
			},
			{
				// A new certificate replaces the existing one without
				// affecting the silo's other certificates.
				Config: newConfig("rotated", "2m", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							resourceName,
							plancheck.ResourceActionReplace,
						),
						plancheck.ExpectResourceAction(
							siloResourceName,
							plancheck.ResourceActionNoop,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName, "cert",
						"tls_self_signed_cert.rotated", "cert_pem",
					),
					func(s *terraform.State) error {
						id := s.RootModule().Resources[resourceName].Primary.ID
						if id == certID {
							return fmt.Errorf("certificate %s was not replaced", id)
						}
						return nil
					},
					checkSiloCertificates(siloName, "initial", "managed"),
				),
			},
			{
				// Removing the resource deletes only its certificate while
				// the silo and its other certificates remain.
				Config: newConfig("rotated", "2m", true),
				Check:  checkSiloCertificates(siloName, "initial"),
			},
		},
	})
}

// checkSiloCertificates verifies the silo has exactly the named certificates.
func checkSiloCertificates(
	siloName string,
	names ...string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		certs, err := client.SiloCertificateListAllPages(
			ctx,
			oxide.SiloCertificateListParams{Silo: oxide.NameOrId(siloName)},
		)
		if err != nil {
			return err
		}

		var got []string
		for _, cert := range certs {
			got = append(got, string(cert.Name))
		}
		if !slices.Equal(got, names) {
			return fmt.Errorf("expected certificates %v, got %v", names, got)
		}

		return nil
	}
}

func testAccResourceDestroy(s *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_system_certificate" {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		_, err := client.SiloCertificateView(
			ctx,
			oxide.SiloCertificateViewParams{
				Certificate: oxide.NameOrId(rs.Primary.ID),
			},
		)
		if err != nil && errors.Is(err, oxide.ErrHTTP404) {
			continue
		}

		return fmt.Errorf("silo certificate (%v) still exists", rs.Primary.ID)
	}

	return nil
}
