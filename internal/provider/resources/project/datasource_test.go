// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package project_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider"
)

type dataSourceProjectConfig struct {
	BlockName        string
	SupportBlockName string
}

var dataSourceProjectConfigTpl = `
data "oxide_project" "{{.BlockName}}" {
  name = "tf-acc-test"
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceProject_full(t *testing.T) {
	blockName := provider.NewBlockName("datasource-project")
	config, err := provider.ParsedAccConfig(
		dataSourceProjectConfig{
			BlockName:        blockName,
			SupportBlockName: provider.NewBlockName("support"),
		},
		dataSourceProjectConfigTpl,
	)
	if err != nil {
		t.Errorf(
			"error parsing config template data: %e", err,
		)
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { provider.PreCheck(t) },
		ProtoV6ProviderFactories: provider.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: checkDataSourceProject(
					fmt.Sprintf(
						"data.oxide_project.%s", blockName,
					),
				),
			},
		},
	})
}

func checkDataSourceProject(
	dataName string,
) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		[]resource.TestCheckFunc{
			resource.TestCheckResourceAttrSet(
				dataName, "name",
			),
			resource.TestCheckResourceAttr(
				dataName, "timeouts.read", "1m",
			),
			resource.TestCheckResourceAttrSet(
				dataName, "id",
			),
			resource.TestCheckResourceAttrSet(
				dataName, "description",
			),
			resource.TestCheckResourceAttrSet(
				dataName, "time_created",
			),
			resource.TestCheckResourceAttrSet(
				dataName, "time_modified",
			),
		}...,
	)
}
