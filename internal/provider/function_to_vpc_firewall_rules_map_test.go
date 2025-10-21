// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"
)

func TestAccFunctionToVPCFirewallRulesMap(t *testing.T) {
	blockName := newBlockName("firewall_rules")
	supportBlockName := newBlockName("support")
	supportBlockName2 := newBlockName("support")
	vpcName := newResourceName()
	resourceName := fmt.Sprintf("oxide_vpc_firewall_rules.%s", blockName)

	tplData := resourceFirewallRulesConfig{
		BlockName:         blockName,
		SupportBlockName:  supportBlockName,
		SupportBlockName2: supportBlockName2,
		VPCName:           vpcName,
	}

	// Generate initial configuration with old schema.
	configWithSet, err := parsedAccConfig(
		tplData,
		testAccFunctionToVPCFirewallRulesConfigTpl,
	)
	require.NoError(t, err)

	// Generate updated configuration using the function to convert the rules
	// set to a map.
	configWithFunction, err := parsedAccConfig(
		tplData,
		functionToVPCFirewallRulesConfigUpdatedTpl,
	)
	require.NoError(t, err)

	// Generate final configuration using the new map schema for rules.
	configWithMap, err := parsedAccConfig(
		tplData,
		testAccFunctionToVPCFirewallRulesConfigUpdatedMapTpl,
	)
	require.NoError(t, err)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccFirewallRulesDestroy,
		Steps: []resource.TestStep{
			{
				// Start with the old format with rules as a set.
				ExternalProviders: map[string]resource.ExternalProvider{
					"oxide": {
						Source:            "oxidecomputer/oxide",
						VersionConstraint: "0.14.1",
					},
				},
				Config: configWithSet,
				Check:  testFunctionToVPCFirewallRulesMap_checkSet(resourceName),
			},
			{
				// Apply provider function to migrate to map.
				ExternalProviders:        map[string]resource.ExternalProvider{},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
				Config:                   configWithFunction,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Verify the function doesn't cause any changes.
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: testFunctionToVPCFirewallRulesMap_checkMap(resourceName),
			},
			{
				// Update configuration to the new schema with rules as a map.
				ExternalProviders:        map[string]resource.ExternalProvider{},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
				Config:                   configWithMap,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						// Verify migrating to map after using the function
						// doesn't cause any changes.
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: testFunctionToVPCFirewallRulesMap_checkMap(resourceName),
			},
		},
	})
}

var (
	testAccFunctionToVPCFirewallRulesConfigTpl = fmt.Sprintf(`
data "oxide_project" "{{.SupportBlockName}}" {
  name = "tf-acc-test"
}

resource "oxide_vpc" "{{.SupportBlockName2}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test vpc"
  name        = "{{.VPCName}}"
  dns_name    = "my-vpc-dns"
}

resource "oxide_vpc_firewall_rules" "{{.BlockName}}" {
  vpc_id = oxide_vpc.{{.SupportBlockName2}}.id

  rules = [
%s
  ]

  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "4m"
  }
}
`, testAccFunctionToVPCFirewallRulesSetTpl)

	functionToVPCFirewallRulesConfigUpdatedTpl = fmt.Sprintf(`
data "oxide_project" "{{.SupportBlockName}}" {
  name = "tf-acc-test"
}

resource "oxide_vpc" "{{.SupportBlockName2}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test vpc"
  name        = "{{.VPCName}}"
  dns_name    = "my-vpc-dns"
}

resource "oxide_vpc_firewall_rules" "{{.BlockName}}" {
  vpc_id = oxide_vpc.{{.SupportBlockName2}}.id

  rules = provider::oxide::to_vpc_firewall_rules_map(jsonencode([
%s
  ]))

  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "4m"
  }
}
`, testAccFunctionToVPCFirewallRulesSetTpl)

	testAccFunctionToVPCFirewallRulesConfigUpdatedMapTpl = fmt.Sprintf(`
data "oxide_project" "{{.SupportBlockName}}" {
  name = "tf-acc-test"
}

resource "oxide_vpc" "{{.SupportBlockName2}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test vpc"
  name        = "{{.VPCName}}"
  dns_name    = "my-vpc-dns"
}

resource "oxide_vpc_firewall_rules" "{{.BlockName}}" {
  vpc_id = oxide_vpc.{{.SupportBlockName2}}.id

  rules = {
%s
  }

  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "4m"
  }
}
`, testAccFunctionToVPCFirewallRulesMapTpl)

	testAccFunctionToVPCFirewallRulesSetTpl = `
{
  action      = "allow"
  description = "Allow HTTP and HTTPS."
  name        = "allow-http-https"
  direction   = "inbound"
  priority    = 50
  status      = "enabled"
  filters = {
    hosts = [
      {
        type  = "vpc"
        value = oxide_vpc.{{.SupportBlockName2}}.name
      }
    ]
    ports = ["80", 443]
    protocols = [
      { type = "tcp" },
      { type = "udp" },
    ]
  },
  targets = [
    {
      type  = "subnet"
      value = "default"
    }
  ]
},
{
  action      = "allow"
  description = "ICMP rules."
  name        = "icmp"
  direction   = "inbound"
  priority    = 50
  status      = "enabled"
  filters = {
    protocols = [
      {
        type = "icmp",
      },
      {
        type      = "icmp",
        icmp_type = 0
      },
      {
        type      = "icmp",
        icmp_type = 0
        icmp_code = "1-3"
      },
    ]
  },
  targets = [
    {
      type  = "vpc"
      value = oxide_vpc.{{.SupportBlockName2}}.name
    }
  ]
},
`

	testAccFunctionToVPCFirewallRulesMapTpl = `
allow-http-https = {
  action      = "allow"
  description = "Allow HTTP and HTTPS."
  direction   = "inbound"
  priority    = 50
  status      = "enabled"
  filters = {
    hosts = [
      {
        type  = "vpc"
        value = oxide_vpc.{{.SupportBlockName2}}.name
      }
    ]
    ports = ["80", 443]
    protocols = [
      { type = "tcp" },
      { type = "udp" },
    ]
  },
  targets = [
    {
      type  = "subnet"
      value = "default"
    }
  ]
},
icmp = {
  action      = "allow"
  description = "ICMP rules."
  direction   = "inbound"
  priority    = 50
  status      = "enabled"
  filters = {
    protocols = [
      {
        type = "icmp",
      },
      {
        type      = "icmp",
        icmp_type = 0
      },
      {
        type      = "icmp",
        icmp_type = 0
        icmp_code = "1-3"
      },
    ]
  },
  targets = [
    {
      type  = "vpc"
      value = oxide_vpc.{{.SupportBlockName2}}.name
    }
  ]
}
`
)

func testFunctionToVPCFirewallRulesMap_checkSet(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		// Rule 1.
		resource.TestCheckResourceAttr(resourceName, "rules.0.action", "allow"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.name", "allow-http-https"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.description", "Allow HTTP and HTTPS."),
		resource.TestCheckResourceAttr(resourceName, "rules.0.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "50"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.status", "enabled"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.protocols.0.type", "tcp"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.protocols.1.type", "udp"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.targets.0.type", "subnet"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.targets.0.value"),
		// Rule 2.
		resource.TestCheckResourceAttr(resourceName, "rules.1.action", "allow"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.name", "icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.description", "ICMP rules."),
		resource.TestCheckResourceAttr(resourceName, "rules.1.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.priority", "50"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.status", "enabled"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.filters.protocols.0.type", "icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.filters.protocols.0.icmp_type", "0"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.filters.protocols.0.icmp_code", "1-3"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.filters.protocols.1.type", "icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.filters.protocols.1.icmp_type", "0"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.filters.protocols.2.type", "icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.targets.0.type", "vpc"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.1.targets.0.value"),
	}...)
}

func testFunctionToVPCFirewallRulesMap_checkMap(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		// Rule 1.
		resource.TestCheckResourceAttr(resourceName, "rules.allow-http-https.action", "allow"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-http-https.name", "allow-http-https"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-http-https.description", "Allow HTTP and HTTPS."),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-http-https.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-http-https.priority", "50"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-http-https.status", "enabled"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-http-https.filters.protocols.0.type", "tcp"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-http-https.filters.protocols.1.type", "udp"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-http-https.targets.0.type", "subnet"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-http-https.targets.0.value"),
		// Rule 2.
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.action", "allow"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.name", "icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.description", "ICMP rules."),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.priority", "50"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.status", "enabled"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.filters.protocols.0.type", "icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.filters.protocols.0.icmp_type", "0"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.filters.protocols.0.icmp_code", "1-3"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.filters.protocols.1.type", "icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.filters.protocols.1.icmp_type", "0"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.filters.protocols.2.type", "icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.icmp.targets.0.type", "vpc"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.icmp.targets.0.value"),
	}...)
}
