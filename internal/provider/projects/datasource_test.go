// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package projects_test

import (
	"fmt"
	"testing"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceConfig struct {
	BlockName string
}

var dataSourceConfigTpl = `
data "oxide_projects" "{{.BlockName}}" {
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceProjects_full(t *testing.T) {
	blockName := sharedtest.NewBlockName("datasource-projects")
	config, err := sharedtest.ParsedAccConfig(
		dataSourceConfig{
			BlockName: blockName,
		},
		dataSourceConfigTpl,
	)
	if err != nil {
		t.Errorf("error parsing config template data: %e", err)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: checkDataSource(
					fmt.Sprintf("data.oxide_projects.%s", blockName),
				),
			},
		},
	})
}

func checkDataSource(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.id"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.name"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.time_created"),
		resource.TestCheckResourceAttrSet(dataName, "projects.0.time_modified"),
	}...)
}
