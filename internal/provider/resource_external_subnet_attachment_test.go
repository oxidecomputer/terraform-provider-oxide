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

type externalSubnetAttachmentTestConfig struct {
	PoolName         string
	PoolDescription  string
	PoolMemberSubnet string
	MaxPrefixLength  int

	SubnetName        string
	SubnetDescription string
	SubnetCIDR        string

	InstanceName string

	IncludeAttachment bool
}

var externalSubnetAttachmentConfigTpl = `
data "oxide_project" "test" {
	name = "tf-acc-test"
}

data "oxide_silo" "test" {
	name = "test-suite-silo"
}

data "oxide_vpc_subnet" "default" {
	project_name = "tf-acc-test"
	vpc_name     = "default"
	name         = "default"
}

resource "oxide_subnet_pool" "test" {
	name        = "{{.PoolName}}"
	description = "{{.PoolDescription}}"
	ip_version  = "v4"
}

resource "oxide_subnet_pool_member" "test" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = "{{.PoolMemberSubnet}}"
	max_prefix_length = {{.MaxPrefixLength}}
}

resource "oxide_subnet_pool_silo_link" "test" {
	subnet_pool_id = oxide_subnet_pool.test.id
	silo_id        = data.oxide_silo.test.id
	is_default     = false
}

resource "oxide_external_subnet" "test" {
	project_id  = data.oxide_project.test.id
	name        = "{{.SubnetName}}"
	description = "{{.SubnetDescription}}"
	subnet      = "{{.SubnetCIDR}}"
	depends_on  = [oxide_subnet_pool_silo_link.test, oxide_subnet_pool_member.test]
}

resource "oxide_instance" "test" {
	project_id      = data.oxide_project.test.id
	name            = "{{.InstanceName}}"
	description     = "a test instance for external subnet attachment"
	hostname        = "terraform-acc-myhost"
	memory          = 1073741824
	ncpus           = 1
	start_on_create = false
	network_interfaces = [
		{
			subnet_id   = data.oxide_vpc_subnet.default.id
			vpc_id      = data.oxide_vpc_subnet.default.vpc_id
			description = "nic"
			name        = "nic0"

			ip_config = {
				v4 = {
					ip = "auto"
				}
			}
		}
	]
}

{{- if .IncludeAttachment}}

resource "oxide_external_subnet_attachment" "test" {
	external_subnet_id = oxide_external_subnet.test.id
	instance_id        = oxide_instance.test.id
	timeouts = {
		create = "1m"
		read   = "1m"
		delete = "1m"
	}
}
{{- end}}
`

func buildExternalSubnetAttachmentConfig(
	t *testing.T,
	cfg externalSubnetAttachmentTestConfig,
) string {
	t.Helper()
	config, err := parsedAccConfig(cfg, externalSubnetAttachmentConfigTpl)
	if err != nil {
		t.Fatalf("error parsing config template: %v", err)
	}
	return config
}

func TestAccResourceExternalSubnetAttachment_full(t *testing.T) {
	resourceName := "oxide_external_subnet_attachment.test"

	subnet := nextSubnetCIDR(t)

	baseConfig := externalSubnetAttachmentTestConfig{
		PoolName:          "terraform-acc-ext-subnet-attach",
		PoolDescription:   "a subnet pool for external subnet attachment tests",
		PoolMemberSubnet:  subnet,
		MaxPrefixLength:   30,
		SubnetName:        "terraform-acc-ext-subnet-attach",
		SubnetDescription: "an external subnet for attachment test",
		SubnetCIDR:        subnet,
		InstanceName:      "terraform-acc-ext-subnet-attach",
	}

	// Config with attachment.
	attachedConfig := baseConfig
	attachedConfig.IncludeAttachment = true

	// Config without attachment.
	detachedConfig := baseConfig
	detachedConfig.IncludeAttachment = false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		// Note that this check isn't particularly useful, since destroying the subnet and instance
		// will automatically destroy the attachment. We include it for completeness, but the real
		// test is in `testAccExternalSubnetVerifyDetached` below.
		CheckDestroy: testAccExternalSubnetAttachmentDestroy,
		Steps: []resource.TestStep{
			// Create attachment.
			{
				Config: buildExternalSubnetAttachmentConfig(t, attachedConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(
						resourceName, "external_subnet_id",
						"oxide_external_subnet.test", "id",
					),
					resource.TestCheckResourceAttrPair(
						resourceName, "instance_id",
						"oxide_instance.test", "id",
					),
				),
			},
			// Import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"timeouts",
				},
			},
			// Remove attachment.
			{
				Config: buildExternalSubnetAttachmentConfig(t, detachedConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("oxide_external_subnet.test", "id"),
					resource.TestCheckResourceAttrSet("oxide_instance.test", "id"),
					testAccExternalSubnetVerifyDetached("oxide_external_subnet.test"),
				),
			},
		},
	})
}

func TestAccResourceExternalSubnetAttachment_disappears(t *testing.T) {
	resourceName := "oxide_external_subnet_attachment.test"

	subnet := nextSubnetCIDR(t)

	config := buildExternalSubnetAttachmentConfig(t, externalSubnetAttachmentTestConfig{
		PoolName:          "terraform-acc-ext-subnet-attach-pool-dis",
		PoolDescription:   "a subnet pool for disappears test",
		PoolMemberSubnet:  subnet,
		MaxPrefixLength:   30,
		SubnetName:        "terraform-acc-ext-subnet-attach-dis",
		SubnetDescription: "an external subnet for disappears test",
		SubnetCIDR:        subnet,
		InstanceName:      "terraform-acc-ext-subnet-attach-dis",
		IncludeAttachment: true,
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccExternalSubnetAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					testAccExternalSubnetAttachmentDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccExternalSubnetAttachmentDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := newTestClient()
		if err != nil {
			return err
		}

		_, err = client.ExternalSubnetDetach(
			context.Background(),
			oxide.ExternalSubnetDetachParams{ExternalSubnet: oxide.NameOrId(rs.Primary.ID)},
		)
		return err
	}
}

func testAccExternalSubnetVerifyDetached(
	subnetResourceName string,
) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[subnetResourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", subnetResourceName)
		}

		client, err := newTestClient()
		if err != nil {
			return err
		}

		res, err := client.ExternalSubnetView(
			context.Background(),
			oxide.ExternalSubnetViewParams{
				ExternalSubnet: oxide.NameOrId(rs.Primary.ID),
			},
		)
		if err != nil {
			return fmt.Errorf(
				"error viewing external subnet: %v", err,
			)
		}

		if res.InstanceId != "" {
			return fmt.Errorf(
				"external subnet %s still attached to instance %s",
				res.Id,
				res.InstanceId,
			)
		}

		return nil
	}
}

func testAccExternalSubnetAttachmentDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_external_subnet_attachment" {
			continue
		}

		ctx := context.Background()

		res, err := client.ExternalSubnetView(
			ctx,
			oxide.ExternalSubnetViewParams{ExternalSubnet: oxide.NameOrId(rs.Primary.ID)},
		)
		if err != nil && is404(err) {
			continue
		}
		if err == nil && res.InstanceId != "" {
			return fmt.Errorf(
				"external_subnet_attachment (%v) still exists (attached to instance %v)",
				res.Id,
				res.InstanceId,
			)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
