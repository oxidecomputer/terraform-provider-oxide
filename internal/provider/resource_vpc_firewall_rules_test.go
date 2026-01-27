// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
	"github.com/stretchr/testify/require"
)

type resourceFirewallRulesConfig struct {
	BlockName         string
	VPCName           string
	SupportBlockName  string
	SupportBlockName2 string
}

var resourceFirewallRulesConfigTpl = `
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
    custom-deny-http = {
      action      = "deny"
      description = "custom deny"
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
        ports     = ["8123"]
        protocols = [
          {
            type = "tcp"
          },
        ]
      },
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    },
    allow-internal-inbound = {
      action      = "allow"
      description = "custom allow"
      direction   = "inbound"
      priority    = 65534
      status      = "enabled"
      filters = {
        hosts = [
          {
            type  = "vpc"
            value = oxide_vpc.{{.SupportBlockName2}}.name
          }
        ]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    }
  }
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "4m"
  }
}
`

var resourceFirewallRulesUpdateConfigTpl = `
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
    allow-https = {
      action      = "allow"
      description = "Allow HTTPS."
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
        ports     = ["443"]
        protocols = [
          {
            type = "tcp"
          }
        ]
      },
      targets = [
        {
          type  = "subnet"
          value = oxide_vpc.{{.SupportBlockName2}}.name
        }
      ]
    },
    allow-ssh = {
      action      = "allow"
      description = "Allow SSH."
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
        ports     = ["22"]
        protocols = [
          {
            type = "tcp"
          }
        ]
      },
      targets = [
        {
          type  = "subnet"
          value = oxide_vpc.{{.SupportBlockName2}}.name
        }
      ]
    }
  }
}
`

var resourceFirewallRulesUpdateConfigTpl2 = `
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
    allow-https = {
      action      = "allow"
      description = "Allow HTTPS."
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
        ports     = ["443"]
        protocols = [
          {
            type = "tcp"
          }
        ]
      },
      targets = [
        {
          type  = "subnet"
          value = oxide_vpc.{{.SupportBlockName2}}.name
        }
      ]
    },
  }
}
`

var resourceFirewallRulesUpdateConfigTpl3 = `
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
  rules = {}
}
`

var resourceFirewallRulesUpdateConfigTpl4 = `
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
    allow-icmp = {
      action      = "allow"
      description = "ICMP rule."
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        protocols = [
          {
            type = "icmp"
            icmp_type = 3
          },
        ]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    }
  }
}
`

var resourceFirewallRulesUpdateConfigTpl5 = `
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
    allow-icmp = {
      action      = "allow"
      description = "ICMP rule."
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        protocols = [
          {
            type = "icmp"
            icmp_type = 0
            icmp_code = "1-3"
          },
        ]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    }
  }
}
`

var resourceFirewallRulesUpdateConfigTplV1 = `
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
    {
      action      = "allow"
      name        = "allow-icmp"
      description = "ICMP rule."
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        protocols = [
          {
            type = "icmp"
            icmp_type = 0
            icmp_code = "1-3"
          },
        ]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    }
  ]
}
`

func TestAccCloudResourceFirewallRules_full(t *testing.T) {
	blockName := newBlockName("firewall_rules")
	supportBlockName := newBlockName("support")
	supportBlockName2 := newBlockName("support")
	vpcName := newResourceName()
	resourceName := fmt.Sprintf("oxide_vpc_firewall_rules.%s", blockName)
	config, err := parsedAccConfig(
		resourceFirewallRulesConfig{
			BlockName:         blockName,
			SupportBlockName:  supportBlockName,
			SupportBlockName2: supportBlockName2,
			VPCName:           vpcName,
		},
		resourceFirewallRulesConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configUpdate, err := parsedAccConfig(
		resourceFirewallRulesConfig{
			BlockName:         blockName,
			SupportBlockName:  supportBlockName,
			SupportBlockName2: supportBlockName2,
			VPCName:           vpcName,
		},
		resourceFirewallRulesUpdateConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing update config template data: %e", err)
	}

	configUpdate2, err := parsedAccConfig(
		resourceFirewallRulesConfig{
			BlockName:         blockName,
			SupportBlockName:  supportBlockName,
			SupportBlockName2: supportBlockName2,
			VPCName:           vpcName,
		},
		resourceFirewallRulesUpdateConfigTpl2,
	)
	if err != nil {
		t.Errorf("error parsing update config 2 template data: %e", err)
	}

	configUpdate3, err := parsedAccConfig(
		resourceFirewallRulesConfig{
			BlockName:         blockName,
			SupportBlockName:  supportBlockName,
			SupportBlockName2: supportBlockName2,
			VPCName:           vpcName,
		},
		resourceFirewallRulesUpdateConfigTpl3,
	)
	if err != nil {
		t.Errorf("error parsing update config 3 template data: %e", err)
	}

	configUpdate4, err := parsedAccConfig(
		resourceFirewallRulesConfig{
			BlockName:         blockName,
			SupportBlockName:  supportBlockName,
			SupportBlockName2: supportBlockName2,
			VPCName:           vpcName,
		},
		resourceFirewallRulesUpdateConfigTpl4,
	)
	if err != nil {
		t.Errorf("error parsing update config 4 template data: %e", err)
	}

	configUpdate5, err := parsedAccConfig(
		resourceFirewallRulesConfig{
			BlockName:         blockName,
			SupportBlockName:  supportBlockName,
			SupportBlockName2: supportBlockName2,
			VPCName:           vpcName,
		},
		resourceFirewallRulesUpdateConfigTpl5,
	)
	if err != nil {
		t.Errorf("error parsing update config 5 template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy:             testAccFirewallRulesDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkResourceFirewallRules(resourceName, vpcName),
			},
			{
				Config: configUpdate,
				Check:  checkResourceFirewallRulesUpdate(resourceName),
			},
			{
				Config: configUpdate2,
				Check:  checkResourceFirewallRulesUpdate2(resourceName, vpcName),
			},
			{
				Config: configUpdate3,
				Check:  checkResourceFirewallRulesUpdate3(resourceName),
			},
			{
				Config: configUpdate4,
				Check:  checkResourceFirewallRulesUpdate4(resourceName),
			},
			{
				Config: configUpdate5,
				Check:  checkResourceFirewallRulesUpdate5(resourceName),
			},
		},
	})
}

func TestAccCloudResourceFirewallRules_v1_upgrade(t *testing.T) {
	blockName := newBlockName("firewall_rules")
	supportBlockName := newBlockName("support")
	supportBlockName2 := newBlockName("support")
	vpcName := newResourceName()
	resourceName := fmt.Sprintf("oxide_vpc_firewall_rules.%s", blockName)
	configV1, err := parsedAccConfig(
		resourceFirewallRulesConfig{
			BlockName:         blockName,
			SupportBlockName:  supportBlockName,
			SupportBlockName2: supportBlockName2,
			VPCName:           vpcName,
		},
		resourceFirewallRulesUpdateConfigTplV1,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	configV2, err := parsedAccConfig(
		resourceFirewallRulesConfig{
			BlockName:         blockName,
			SupportBlockName:  supportBlockName,
			SupportBlockName2: supportBlockName2,
			VPCName:           vpcName,
		},
		resourceFirewallRulesUpdateConfigTpl5,
	)
	if err != nil {
		t.Errorf("error parsing update config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccFirewallRulesDestroy,
		Steps: []resource.TestStep{
			// Initial state with v1 of oxide_vpc_firewall_rules.
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"oxide": {
						Source:            "oxidecomputer/oxide",
						VersionConstraint: "0.14.1",
					},
				},
				Config: configV1,
				Check:  checkResourceFirewallRulesV1(resourceName),
			},
			// Upgrade provider to use the latest version of oxide_vpc_firewall_rules.
			{
				ExternalProviders:        map[string]resource.ExternalProvider{},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
				Config:                   configV2,
				Check:                    checkResourceFirewallRulesUpdate5(resourceName),
			},
		},
	})
}

func checkResourceFirewallRules(resourceName, vpcName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		// We only check that these are set as we cannot guarantee order
		resource.TestCheckResourceAttrSet(resourceName, "rules.custom-deny-http.action"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.custom-deny-http.description"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.custom-deny-http.direction"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.custom-deny-http.filters.hosts.0.type",
			"vpc",
		),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.custom-deny-http.filters.hosts.0.value",
			vpcName,
		),
		resource.TestCheckResourceAttrSet(resourceName, "rules.custom-deny-http.name"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.custom-deny-http.name",
			"custom-deny-http",
		),
		resource.TestCheckResourceAttrSet(resourceName, "rules.custom-deny-http.priority"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.custom-deny-http.status"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.custom-deny-http.targets.0.type",
			"subnet",
		),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.custom-deny-http.targets.0.value",
			"default",
		),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-internal-inbound.action"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-internal-inbound.description"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-internal-inbound.direction"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-internal-inbound.filters.hosts.0.type",
			"vpc",
		),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-internal-inbound.filters.hosts.0.value",
			vpcName,
		),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-internal-inbound.name"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-internal-inbound.name",
			"allow-internal-inbound",
		),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-internal-inbound.priority"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-internal-inbound.status"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-internal-inbound.targets.0.type",
			"subnet",
		),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-internal-inbound.targets.0.value",
			"default",
		),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func checkResourceFirewallRulesUpdate(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		// Rule 1.
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.action", "allow"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-https.description",
			"Allow HTTPS.",
		),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.name", "allow-https"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.priority", "50"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.status", "enabled"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-https.filters.hosts.0.type"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-https.filters.hosts.0.value"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.filters.ports.0", "443"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-https.filters.protocols.0.type",
			"tcp",
		),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.targets.0.type", "subnet"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-https.targets.0.value"),
		// Rule 2.
		resource.TestCheckResourceAttr(resourceName, "rules.allow-ssh.action", "allow"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-ssh.description", "Allow SSH."),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-ssh.name", "allow-ssh"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-ssh.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-ssh.priority", "50"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-ssh.status", "enabled"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-ssh.filters.hosts.0.type"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-ssh.filters.hosts.0.value"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-ssh.filters.ports.0", "22"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-ssh.filters.protocols.0.type",
			"tcp",
		),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-ssh.targets.0.type", "subnet"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-ssh.targets.0.value"),
	}...)
}

func checkResourceFirewallRulesUpdate2(resourceName, vpcName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		// Rule 1.
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.action", "allow"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-https.description",
			"Allow HTTPS.",
		),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.name", "allow-https"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.priority", "50"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.status", "enabled"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-https.filters.hosts.0.type"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-https.filters.hosts.0.value"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.filters.ports.0", "443"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-https.filters.protocols.0.type",
			"tcp",
		),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-https.targets.0.type", "subnet"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-https.targets.0.value"),
	}...)
}

func checkResourceFirewallRulesUpdate3(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttr(resourceName, "rules.#", "0"),
	}...)
}

func checkResourceFirewallRulesUpdate4(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		// Rule 1.
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.action", "allow"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.description", "ICMP rule."),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.name", "allow-icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.priority", "65535"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.status", "enabled"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-icmp.filters.protocols.0.type",
			"icmp",
		),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-icmp.filters.protocols.0.icmp_type",
			"3",
		),
		resource.TestCheckNoResourceAttr(
			resourceName,
			"rules.allow-icmp.filters.protocols.0.icmp_code",
		),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.targets.0.type", "subnet"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-icmp.targets.0.value"),
	}...)
}

func checkResourceFirewallRulesUpdate5(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		// Rule 1.
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.action", "allow"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.description", "ICMP rule."),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.name", "allow-icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.priority", "65535"),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.status", "enabled"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-icmp.filters.protocols.0.type",
			"icmp",
		),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-icmp.filters.protocols.0.icmp_type",
			"0",
		),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.allow-icmp.filters.protocols.0.icmp_code",
			"1-3",
		),
		resource.TestCheckResourceAttr(resourceName, "rules.allow-icmp.targets.0.type", "subnet"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.allow-icmp.targets.0.value"),
	}...)
}

func checkResourceFirewallRulesV1(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttrSet(resourceName, "time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "time_modified"),
		// Rule 1.
		resource.TestCheckResourceAttr(resourceName, "rules.0.action", "allow"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.name", "allow-icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.description", "ICMP rule."),
		resource.TestCheckResourceAttr(resourceName, "rules.0.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "65535"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.status", "enabled"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.protocols.0.type", "icmp"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.protocols.0.icmp_type", "0"),
		resource.TestCheckResourceAttr(
			resourceName,
			"rules.0.filters.protocols.0.icmp_code",
			"1-3",
		),
		resource.TestCheckResourceAttr(resourceName, "rules.0.targets.0.type", "subnet"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.targets.0.value"),
	}...)
}

func testAccFirewallRulesDestroy(s *terraform.State) error {
	client, err := newTestClient()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "oxide_vpc_firewall_rules" {
			continue
		}

		params := oxide.VpcFirewallRulesViewParams{
			Vpc: oxide.NameOrId(rs.Primary.Attributes["vpc_id"]),
		}

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, time.Minute)
		defer cancel()

		res, err := client.VpcFirewallRulesView(ctx, params)
		if err != nil && is404(err) {
			continue
		}

		if len(res.Rules) < 1 {
			continue
		}

		return fmt.Errorf("vpc firewall rules for VPC (%v) still exist", &res.Rules[0].VpcId)
	}

	return nil
}

func TestFirewallRules_name(t *testing.T) {
	testCases := []struct {
		name  string
		value string
		valid bool
	}{
		{
			name:  "valid name",
			value: "my-rule-01",
			valid: true,
		},
		{
			name:  "valid long name",
			value: strings.Repeat("a", 63),
			valid: true,
		},
		{
			name:  "no caps as first char",
			value: "My-rule-01",
			valid: false,
		},
		{
			name:  "no non-ASCII",
			value: "my-ruleš-01",
			valid: false,
		},
		{
			name:  "no special characters other than dash",
			value: "my_rule_01",
			valid: false,
		},
		{
			name:  "no dash at the end",
			value: "my-rule-01-",
			valid: false,
		},
		{
			name:  "no more than 63 chars",
			value: strings.Repeat("a", 64),
			valid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.valid, vpcFirewallRuleNameRegexp.MatchString(tc.value))
		})
	}
}

func TestFirewallRules_v0_upgrade(t *testing.T) {
	var version atomic.Int32
	version.Store(15)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var respFile string
		switch version.Load() {
		case 15:
			respFile = "r15_get_vpc_firewall_rules.json"
		case 16:
			respFile = "r16_get_vpc_firewall_rules.json"
		}

		respPath := filepath.Join("test-fixtures", "resource_vpc_firewall_rules", respFile)
		f, err := testFixtures.ReadFile(respPath)
		if err != nil {
			t.Errorf("failed to read file %q: %v", respPath, err)
			http.Error(w, "failed to read file", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(f)
	}))
	t.Cleanup(func() {
		ts.Close()
	})

	//lintignore:AT004 // Provider must connect to test server.
	configV0 := fmt.Sprintf(`
provider "oxide" {
  host  = "%s"
  token = "fake"
}

resource "oxide_vpc_firewall_rules" "test" {
  vpc_id = "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  rules = [
    {
      action      = "allow"
      name        = "single-protocol"
      description = "Single protocol"
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        ports     = ["22"]
        protocols = ["TCP"]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    },
    {
      action      = "allow"
      name        = "multiple-protocols"
      description = "Multiple protocols"
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        ports     = ["22"]
        protocols = ["TCP", "UDP"]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    },
    {
      action      = "allow"
      name        = "no-protocol"
      description = "No protocol"
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        ports = ["22"]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    }
  ]
}
`, ts.URL)

	//lintignore:AT004 // Provider must connect to test server.
	configV1 := fmt.Sprintf(`
provider "oxide" {
  host  = "%s"
  token = "fake"
}

resource "oxide_vpc_firewall_rules" "test" {
  vpc_id = "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  rules = [
    {
      action      = "allow"
      name        = "single-protocol"
      description = "Single protocol"
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        ports = ["22"]
        protocols = [
          { type : "tcp" }
        ]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    },
    {
      action      = "allow"
      name        = "multiple-protocols"
      description = "Multiple protocols"
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        ports = ["22"]
        protocols = [
          { type : "tcp" },
          { type : "udp" }
        ]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    },
    {
      action      = "allow"
      name        = "no-protocol"
      description = "No protocol"
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        ports = ["22"]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    }
  ]
}
`, ts.URL)

	//lintignore:AT004 // Provider must connect to test server.
	configLatest := fmt.Sprintf(`
provider "oxide" {
  host  = "%s"
  token = "fake"
}

resource "oxide_vpc_firewall_rules" "test" {
  vpc_id = "3fa85f64-5717-4562-b3fc-2c963f66afa6"
  rules = {
    single-protocol = {
      action      = "allow"
      description = "Single protocol"
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        ports = ["22"]
        protocols = [
          { type : "tcp" }
        ]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    },
    multiple-protocols = {
      action      = "allow"
      description = "Multiple protocols"
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        ports = ["22"]
        protocols = [
          { type : "tcp" },
          { type : "udp" }
        ]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    },
    no-protocol = {
      action      = "allow"
      description = "No protocol"
      direction   = "inbound"
      priority    = 65535
      status      = "enabled"
      filters = {
        ports = ["22"]
      }
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    }
  }
}
`, ts.URL)

	resource.UnitTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Initial state with v0.12.0 and R15.
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"oxide": {
						Source:            "oxidecomputer/oxide",
						VersionConstraint: "0.12.0",
					},
				},
				PreConfig: func() {
					version.Store(15)
				},
				Config: configV0,
			},
			// Upgrade to current version and R16.
			{
				ExternalProviders:        map[string]resource.ExternalProvider{},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
				PreConfig: func() {
					version.Store(16)
				},
				Config: configLatest,
			},
		},
	})

	resource.UnitTest(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Initial state with v0.13.0 and R16.
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"oxide": {
						Source:            "oxidecomputer/oxide",
						VersionConstraint: "0.13.0",
					},
				},
				PreConfig: func() {
					version.Store(16)
				},
				Config: configV1,
			},
			// Upgrade to current version.
			{
				ExternalProviders:        map[string]resource.ExternalProvider{},
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
				PreConfig: func() {
					version.Store(16)
				},
				Config: configLatest,
			},
		},
	})
}
