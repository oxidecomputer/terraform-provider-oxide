// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceAntiAffinityGroupConfig struct {
	BlockName         string
	Name              string
	SupportBlockName  string
	SupportBlockName2 string
}

var dataSourceAntiAffinityGroupConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_anti_affinity_group" "{{.SupportBlockName2}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  name        = "{{.Name}}"
  description = "a group"
  policy      = "allow"
}

data "oxide_anti_affinity_group" "{{.BlockName}}" {
  project_name = "tf-acc-test"
  name         = oxide_anti_affinity_group.{{.SupportBlockName2}}.name
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceAntiAffinityGroup_full(t *testing.T) {
	blockName := newBlockName("datasource-anti-affinity-group")
	resourceName := newResourceName()
	config, err := parsedAccConfig(
		dataSourceAntiAffinityGroupConfig{
			BlockName:         blockName,
			SupportBlockName:  newBlockName("support"),
			SupportBlockName2: newBlockName("support"),
			Name:              resourceName,
		},
		dataSourceAntiAffinityGroupConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: checkDataSourceAntiAffinityGroup(
					fmt.Sprintf("data.oxide_anti_affinity_group.%s", blockName),
					resourceName,
				),
			},
		},
	})
}

func checkDataSourceAntiAffinityGroup(dataName, keyName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "name", keyName),
		resource.TestCheckResourceAttr(dataName, "description", "a group"),
		resource.TestCheckResourceAttr(dataName, "policy", "allow"),
		resource.TestCheckResourceAttr(dataName, "failure_domain", "sled"),
		resource.TestCheckResourceAttrSet(dataName, "project_id"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}
