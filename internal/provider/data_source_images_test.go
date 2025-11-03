// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceProjectImagesConfig struct {
	BlockName        string
	SupportBlockName string
}

type dataSourceSiloImagesConfig struct {
	BlockName string
}

var dataSourceProjectImagesConfigTpl = `
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

var dataSourceSiloImagesConfigTpl = `
data "oxide_images" "{{.BlockName}}" {}
`

// NB: The project must be populated with at least one image for this test to pass
func TestAccCloudDataSourceImages_project(t *testing.T) {
	blockName := newBlockName("datasource-images")
	config, err := parsedAccConfig(
		dataSourceProjectImagesConfig{
			BlockName:        blockName,
			SupportBlockName: newBlockName("support"),
		},
		dataSourceProjectImagesConfigTpl,
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
				Check: checkDataSourceProjectImages(
					fmt.Sprintf("data.oxide_images.%s", blockName),
				),
			},
		},
	})
}

// NB: The silo must be populated with at least one image for this test to pass
func TestAccCloudDataSourceImages_silo(t *testing.T) {
	blockName := newBlockName("datasource-images")
	config, err := parsedAccConfig(
		dataSourceSiloImagesConfig{
			BlockName: blockName,
		},
		dataSourceSiloImagesConfigTpl,
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
				Check: checkDataSourceSiloImages(
					fmt.Sprintf("data.oxide_images.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceProjectImages(dataName string) resource.TestCheckFunc {
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

func checkDataSourceSiloImages(dataName string) resource.TestCheckFunc {
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
