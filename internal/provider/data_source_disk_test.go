// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

type dataSourceDiskConfig struct {
	DiskName string
}

var dataSourceDiskConfigTpl = `
data "oxide_project" "test" {
	name = "tf-acc-test"
}

resource "oxide_disk" "test" {
  project_id  = data.oxide_project.test.id
  description = "a test disk for data source"
  name        = "{{.DiskName}}"
  size        = 1073741824
  block_size  = 512
}

data "oxide_disk" "test" {
  project_name = data.oxide_project.test.name
  name         = oxide_disk.test.name
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceDisk_full(t *testing.T) {
	diskName := newResourceName()
	config, err := parsedAccConfig(
		dataSourceDiskConfig{
			DiskName: diskName,
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
				Check:  checkDataSourceDisk("data.oxide_disk.test", diskName),
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
		resource.TestCheckResourceAttr(dataName, "disk_type", "distributed"),
		resource.TestCheckResourceAttr(dataName, "read_only", "false"),
		resource.TestCheckResourceAttrSet(dataName, "project_id"),
		resource.TestCheckResourceAttr(dataName, "state.state", "detached"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}
