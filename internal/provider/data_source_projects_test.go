// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceProjectsConfig struct {
	BlockName string
}

var dataSourceProjectsConfigTpl = `
data "oxide_projects" "{{.BlockName}}" {
  timeouts = {
    read = "1m"
  }
}
`

func TestAccDataSourceProjects_full(t *testing.T) {
	blockName := newBlockName("datasource-projects")
	config, err := parsedAccConfig(
		dataSourceProjectsConfig{
			BlockName: blockName,
		},
		dataSourceProjectsConfigTpl,
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
				Check: checkDataSourceProjects(
					fmt.Sprintf("data.oxide_projects.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceProjects(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.id"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.name"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.time_created"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.time_modified"),
	}...)
}
