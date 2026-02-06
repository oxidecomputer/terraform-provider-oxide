// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// externalSubnetTestConfig holds parameters for generating test configurations.
// Note that subnet pools cannot overlap, and tests run in parallel, so we need
// to use non-overlapping pools across tests.
type externalSubnetTestConfig struct {
	PoolName         string
	PoolDescription  string
	PoolIPVersion    string
	PoolMemberSubnet string
	MaxPrefixLength  int
	IsDefault        bool

	SubnetName        string
	SubnetDescription string
	SubnetPrefixLen   int
	SubnetCIDR        string
	SubnetPool        string
	SubnetIPVersion   string
}

var externalSubnetConfigTpl = `
data "oxide_project" "test" {
	name = "tf-acc-test"
}

data "oxide_silo" "test" {
	name = "test-suite-silo"
}

resource "oxide_subnet_pool" "test" {
	name        = "{{.PoolName}}"
	description = "{{.PoolDescription}}"
	ip_version  = "{{.PoolIPVersion}}"
}

resource "oxide_subnet_pool_member" "test" {
	subnet_pool_id    = oxide_subnet_pool.test.id
	subnet            = "{{.PoolMemberSubnet}}"
	max_prefix_length = {{.MaxPrefixLength}}
}

resource "oxide_subnet_pool_silo_link" "test" {
	subnet_pool_id = oxide_subnet_pool.test.id
	silo_id        = data.oxide_silo.test.id
	is_default     = {{.IsDefault}}
}

resource "oxide_external_subnet" "test" {
	project_id  = data.oxide_project.test.id
	name        = "{{.SubnetName}}"
	description = "{{.SubnetDescription}}"
{{- if .SubnetCIDR}}
	subnet      = "{{.SubnetCIDR}}"
{{- else}}
	prefix_len  = {{.SubnetPrefixLen}}
{{- end}}
{{- if .SubnetPool}}
	subnet_pool_id = {{.SubnetPool}}
{{- end}}
{{- if .SubnetIPVersion}}
	ip_version  = "{{.SubnetIPVersion}}"
{{- end}}
	depends_on  = [oxide_subnet_pool_silo_link.test, oxide_subnet_pool_member.test]
	timeouts = {
		create = "1m"
		read   = "1m"
		update = "1m"
		delete = "1m"
	}
}
`

func buildExternalSubnetConfig(t *testing.T, cfg externalSubnetTestConfig) string {
	t.Helper()
	config, err := parsedAccConfig(cfg, externalSubnetConfigTpl)
	if err != nil {
		t.Fatalf("error parsing config template: %v", err)
	}
	return config
}

func TestAccResourceExternalSubnet_full(t *testing.T) {
	resourceName := "oxide_external_subnet.test"
	var originalID string

	baseConfig := externalSubnetTestConfig{
		PoolName:         "terraform-acc-ext-subnet-pool",
		PoolDescription:  "a subnet pool for external subnet tests",
		PoolIPVersion:    "v4",
		PoolMemberSubnet: fmt.Sprintf("192.%d.%d.0/24", rand.IntN(255), rand.IntN(255)),
		MaxPrefixLength:  30,
		IsDefault:        true,
		SubnetIPVersion:  "v4",
	}

	// Create config.
	createConfig := baseConfig
	createConfig.SubnetName = "terraform-acc-external-subnet"
	createConfig.SubnetDescription = "a test external subnet"
	createConfig.SubnetPrefixLen = 28

	// In-place update config.
	updateConfig := baseConfig
	updateConfig.SubnetName = "terraform-acc-external-subnet-updated"
	updateConfig.SubnetDescription = "an updated external subnet"
	updateConfig.SubnetPrefixLen = 28

	// Replace config.
	replaceConfig := baseConfig
	replaceConfig.SubnetName = "terraform-acc-external-subnet-updated"
	replaceConfig.SubnetDescription = "an updated external subnet"
	replaceConfig.SubnetPrefixLen = 29

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccExternalSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: buildExternalSubnetConfig(t, createConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						"terraform-acc-external-subnet",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"description",
						"a test external subnet",
					),
					resource.TestCheckResourceAttrSet(resourceName, "project_id"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet"),
					resource.TestCheckResourceAttr(resourceName, "prefix_len", "28"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_pool_member_id"),
					resource.TestCheckResourceAttrSet(resourceName, "time_created"),
					resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
					testAccCaptureResourceID(resourceName, &originalID),
				),
			},
			// Update in place.
			{
				Config: buildExternalSubnetConfig(t, updateConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPtr(resourceName, "id", &originalID),
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						"terraform-acc-external-subnet-updated",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"description",
						"an updated external subnet",
					),
				),
			},
			// Replace.
			{
				Config: buildExternalSubnetConfig(t, replaceConfig),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVerifyResourceIDChanged(resourceName, &originalID),
					resource.TestCheckResourceAttr(resourceName, "prefix_len", "29"),
				),
			},
			// Import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"timeouts",
					"prefix_len",
					"ip_version",
				},
			},
		},
	})
}

func TestAccResourceExternalSubnet_explicit(t *testing.T) {
	resourceName := "oxide_external_subnet.test"

	config := buildExternalSubnetConfig(t, externalSubnetTestConfig{
		PoolName:          "terraform-acc-ext-subnet-pool-explicit",
		PoolDescription:   "a subnet pool for explicit external subnet tests",
		PoolIPVersion:     "v4",
		PoolMemberSubnet:  "198.51.100.0/24",
		MaxPrefixLength:   30,
		IsDefault:         false,
		SubnetName:        "terraform-acc-external-subnet-explicit",
		SubnetDescription: "an explicit external subnet",
		SubnetCIDR:        "198.51.100.0/28",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccExternalSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						"terraform-acc-external-subnet-explicit",
					),
					resource.TestCheckResourceAttr(resourceName, "subnet", "198.51.100.0/28"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_pool_member_id"),
				),
			},
		},
	})
}

func TestAccResourceExternalSubnet_withPool(t *testing.T) {
	resourceName := "oxide_external_subnet.test"

	config := buildExternalSubnetConfig(t, externalSubnetTestConfig{
		PoolName:          "terraform-acc-ext-subnet-pool-with-pool",
		PoolDescription:   "a subnet pool for pool selection tests",
		PoolIPVersion:     "v4",
		PoolMemberSubnet:  "203.0.113.0/24",
		MaxPrefixLength:   30,
		IsDefault:         false,
		SubnetName:        "terraform-acc-external-subnet-with-pool",
		SubnetDescription: "an external subnet with explicit pool",
		SubnetPrefixLen:   28,
		SubnetPool:        "oxide_subnet_pool.test.id",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccExternalSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						"terraform-acc-external-subnet-with-pool",
					),
					resource.TestCheckResourceAttrSet(resourceName, "subnet"),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"subnet_pool_id",
						"oxide_subnet_pool.test",
						"id",
					),
				),
			},
		},
	})
}

func TestAccResourceExternalSubnet_disappears(t *testing.T) {
	resourceName := "oxide_external_subnet.test"

	config := buildExternalSubnetConfig(t, externalSubnetTestConfig{
		PoolName:          "terraform-acc-ext-subnet-pool-disappears",
		PoolDescription:   "a subnet pool for disappears test",
		PoolIPVersion:     "v4",
		PoolMemberSubnet:  "100.64.0.0/24",
		MaxPrefixLength:   30,
		IsDefault:         false,
		SubnetName:        "terraform-acc-external-subnet-disappears",
		SubnetDescription: "external subnet for disappears test",
		SubnetPrefixLen:   28,
		SubnetPool:        "oxide_subnet_pool.test.id",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccExternalSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					testAccExternalSubnetDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccExternalSubnetDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := newTestClient()
		if err != nil {
			return err
		}

		return client.ExternalSubnetDelete(
			context.Background(),
			oxide.ExternalSubnetDeleteParams{ExternalSubnet: oxide.NameOrId(rs.Primary.ID)},
		)
	}
}

func TestAccResourceExternalSubnet_v6(t *testing.T) {
	resourceName := "oxide_external_subnet.test"

	config := buildExternalSubnetConfig(t, externalSubnetTestConfig{
		PoolName:          "terraform-acc-ext-subnet-pool-v6",
		PoolDescription:   "a subnet pool for IPv6 external subnet tests",
		PoolIPVersion:     "v6",
		PoolMemberSubnet:  "2001:db8::/32",
		MaxPrefixLength:   96,
		IsDefault:         true,
		SubnetName:        "terraform-acc-external-subnet-v6",
		SubnetDescription: "an IPv6 external subnet",
		SubnetPrefixLen:   64,
		SubnetIPVersion:   "v6",
	})

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccExternalSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						"terraform-acc-external-subnet-v6",
					),
					resource.TestCheckResourceAttrSet(resourceName, "subnet"),
					resource.TestCheckResourceAttr(resourceName, "prefix_len", "64"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_pool_id"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_pool_member_id"),
				),
			},
		},
	})
}

func testAccExternalSubnetDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_external_subnet" {
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
		if err == nil {
			return fmt.Errorf("external_subnet (%v) still exists", res.Name)
		}
		return err
	}

	return nil
}
