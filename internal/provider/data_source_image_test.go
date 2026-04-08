// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

type dataSourceImageConfig struct {
	BlockName         string
	SupportBlockName  string
	SupportBlockName2 string
}

var dataSourceImageConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

data "oxide_images" "{{.SupportBlockName2}}" {
  project_id = data.oxide_project.{{.SupportBlockName}}.id
}

data "oxide_image" "{{.BlockName}}" {
  project_name = "tf-acc-test"
  name = element(tolist(data.oxide_images.{{.SupportBlockName2}}.images[*].name), 0)
  timeouts = {
    read = "1m"
  }
}
`

// NB: The project must be populated with at least one image for this test to pass
func TestAccCloudDataSourceImage_full(t *testing.T) {
	blockName := sharedtest.NewBlockName("datasource-image")
	config, err := sharedtest.ParsedAccConfig(
		dataSourceImageConfig{
			BlockName:         blockName,
			SupportBlockName:  sharedtest.NewBlockName("support"),
			SupportBlockName2: sharedtest.NewBlockName("support"),
		},
		dataSourceImageConfigTpl,
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
				Check: checkDataSourceImage(
					fmt.Sprintf("data.oxide_image.%s", blockName),
				),
			},
		},
	})
}

func checkDataSourceImage(dataName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
		resource.TestCheckResourceAttrSet(dataName, "block_size"),
		resource.TestCheckResourceAttrSet(dataName, "description"),
		resource.TestCheckResourceAttrSet(dataName, "os"),
		resource.TestCheckResourceAttrSet(dataName, "project_id"),
		resource.TestCheckResourceAttrSet(dataName, "project_name"),
		resource.TestCheckResourceAttrSet(dataName, "name"),
		resource.TestCheckResourceAttrSet(dataName, "size"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttrSet(dataName, "version"),
	}...)
}
