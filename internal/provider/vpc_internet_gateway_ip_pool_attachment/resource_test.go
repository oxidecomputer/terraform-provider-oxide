// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcinternetgatewayippoolattachment_test

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
	AttachmentDescription  string
	AttachmentGatewayBlock string
	AttachmentName         string
	GatewayOneBlock        string
	GatewayOneName         string
	GatewayTwoBlock        string
	GatewayTwoName         string
	VPCOneBlock            string
	VPCOneName             string
	VPCTwoBlock            string
	VPCTwoName             string
}

const resourceConfigTpl = `
data "oxide_project" "test" {
	name = "tf-acc-test"
}

data "oxide_ip_pool" "test" {
	name = "non-default"
}

resource "oxide_vpc" "{{.VPCOneBlock}}" {
	project_id  = data.oxide_project.test.id
	description = "first VPC for internet gateway IP pool attachment tests"
	name        = "{{.VPCOneName}}"
	dns_name    = "attachment-one"
}

resource "oxide_vpc_internet_gateway" "{{.GatewayOneBlock}}" {
	description    = "first internet gateway for IP pool attachment tests"
	name           = "{{.GatewayOneName}}"
	vpc_id         = oxide_vpc.{{.VPCOneBlock}}.id
	cascade_delete = true
}

resource "oxide_vpc" "{{.VPCTwoBlock}}" {
	project_id  = data.oxide_project.test.id
	description = "second VPC for internet gateway IP pool attachment tests"
	name        = "{{.VPCTwoName}}"
	dns_name    = "attachment-two"
}

resource "oxide_vpc_internet_gateway" "{{.GatewayTwoBlock}}" {
	description    = "second internet gateway for IP pool attachment tests"
	name           = "{{.GatewayTwoName}}"
	vpc_id         = oxide_vpc.{{.VPCTwoBlock}}.id
	cascade_delete = true
}

resource "oxide_vpc_internet_gateway_ip_pool_attachment" "test" {
	name        = "{{.AttachmentName}}"
	description = "{{.AttachmentDescription}}"
	gateway_id  = oxide_vpc_internet_gateway.{{.AttachmentGatewayBlock}}.id
	ip_pool_id  = data.oxide_ip_pool.test.id
	timeouts = {
		create = "1m"
		read   = "1m"
		delete = "1m"
	}
}
`

func TestAccResourceVPCInternetGatewayIPPoolAttachment_full(t *testing.T) {
	resourceName := "oxide_vpc_internet_gateway_ip_pool_attachment.test"
	cfg := resourceConfig{
		AttachmentDescription: "an internet gateway IP pool attachment",
		AttachmentName:        sharedtest.NewResourceName(),
		GatewayOneBlock:       sharedtest.NewBlockName("gateway_one"),
		GatewayOneName:        sharedtest.NewResourceName(),
		GatewayTwoBlock:       sharedtest.NewBlockName("gateway_two"),
		GatewayTwoName:        sharedtest.NewResourceName(),
		VPCOneBlock:           sharedtest.NewBlockName("vpc_one"),
		VPCOneName:            sharedtest.NewResourceName(),
		VPCTwoBlock:           sharedtest.NewBlockName("vpc_two"),
		VPCTwoName:            sharedtest.NewResourceName(),
	}

	firstCfg := cfg
	firstCfg.AttachmentGatewayBlock = cfg.GatewayOneBlock
	secondCfg := cfg
	secondCfg.AttachmentGatewayBlock = cfg.GatewayTwoBlock
	secondCfg.AttachmentDescription = "an updated internet gateway IP pool attachment"
	secondCfg.AttachmentName += "-updated"

	var firstAttachmentID, secondAttachmentID string
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: sharedtest.ParsedAccConfig(t, firstCfg, resourceConfigTpl),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkResource(resourceName, firstCfg),
					sharedtest.CaptureResourceID(resourceName, &firstAttachmentID),
				),
			},
			// Changing immutable attachment fields replaces the attachment.
			{
				Config: sharedtest.ParsedAccConfig(t, secondCfg, resourceConfigTpl),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkResource(resourceName, secondCfg),
					sharedtest.VerifyResourceIDChanged(resourceName, &firstAttachmentID),
					sharedtest.CaptureResourceID(resourceName, &secondAttachmentID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"timeouts",
				},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					gatewayResource := fmt.Sprintf(
						"oxide_vpc_internet_gateway.%s",
						cfg.GatewayTwoBlock,
					)
					gatewayState, ok := s.RootModule().Resources[gatewayResource]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", gatewayResource)
					}
					attachmentState, ok := s.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return fmt.Sprintf(
						"%s/%s",
						gatewayState.Primary.ID,
						attachmentState.Primary.ID,
					), nil
				},
			},
			// Verify the imported attachment remains unchanged by the configuration.
			{
				Config: sharedtest.ParsedAccConfig(t, secondCfg, resourceConfigTpl),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkResource(resourceName, secondCfg),
					resource.TestCheckResourceAttrPtr(
						resourceName,
						"id",
						&secondAttachmentID,
					),
				),
			},
			// Delete the attachment out of band and verify Read plans recreation.
			{
				Config: sharedtest.ParsedAccConfig(t, secondCfg, resourceConfigTpl),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkResource(resourceName, secondCfg),
					testAccResourceDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func checkResource(resourceName string, config resourceConfig) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttr(resourceName, "name", config.AttachmentName),
		resource.TestCheckResourceAttr(
			resourceName,
			"description",
			config.AttachmentDescription,
		),
		resource.TestCheckResourceAttrSet(resourceName, "ip_pool_id"),
		resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "1m"),
	)
}

func testAccResourceDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resourceState, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}
		return client.InternetGatewayIpPoolDelete(
			context.Background(),
			oxide.InternetGatewayIpPoolDeleteParams{
				Pool:    oxide.NameOrId(resourceState.Primary.ID),
				Cascade: oxide.NewPointer(false),
			},
		)
	}
}

func testAccResourceDestroy(s *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, resourceState := range s.RootModule().Resources {
		if resourceState.Type != "oxide_vpc_internet_gateway_ip_pool_attachment" {
			continue
		}

		attachments, err := client.InternetGatewayIpPoolListAllPages(
			context.Background(),
			oxide.InternetGatewayIpPoolListParams{
				Gateway: oxide.NameOrId(resourceState.Primary.Attributes["gateway_id"]),
			},
		)
		if err != nil {
			if errors.Is(err, oxide.ErrHTTP404) {
				continue
			}
			return err
		}

		if slices.ContainsFunc(attachments, func(element oxide.InternetGatewayIpPool) bool {
			return element.Id == resourceState.Primary.ID
		}) {
			return fmt.Errorf(
				"internet gateway IP pool attachment %s still exists",
				resourceState.Primary.ID,
			)
		}
	}

	return nil
}
