// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsilo_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	fwresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
	"github.com/stretchr/testify/require"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
	systemsilo "github.com/oxidecomputer/terraform-provider-oxide/internal/provider/system_silo"
)

type resourceConfig struct {
	BlockName      string
	ResourceType   string
	SiloName       string
	SiloDNSName    string
	MoveFromLegacy bool
}

var resourceConfigTpl = `
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

resource "{{.ResourceType}}" "{{.BlockName}}" {
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

{{ if .MoveFromLegacy }}
moved {
  from = oxide_silo.{{.BlockName}}
  to   = oxide_system_silo.{{.BlockName}}
}
{{ end }}
`

var resourceUpdateConfigTpl = `
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

resource "{{.ResourceType}}" "{{.BlockName}}" {
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

func TestSiloResourceMetadataAndDeprecation(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		resource           fwresource.Resource
		typeName           string
		deprecationMessage string
		stateMoverCount    int
	}{
		"canonical": {
			resource:        systemsilo.NewResource(),
			typeName:        "oxide_system_silo",
			stateMoverCount: 1,
		},
		"deprecated": {
			resource: systemsilo.NewDeprecatedResource(),
			typeName: "oxide_silo",
			deprecationMessage: "Use oxide_system_silo instead. " +
				"The oxide_silo resource will be removed in a future release.",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			metadataResponse := &fwresource.MetadataResponse{}
			test.resource.Metadata(
				ctx,
				fwresource.MetadataRequest{ProviderTypeName: "oxide"},
				metadataResponse,
			)
			require.Equal(t, test.typeName, metadataResponse.TypeName)

			schemaResponse := &fwresource.SchemaResponse{}
			test.resource.Schema(
				ctx,
				fwresource.SchemaRequest{},
				schemaResponse,
			)
			require.Equal(
				t,
				test.deprecationMessage,
				schemaResponse.Schema.DeprecationMessage,
			)

			resourceWithMoveState, ok := test.resource.(fwresource.ResourceWithMoveState)
			require.True(t, ok)
			require.Len(
				t,
				resourceWithMoveState.MoveState(ctx),
				test.stateMoverCount,
			)
		})
	}
}

func TestSiloResourceMoveStateRejectsUnexpectedSource(t *testing.T) {
	t.Parallel()

	resourceWithMoveState := systemsilo.NewResource().(fwresource.ResourceWithMoveState)
	mover := resourceWithMoveState.MoveState(context.Background())[0]

	tests := map[string]fwresource.MoveStateRequest{
		"resource type": {
			SourceTypeName:        "oxide_other",
			SourceSchemaVersion:   0,
			SourceProviderAddress: "registry.terraform.io/oxidecomputer/oxide",
		},
		"schema version": {
			SourceTypeName:        "oxide_silo",
			SourceSchemaVersion:   1,
			SourceProviderAddress: "registry.terraform.io/oxidecomputer/oxide",
		},
		"provider address": {
			SourceTypeName:        "oxide_silo",
			SourceSchemaVersion:   0,
			SourceProviderAddress: "registry.terraform.io/hashicorp/oxide",
		},
	}

	for name, request := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			response := &fwresource.MoveStateResponse{}
			mover.StateMover(context.Background(), request, response)

			require.Empty(t, response.Diagnostics)
			require.Nil(t, response.TargetState.Schema)
		})
	}
}

func TestSiloResourceMoveStateAcceptsOpenTofuProvider(t *testing.T) {
	t.Parallel()

	resourceWithMoveState := systemsilo.NewResource().(fwresource.ResourceWithMoveState)
	mover := resourceWithMoveState.MoveState(context.Background())[0]
	request := fwresource.MoveStateRequest{
		SourceTypeName:        "oxide_silo",
		SourceSchemaVersion:   0,
		SourceProviderAddress: "registry.opentofu.org/oxidecomputer/oxide",
	}
	response := &fwresource.MoveStateResponse{}

	mover.StateMover(context.Background(), request, response)

	require.Len(t, response.Diagnostics, 1)
	require.Equal(
		t,
		"Unable to Move Silo State",
		response.Diagnostics[0].Summary(),
	)
}

func TestAccSiloResourceSilo_full(t *testing.T) {
	t.Setenv(resource.EnvTfAccProviderNamespace, "oxidecomputer")

	siloName := sharedtest.NewResourceName()
	blockName := sharedtest.NewBlockName("silo")
	resourceName := fmt.Sprintf("oxide_system_silo.%s", blockName)

	dnsName := sharedtest.SiloDNSName()

	config := sharedtest.ParsedAccConfig(t,
		resourceConfig{
			BlockName:    blockName,
			ResourceType: "oxide_system_silo",
			SiloName:     siloName,
			SiloDNSName:  dnsName,
		},
		resourceConfigTpl,
	)

	configUpdate := sharedtest.ParsedAccConfig(t,
		resourceConfig{
			BlockName:    blockName,
			ResourceType: "oxide_system_silo",
			SiloName:     siloName,
			SiloDNSName:  dnsName,
		},
		resourceUpdateConfigTpl,
	)

	// Silo creation and deletion can cause database contention in nexus,
	// so run all related tests in series:
	// https://github.com/oxidecomputer/omicron/issues/9851
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		CheckDestroy: testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResource(resourceName, siloName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceUpdate(resourceName, siloName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSiloResourceSilo_moveFromDeprecated(t *testing.T) {
	t.Setenv(resource.EnvTfAccProviderNamespace, "oxidecomputer")

	siloName := sharedtest.NewResourceName()
	blockName := sharedtest.NewBlockName("silo")
	deprecatedResourceName := fmt.Sprintf("oxide_silo.%s", blockName)
	resourceName := fmt.Sprintf("oxide_system_silo.%s", blockName)
	dnsName := sharedtest.SiloDNSName()

	deprecatedConfig := sharedtest.ParsedAccConfig(t,
		resourceConfig{
			BlockName:    blockName,
			ResourceType: "oxide_silo",
			SiloName:     siloName,
			SiloDNSName:  dnsName,
		},
		resourceConfigTpl,
	)
	config := sharedtest.ParsedAccConfig(t,
		resourceConfig{
			BlockName:      blockName,
			ResourceType:   "oxide_system_silo",
			SiloName:       siloName,
			SiloDNSName:    dnsName,
			MoveFromLegacy: true,
		},
		resourceConfigTpl,
	)

	var siloID string
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		ExternalProviders: map[string]resource.ExternalProvider{
			"tls": {
				Source: "hashicorp/tls",
			},
		},
		CheckDestroy: testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: deprecatedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					checkResource(deprecatedResourceName, siloName),
					sharedtest.CaptureResourceID(
						deprecatedResourceName,
						&siloID,
					),
				),
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkResource(resourceName, siloName),
					resource.TestCheckResourceAttrPtr(
						resourceName,
						"id",
						&siloID,
					),
				),
			},
		},
	})
}

func checkResource(resourceName string, siloName string) resource.TestCheckFunc {
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

func checkResourceUpdate(resourceName string, siloName string) resource.TestCheckFunc {
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

func testAccResourceDestroy(s *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_silo" && rs.Type != "oxide_system_silo" {
			continue
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		params := oxide.SiloViewParams{
			Silo: oxide.NameOrId(rs.Primary.Attributes["id"]),
		}

		res, err := client.SiloView(ctx, params)
		if err != nil && errors.Is(err, oxide.ErrHTTP404) {
			continue
		}

		return fmt.Errorf("silo (%v) still exists", &res.Name)
	}

	return nil
}
