// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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

type resourceSiloIdentifyProviderConfig struct {
	SiloBlockName                     string
	SiloDNSName                       string
	SiloName                          string
	SiloSamlIdentityProviderBlockName string
	SiloSamlIdentityProviderName      string
}

var resourceSiloIdentityProviderConfigTpl = `
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

resource "oxide_silo" "{{.SiloBlockName}}" {
  name             = "{{.SiloName}}"
  description      = "Managed by Terraform."
  discoverable     = true
  identity_mode    = "saml_jit"
  admin_group_name = "admin"

  quotas = {
    cpus    = 2
    memory  = 137438953472
    storage = 549755813888
  }

  mapped_fleet_roles = {
    admin  = ["admin", "collaborator"]
    viewer = ["viewer"]
  }

  tls_certificates = [
    {
      name        = "self-signed-wildcard"
      description = "Self-signed wildcard certificate for {{.SiloDNSName}}"
      cert        = tls_self_signed_cert.self-signed.cert_pem
      key         = tls_private_key.self-signed.private_key_pem
      service     = "external_api"
    },
  ]
}

resource "oxide_silo_saml_identity_provider" "{{.SiloSamlIdentityProviderBlockName}}" {
  silo                    = oxide_silo.{{.SiloBlockName}}.id
  name                    = "{{.SiloSamlIdentityProviderName}}"
  description             = "Managed by Terraform."
  group_attribute_name    = "example"
  idp_entity_id           = "example"
  acs_url                 = "https://example.com"
  slo_url                 = "https://example.com"
  sp_client_id            = "example"
  technical_contact_email = "example@example.com"

  idp_metadata_source = {
    type = "base64_encoded_xml"
    data = "PG1kOkVudGl0eURlc2NyaXB0b3IKCXhtbG5zPSJ1cm46b2FzaXM6bmFtZXM6dGM6U0FNTDoyLjA6bWV0YWRhdGEiCgl4bWxuczptZD0idXJuOm9hc2lzOm5hbWVzOnRjOlNBTUw6Mi4wOm1ldGFkYXRhIgoJeG1sbnM6c2FtbD0idXJuOm9hc2lzOm5hbWVzOnRjOlNBTUw6Mi4wOmFzc2VydGlvbiIKCXhtbG5zOmRzPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwLzA5L3htbGRzaWcjIiBlbnRpdHlJRD0iaHR0cHM6Ly9leGFtcGxlLmNvbSI+Cgk8bWQ6SURQU1NPRGVzY3JpcHRvciBXYW50QXV0aG5SZXF1ZXN0c1NpZ25lZD0idHJ1ZSIgcHJvdG9jb2xTdXBwb3J0RW51bWVyYXRpb249InVybjpvYXNpczpuYW1lczp0YzpTQU1MOjIuMDpwcm90b2NvbCI+CgkJPG1kOktleURlc2NyaXB0b3IgdXNlPSJzaWduaW5nIj4KCQkJPGRzOktleUluZm8+CgkJCQk8ZHM6S2V5TmFtZT5xWlc3N3I2Vy1NQVhCQURPWkdfb0lVeGlWSmhzWHJGa2tEUlFlQWREYzhjPC9kczpLZXlOYW1lPgoJCQkJPGRzOlg1MDlEYXRhPgoJCQkJCTxkczpYNTA5Q2VydGlmaWNhdGU+TUlJQ3VUQ0NBYUVDQmdHVUdOYUluekFOQmdrcWhraUc5dzBCQVFzRkFEQWdNUjR3SEFZRFZRUUREQlZrWlcxdkxUQXdaakF6WWpoaFpEUmpaVFExT0RJd0hoY05NalF4TWpNd01UZ3pNREF3V2hjTk16UXhNak13TVRnek1UUXdXakFnTVI0d0hBWURWUVFEREJWa1pXMXZMVEF3WmpBellqaGhaRFJqWlRRMU9ESXdnZ0VpTUEwR0NTcUdTSWIzRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFDc04yR2Y4Z040b0hHSVI3NXZTaDBZakc0ejFyNytLSGx2cG84QnZmRm9hVk5QR255NHNOMVJGdlI5V25pOVMvM0lXRHNjaDV3NTZnMnk3MFNYcmloUWVKZUptUHhucVd1cUFuSDVLeUgxcjFZeVNqK3pHRGJpRHJyM3pBNlYvdFErUHlJZ0R1cUEvaGg1cmxoRndwOEdQRndnZFBCNTEyK2x5MmR5bTkzQ1BrdDdTdk1KQXhlOHFWYWZPTU9nVEIzcUdiT25jSDdYd1BMMnlhTUhKYUlsTFMwSHh5Ti81S1RrUk5aZERwb25JTFYvajlZT2hZSDdJRDl3c3Y2dlR2NnM3Y3BST0dPMmdFOUVPM1pzTTlwUWlxMjN0RGlTUjloY3BvT2piOElyc0VzMXYxZDlkUGRTV2xybSt4L0U1THlZb1VVT1RUdnlpWDdrU0dPVVFMbS9BZ01CQUFFd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFHTGhBSXlITURVSk9QNkttNzRIYjhUSncxdVh4bVJXRjBRcDBza1BtcUxFSEdRYVpic0R4YVBNYzR0ZDI2TUY0R2dMZ2FNZXVnQlkwZk12d0ErMGJDR0EwY2hLVWpJQUwybEg0UGpzVG15cGliVWMySDFlU1BsOTNoYlBzaTFPSm85bTRvblVHcmg3Z3hHWWJ5Tm85R0tBemgvMmZyZVNRbE82K1pmZkdmSlhabEtTT3pnangzcUVYOTc1N2t1Q3VYWVdtQ3hGbmROd0h2ZjdOTVJENUlQOHRYeEN4a09sbUFBZjRKbUlBbnNsNVp1KzR6K0NuZE1vNkYxMjFsT0t4Tkt2Y0ZYaHFQaHJyd1krZlFreEpXdWprVGp2dTRTN1FsM0dOMzVDWURZdDg5a2drL212VmVCaHVRd1pBR3dWRzE3RnlsN3BNRlFyTGtUQ2ErQmJyY1U9PC9kczpYNTA5Q2VydGlmaWNhdGU+CgkJCQk8L2RzOlg1MDlEYXRhPgoJCQk8L2RzOktleUluZm8+CgkJPC9tZDpLZXlEZXNjcmlwdG9yPgoJCTxtZDpBcnRpZmFjdFJlc29sdXRpb25TZXJ2aWNlIEJpbmRpbmc9InVybjpvYXNpczpuYW1lczp0YzpTQU1MOjIuMDpiaW5kaW5nczpTT0FQIiBMb2NhdGlvbj0iaHR0cHM6Ly9leGFtcGxlLmNvbSIgaW5kZXg9IjAiLz4KCQk8bWQ6U2luZ2xlTG9nb3V0U2VydmljZSBCaW5kaW5nPSJ1cm46b2FzaXM6bmFtZXM6dGM6U0FNTDoyLjA6YmluZGluZ3M6SFRUUC1QT1NUIiBMb2NhdGlvbj0iaHR0cHM6Ly9leGFtcGxlLmNvbSIvPgoJCTxtZDpTaW5nbGVMb2dvdXRTZXJ2aWNlIEJpbmRpbmc9InVybjpvYXNpczpuYW1lczp0YzpTQU1MOjIuMDpiaW5kaW5nczpIVFRQLVJlZGlyZWN0IiBMb2NhdGlvbj0iaHR0cHM6Ly9leGFtcGxlLmNvbSIvPgoJCTxtZDpTaW5nbGVMb2dvdXRTZXJ2aWNlIEJpbmRpbmc9InVybjpvYXNpczpuYW1lczp0YzpTQU1MOjIuMDpiaW5kaW5nczpIVFRQLUFydGlmYWN0IiBMb2NhdGlvbj0iaHR0cHM6Ly9leGFtcGxlLmNvbSIvPgoJCTxtZDpTaW5nbGVMb2dvdXRTZXJ2aWNlIEJpbmRpbmc9InVybjpvYXNpczpuYW1lczp0YzpTQU1MOjIuMDpiaW5kaW5nczpTT0FQIiBMb2NhdGlvbj0iaHR0cHM6Ly9leGFtcGxlLmNvbSIvPgoJCTxtZDpOYW1lSURGb3JtYXQ+dXJuOm9hc2lzOm5hbWVzOnRjOlNBTUw6Mi4wOm5hbWVpZC1mb3JtYXQ6cGVyc2lzdGVudDwvbWQ6TmFtZUlERm9ybWF0PgoJCTxtZDpOYW1lSURGb3JtYXQ+dXJuOm9hc2lzOm5hbWVzOnRjOlNBTUw6Mi4wOm5hbWVpZC1mb3JtYXQ6dHJhbnNpZW50PC9tZDpOYW1lSURGb3JtYXQ+CgkJPG1kOk5hbWVJREZvcm1hdD51cm46b2FzaXM6bmFtZXM6dGM6U0FNTDoxLjE6bmFtZWlkLWZvcm1hdDp1bnNwZWNpZmllZDwvbWQ6TmFtZUlERm9ybWF0PgoJCTxtZDpOYW1lSURGb3JtYXQ+dXJuOm9hc2lzOm5hbWVzOnRjOlNBTUw6MS4xOm5hbWVpZC1mb3JtYXQ6ZW1haWxBZGRyZXNzPC9tZDpOYW1lSURGb3JtYXQ+CgkJPG1kOlNpbmdsZVNpZ25PblNlcnZpY2UgQmluZGluZz0idXJuOm9hc2lzOm5hbWVzOnRjOlNBTUw6Mi4wOmJpbmRpbmdzOkhUVFAtUE9TVCIgTG9jYXRpb249Imh0dHBzOi8vZXhhbXBsZS5jb20iLz4KCQk8bWQ6U2luZ2xlU2lnbk9uU2VydmljZSBCaW5kaW5nPSJ1cm46b2FzaXM6bmFtZXM6dGM6U0FNTDoyLjA6YmluZGluZ3M6SFRUUC1SZWRpcmVjdCIgTG9jYXRpb249Imh0dHBzOi8vZXhhbXBsZS5jb20iLz4KCQk8bWQ6U2luZ2xlU2lnbk9uU2VydmljZSBCaW5kaW5nPSJ1cm46b2FzaXM6bmFtZXM6dGM6U0FNTDoyLjA6YmluZGluZ3M6U09BUCIgTG9jYXRpb249Imh0dHBzOi8vZXhhbXBsZS5jb20iLz4KCQk8bWQ6U2luZ2xlU2lnbk9uU2VydmljZSBCaW5kaW5nPSJ1cm46b2FzaXM6bmFtZXM6dGM6U0FNTDoyLjA6YmluZGluZ3M6SFRUUC1BcnRpZmFjdCIgTG9jYXRpb249Imh0dHBzOi8vZXhhbXBsZS5jb20iLz4KCTwvbWQ6SURQU1NPRGVzY3JpcHRvcj4KPC9tZDpFbnRpdHlEZXNjcmlwdG9yPgo="
  }
}
`

func TestAccSiloResourceSiloSamlIdentityProvider_full(t *testing.T) {
	siloBlockName := newBlockName("silo")
	siloName := newResourceName()
	siloSamlIdentityProviderBlockName := newBlockName("silo-idp")
	siloSamlIdentityProviderName := newResourceName()

	siloSamlIdentityProviderResourceID := fmt.Sprintf(
		"oxide_silo_saml_identity_provider.%s",
		siloSamlIdentityProviderBlockName,
	)

	siloDNSName := testAccSiloDNSName()

	config, err := parsedAccConfig(
		resourceSiloIdentifyProviderConfig{
			SiloBlockName:                     siloBlockName,
			SiloDNSName:                       siloDNSName,
			SiloName:                          siloName,
			SiloSamlIdentityProviderBlockName: siloSamlIdentityProviderBlockName,
			SiloSamlIdentityProviderName:      siloSamlIdentityProviderName,
		},
		resourceSiloIdentityProviderConfigTpl,
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
		CheckDestroy: testAccSiloSamlIdentityProviderDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: checkResourceSiloSamlIdentityProvider(
					siloSamlIdentityProviderResourceID,
					siloSamlIdentityProviderName,
				),
			},
		},
	})
}

func checkResourceSiloSamlIdentityProvider(
	resourceID string,
	nameAttr string,
) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceID, "id"),
		resource.TestCheckResourceAttr(resourceID, "name", nameAttr),
		resource.TestCheckResourceAttr(resourceID, "description", "Managed by Terraform."),
		resource.TestCheckResourceAttrSet(resourceID, "silo"),
		resource.TestCheckResourceAttr(resourceID, "group_attribute_name", "example"),
		resource.TestCheckResourceAttr(resourceID, "idp_entity_id", "example"),
		resource.TestCheckResourceAttr(resourceID, "acs_url", "https://example.com"),
		resource.TestCheckResourceAttr(resourceID, "slo_url", "https://example.com"),
		resource.TestCheckResourceAttr(resourceID, "sp_client_id", "example"),
		resource.TestCheckResourceAttr(
			resourceID,
			"technical_contact_email",
			"example@example.com",
		),
		resource.TestCheckResourceAttr(
			resourceID,
			"idp_metadata_source.type",
			"base64_encoded_xml",
		),
		resource.TestCheckResourceAttrSet(resourceID, "idp_metadata_source.data"),
	}...)
}

func testAccSiloSamlIdentityProviderDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_silo_saml_identity_provider" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.SamlIdentityProviderViewParams{
			Provider: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}

		res, err := client.SamlIdentityProviderView(ctx, params)
		if err != nil && is404(err) {
			continue
		}

		return fmt.Errorf("silo saml identity provider (%v) still exists", &res.Name)
	}

	return nil
}
