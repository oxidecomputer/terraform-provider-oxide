// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceDiskConfig struct {
	BlockName         string
	DiskName          string
	SupportBlockName  string
	SupportBlockName2 string
}

var dataSourceDiskConfigTpl = `
data "oxide_project" "{{.SupportBlockName}}" {
	name = "tf-acc-test"
}

resource "oxide_disk" "{{.SupportBlockName2}}" {
  project_id  = data.oxide_project.{{.SupportBlockName}}.id
  description = "a test disk for data source"
  name        = "{{.DiskName}}"
  size        = 1073741824
  block_size  = 512
}

data "oxide_disk" "{{.BlockName}}" {
  project_name = data.oxide_project.{{.SupportBlockName}}.name
  name         = oxide_disk.{{.SupportBlockName2}}.name
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceDisk_full(t *testing.T) {
	blockName := newBlockName("datasource-disk")
	diskName := newResourceName()
	config, err := parsedAccConfig(
		dataSourceDiskConfig{
			BlockName:         blockName,
			DiskName:          diskName,
			SupportBlockName:  newBlockName("support"),
			SupportBlockName2: newBlockName("support-disk"),
		},
		dataSourceDiskConfigTpl,
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
				Check: checkDataSourceDisk(
					fmt.Sprintf("data.oxide_disk.%s", blockName),
					diskName,
				),
			},
		},
	})
}

func checkDataSourceDisk(dataName, diskName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "name", diskName),
		resource.TestCheckResourceAttr(dataName, "description", "a test disk for data source"),
		resource.TestCheckResourceAttr(dataName, "size", "1073741824"),
		resource.TestCheckResourceAttr(dataName, "block_size", "512"),
		resource.TestCheckResourceAttr(dataName, "device_path", "/mnt/"+diskName),
		resource.TestCheckResourceAttrSet(dataName, "project_id"),
		resource.TestCheckResourceAttr(dataName, "state.state", "detached"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}
