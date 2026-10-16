// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package floatingips_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

type dataSourceConfig struct {
	BlockName        string
	ResourceName     string
	SupportBlockName string
}

var dataSourceConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
  name = "tf-acc-test"
}

resource "oxide_floating_ip" "{{.BlockName}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  name        = "{{.ResourceName}}"
  description = "Floating IP from list data source test."
  ip_version  = "v4"
}

data "oxide_floating_ips" "{{.BlockName}}" {
  project = data.oxide_project.{{.SupportBlockName}}.name
  timeouts = {
    read = "1m"
  }

  depends_on = [oxide_floating_ip.{{.BlockName}}]
}

locals {
  listed_floating_ip = one([
    for floating_ip in data.oxide_floating_ips.{{.BlockName}}.floating_ips : floating_ip
    if floating_ip.id == oxide_floating_ip.{{.BlockName}}.id
  ])
}

output "listed_floating_ip_name" {
  value = local.listed_floating_ip.name
}

output "listed_floating_ip_description" {
  value = local.listed_floating_ip.description
}

output "listed_floating_ip_ip" {
  value = local.listed_floating_ip.ip
}

output "listed_floating_ip_instance_id" {
  value = local.listed_floating_ip.instance_id
}

output "listed_floating_ip_ip_pool_id" {
  value = local.listed_floating_ip.ip_pool_id
}

output "listed_floating_ip_project_id" {
  value = local.listed_floating_ip.project_id
}

output "listed_floating_ip_time_created" {
  value = local.listed_floating_ip.time_created
}

output "listed_floating_ip_time_modified" {
  value = local.listed_floating_ip.time_modified
}
`

func TestAccCloudDataSourceFloatingIPs_full(t *testing.T) {
	blockName := sharedtest.NewBlockName("datasource-floating-ips")
	resourceName := sharedtest.NewResourceName()
	config := sharedtest.ParsedAccConfig(t,
		dataSourceConfig{
			BlockName:        blockName,
			ResourceName:     resourceName,
			SupportBlockName: sharedtest.NewBlockName("support"),
		},
		dataSourceConfigTpl,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						fmt.Sprintf("data.oxide_floating_ips.%s", blockName),
						"id",
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("data.oxide_floating_ips.%s", blockName),
						"project",
						"tf-acc-test",
					),
					resource.TestCheckResourceAttr(
						fmt.Sprintf("data.oxide_floating_ips.%s", blockName),
						"timeouts.read",
						"1m",
					),
					resource.TestCheckOutput("listed_floating_ip_name", resourceName),
					resource.TestCheckOutput(
						"listed_floating_ip_description",
						"Floating IP from list data source test.",
					),
					resource.TestMatchOutput("listed_floating_ip_ip", regexp.MustCompile(`.+`)),
					resource.TestCheckOutput("listed_floating_ip_instance_id", ""),
					resource.TestMatchOutput(
						"listed_floating_ip_ip_pool_id",
						regexp.MustCompile(`.+`),
					),
					resource.TestMatchOutput(
						"listed_floating_ip_project_id",
						regexp.MustCompile(`.+`),
					),
					resource.TestMatchOutput(
						"listed_floating_ip_time_created",
						regexp.MustCompile(`.+`),
					),
					resource.TestMatchOutput(
						"listed_floating_ip_time_modified",
						regexp.MustCompile(`.+`),
					),
				),
			},
		},
	})
}
