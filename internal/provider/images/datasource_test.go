// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package images_test

import (
	"fmt"
	"testing"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceProjectConfig struct {
	BlockName        string
	SupportBlockName string
}

type dataSourceSiloConfig struct {
	BlockName string
}

var dataSourceProjectConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

data "oxide_images" "{{.BlockName}}" {
  project_id = data.oxide_project.{{.SupportBlockName}}.id
  timeouts = {
    read = "1m"
  }
}
`

var dataSourceSiloConfigTpl = `
data "oxide_images" "{{.BlockName}}" {}
`

// NB: The project must be populated with at least one image for this test to pass
func TestAccCloudDataSourceImages_project(t *testing.T) {
	blockName := sharedtest.NewBlockName("datasource-images")
	config, err := sharedtest.ParsedAccConfig(
		dataSourceProjectConfig{
			BlockName:        blockName,
			SupportBlockName: sharedtest.NewBlockName("support"),
		},
		dataSourceProjectConfigTpl,
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
				Check: checkDataSourceProject(
					fmt.Sprintf("data.oxide_images.%s", blockName),
				),
			},
		},
	})
}

// NB: The silo must be populated with at least one image for this test to pass
func TestAccCloudDataSourceImages_silo(t *testing.T) {
	blockName := sharedtest.NewBlockName("datasource-images")
	config, err := sharedtest.ParsedAccConfig(
		dataSourceSiloConfig{
			BlockName: blockName,
		},
		dataSourceSiloConfigTpl,
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
				Check: checkDataSourceSilo(
					fmt.Sprintf("data.oxide_images.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceProject(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.block_size"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.description"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.os"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.id"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.name"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.size"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.time_created"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.time_modified"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.version"),
	}...)
}

func checkDataSourceSilo(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.block_size"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.id"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.name"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.size"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.time_created"),
		resource.TestCheckResourceAttrSet(dataName, "images.0.time_modified"),
	}...)
}
