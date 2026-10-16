// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcfirewallrules_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

type dataSourceConfig struct {
	VPCName string
}

var dataSourceConfigTpl = resourceConfigTpl + `
data "oxide_vpc_firewall_rules" "test" {
  project = "tf-acc-test"
  vpc     = oxide_vpc.test_vpc.name

  depends_on = [oxide_vpc_firewall_rules.test]

  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceVPCFirewallRules_full(t *testing.T) {
	vpcName := sharedtest.NewResourceName()
	config := sharedtest.ParsedAccConfig(t,
		dataSourceConfig{VPCName: vpcName},
		dataSourceConfigTpl,
	)
	dataSourceName := "data.oxide_vpc_firewall_rules.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "project", "tf-acc-test"),
					resource.TestCheckResourceAttr(dataSourceName, "vpc", vpcName),
					resource.TestCheckResourceAttr(dataSourceName, "rules.#", "2"),
					resource.TestMatchTypeSetElemNestedAttrs(
						dataSourceName,
						"rules.*",
						map[string]*regexp.Regexp{
							"action":        regexp.MustCompile("^deny$"),
							"description":   regexp.MustCompile("^custom deny$"),
							"direction":     regexp.MustCompile("^inbound$"),
							"id":            regexp.MustCompile(".+"),
							"name":          regexp.MustCompile("^custom-deny-http$"),
							"priority":      regexp.MustCompile("^50$"),
							"status":        regexp.MustCompile("^enabled$"),
							"time_created":  regexp.MustCompile(".+"),
							"time_modified": regexp.MustCompile(".+"),
							"vpc_id":        regexp.MustCompile(".+"),
						},
					),
					resource.TestCheckResourceAttr(dataSourceName, "timeouts.read", "1m"),
				),
			},
		},
	})
}
