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
	project_id        = data.oxide_project.{{.SupportBlockName}}.id
	description       = "a test vpc"
	name              = "{{.VPCName}}"
	dns_name          = "my-vpc-dns"
}

resource "oxide_vpc_firewall_rules" "{{.BlockName}}" {
	vpc_id = oxide_vpc.{{.SupportBlockName2}}.id
	rules = [
	   {
		 action = "deny"
		 description = "custom deny"
		 name = "custom-deny-http"
		 direction = "inbound"
		 priority = 50
		 status = "enabled"
		 filters = {
		   hosts = [
			 {
				type = "vpc"
				value = oxide_vpc.{{.SupportBlockName2}}.name
			 }
		   ]
		   ports = ["8123"]
		   protocols = ["ICMP"]
		 },
		 targets = [
		   {
			  type = "subnet"
			  value = "default"
		   }
		 ]
	   },
	   {
		 action = "allow"
		 name = "allow-internal-inbound"
		 description = "custom allow"
		 direction = "inbound"
		 priority = 65534
		 status = "enabled"
		 filters = {
		   hosts = [
			 {
				type = "vpc"
				value = oxide_vpc.{{.SupportBlockName2}}.name
			 }
		   ]
		   ports = []
		   protocols = []
		 }
		 targets = [
		   {
			  type = "subnet"
			  value = "default"
		   }
		 ]
	   }
	]
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
	project_id        = data.oxide_project.{{.SupportBlockName}}.id
	description       = "a test vpc"
	name              = "{{.VPCName}}"
	dns_name          = "my-vpc-dns"
}

resource "oxide_vpc_firewall_rules" "{{.BlockName}}" {
	vpc_id = oxide_vpc.{{.SupportBlockName2}}.id
	rules = [
	   {
		 action = "deny"
		 description = "custom deny"
		 name = "custom-deny-http"
		 direction = "inbound"
		 priority = 50
		 status = "enabled"
		 filters = {
		   hosts = [
			 {
				type = "vpc"
				value = oxide_vpc.{{.SupportBlockName2}}.name
			 }
		   ]
		   ports = ["8123"]
		   protocols = ["ICMP"]
		 },
		 targets = [
		   {
			  type = "subnet"
			  value = "default"
		   }
		 ]
	   }
	]
}
`

var resourceFirewallRulesUpdateConfigTpl2 = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_vpc" "{{.SupportBlockName2}}" {
	project_id        = data.oxide_project.{{.SupportBlockName}}.id
	description       = "a test vpc"
	name              = "{{.VPCName}}"
	dns_name          = "my-vpc-dns"
}

resource "oxide_vpc_firewall_rules" "{{.BlockName}}" {
	vpc_id = oxide_vpc.{{.SupportBlockName2}}.id
	rules = [
	   {
		 action = "allow"
		 description = "custom allow"
		 name = "custom-allow-http"
		 direction = "outbound"
		 priority = 40
		 status = "disabled"
		 filters = {
		   hosts = [
			 {
				type = "vpc"
				value = oxide_vpc.{{.SupportBlockName2}}.name
			 }
		   ]
		   ports = ["8124"]
		   protocols = ["TCP"]
		 },
		 targets = [
		   {
			  type = "subnet"
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
		t.Errorf("error parsing config template data: %e", err)
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
		t.Errorf("error parsing config template data: %e", err)
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
				Check:  checkResourceFirewallRulesUpdate(resourceName, vpcName),
			},
			{
				Config: configUpdate2,
				Check:  checkResourceFirewallRulesUpdate2(resourceName, vpcName),
			},
		},
	})
}

func checkResourceFirewallRules(resourceName, vpcName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		// We only check that these are set as we cannot guarantee order
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.action"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.description"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.direction"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.hosts.0.type", "vpc"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.hosts.0.value", vpcName),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.id"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.name"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.priority"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.status"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.targets.0.type", "subnet"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.targets.0.value", "default"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.time_modified"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.1.action"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.1.description"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.1.direction"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.filters.hosts.0.type", "vpc"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.filters.hosts.0.value", vpcName),
		resource.TestCheckResourceAttrSet(resourceName, "rules.1.id"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.1.name"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.1.priority"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.1.status"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.targets.0.type", "subnet"),
		resource.TestCheckResourceAttr(resourceName, "rules.1.targets.0.value", "default"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.1.time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.1.time_modified"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.delete", "2m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.create", "3m"),
		resource.TestCheckResourceAttr(resourceName, "timeouts.update", "4m"),
	}...)
}

func checkResourceFirewallRulesUpdate(resourceName, vpcName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.action", "deny"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.description", "custom deny"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.direction", "inbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.hosts.0.type", "vpc"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.hosts.0.value", vpcName),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.ports.0", "8123"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.protocols.0", "ICMP"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.id"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.name", "custom-deny-http"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "50"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.status", "enabled"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.targets.0.type", "subnet"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.targets.0.value", "default"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.time_modified"),
	}...)
}

func checkResourceFirewallRulesUpdate2(resourceName, vpcName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(resourceName, "id"),
		resource.TestCheckResourceAttrSet(resourceName, "vpc_id"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.action", "allow"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.description", "custom allow"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.direction", "outbound"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.hosts.0.type", "vpc"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.hosts.0.value", vpcName),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.ports.0", "8124"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.filters.protocols.0", "TCP"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.id"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.name", "custom-allow-http"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.priority", "40"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.status", "disabled"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.targets.0.type", "subnet"),
		resource.TestCheckResourceAttr(resourceName, "rules.0.targets.0.value", "default"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.time_created"),
		resource.TestCheckResourceAttrSet(resourceName, "rules.0.time_modified"),
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
