// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcinternetgatewayipaddressattachment_test

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

type resourceConfig struct {
	Address                string
	AttachmentGatewayBlock string
	AttachmentName         string
	Description            string
	GatewayName            string
	ReplacementGatewayName string
	ReplacementVPCName     string
	VPCName                string
}

const resourceConfigTpl = `
data "oxide_project" "test" {
  name = "tf-acc-test"
}

resource "oxide_vpc" "test" {
  project_id  = data.oxide_project.test.id
  description = "a test VPC for an internet gateway IP address attachment"
  name        = "{{.VPCName}}"
  dns_name    = "attachment-test"
}

resource "oxide_vpc_internet_gateway" "test" {
  description = "a test internet gateway for an IP address attachment"
  name        = "{{.GatewayName}}"
  vpc_id      = oxide_vpc.test.id
}

resource "oxide_vpc" "replacement" {
  project_id  = data.oxide_project.test.id
  description = "a replacement VPC for an internet gateway IP address attachment"
  name        = "{{.ReplacementVPCName}}"
  dns_name    = "attachment-replacement"
}

resource "oxide_vpc_internet_gateway" "replacement" {
  description = "a replacement internet gateway for an IP address attachment"
  name        = "{{.ReplacementGatewayName}}"
  vpc_id      = oxide_vpc.replacement.id
}

resource "oxide_vpc_internet_gateway_ip_address_attachment" "test" {
  gateway_id  = oxide_vpc_internet_gateway.{{.AttachmentGatewayBlock}}.id
  address     = "{{.Address}}"
  name        = "{{.AttachmentName}}"
  description = "{{.Description}}"
  timeouts = {
    create = "1m"
    read   = "1m"
    delete = "1m"
  }
}
`

func TestAccResourceVPCInternetGatewayIPAddressAttachment_full(t *testing.T) {
	const resourceName = "oxide_vpc_internet_gateway_ip_address_attachment.test"
	config := resourceConfig{
		Address:                "198.51.100.47",
		AttachmentGatewayBlock: "test",
		AttachmentName:         sharedtest.NewResourceName(),
		Description:            "a test IP address attachment",
		GatewayName:            sharedtest.NewResourceName(),
		ReplacementGatewayName: sharedtest.NewResourceName(),
		ReplacementVPCName:     sharedtest.NewResourceName(),
		VPCName:                sharedtest.NewResourceName(),
	}
	replacementConfig := config
	replacementConfig.AttachmentGatewayBlock = "replacement"

	var originalID, replacementID string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildResourceConfig(t, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"gateway_id",
						"oxide_vpc_internet_gateway.test",
						"id",
					),
					resource.TestCheckResourceAttr(resourceName, "address", config.Address),
					resource.TestCheckResourceAttr(resourceName, "name", config.AttachmentName),
					resource.TestCheckResourceAttr(resourceName, "description", config.Description),
					sharedtest.CaptureResourceID(resourceName, &originalID),
				),
			},
			// Changing the gateway ID replaces the attachment.
			{
				Config: buildResourceConfig(t, replacementConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						resourceName,
						"gateway_id",
						"oxide_vpc_internet_gateway.replacement",
						"id",
					),
					sharedtest.VerifyResourceIDChanged(resourceName, &originalID),
					sharedtest.CaptureResourceID(resourceName, &replacementID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"timeouts",
				},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					gateway, ok := state.RootModule().Resources["oxide_vpc_internet_gateway.replacement"]
					if !ok {
						return "", fmt.Errorf("replacement internet gateway not found")
					}
					attachment, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return fmt.Sprintf(
						"%s/%s",
						gateway.Primary.ID,
						attachment.Primary.ID,
					), nil
				},
			},
			{
				Config: buildResourceConfig(t, replacementConfig),
				Check: resource.TestCheckResourceAttrPtr(
					resourceName,
					"id",
					&replacementID,
				),
			},
		},
	})
}

func TestAccResourceVPCInternetGatewayIPAddressAttachment_disappears(t *testing.T) {
	const resourceName = "oxide_vpc_internet_gateway_ip_address_attachment.test"
	config := resourceConfig{
		Address:                "198.51.100.49",
		AttachmentGatewayBlock: "test",
		AttachmentName:         sharedtest.NewResourceName(),
		Description:            "a test IP address attachment",
		GatewayName:            sharedtest.NewResourceName(),
		ReplacementGatewayName: sharedtest.NewResourceName(),
		ReplacementVPCName:     sharedtest.NewResourceName(),
		VPCName:                sharedtest.NewResourceName(),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildResourceConfig(t, config),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					deleteAttachment(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func buildResourceConfig(t *testing.T, config resourceConfig) string {
	t.Helper()
	return sharedtest.ParsedAccConfig(t, config, resourceConfigTpl)
}

func deleteAttachment(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}
		return client.InternetGatewayIpAddressDelete(
			context.Background(),
			oxide.InternetGatewayIpAddressDeleteParams{
				Address: oxide.NameOrId(resourceState.Primary.ID),
				Cascade: oxide.NewPointer(false),
			},
		)
	}
}

func testAccResourceDestroy(state *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, resourceState := range state.RootModule().Resources {
		if resourceState.Type != "oxide_vpc_internet_gateway_ip_address_attachment" {
			continue
		}

		attachments, err := client.InternetGatewayIpAddressListAllPages(
			context.Background(),
			oxide.InternetGatewayIpAddressListParams{
				Gateway: oxide.NameOrId(resourceState.Primary.Attributes["gateway_id"]),
			},
		)
		if err != nil {
			if errors.Is(err, oxide.ErrHTTP404) {
				continue
			}
			return err
		}

		if slices.ContainsFunc(attachments, func(attachment oxide.InternetGatewayIpAddress) bool {
			return attachment.Id == resourceState.Primary.ID
		}) {
			return fmt.Errorf(
				"internet gateway IP address attachment %s still exists",
				resourceState.Primary.ID,
			)
		}
	}

	return nil
}
