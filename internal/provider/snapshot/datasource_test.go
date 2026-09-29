// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package snapshot_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

type dataSourceConfig struct {
	DiskName     string
	SnapshotName string
}

var dataSourceConfigTpl = `
data "oxide_project" "test" {
  name = "tf-acc-test"
}

resource "oxide_disk" "test" {
  project_id  = data.oxide_project.test.id
  description = "a test disk for snapshot data source"
  name        = "{{.DiskName}}"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_snapshot" "test" {
  project_id  = data.oxide_project.test.id
  description = "a test snapshot for data source"
  name        = "{{.SnapshotName}}"
  disk_id     = oxide_disk.test.id
}

data "oxide_snapshot" "test" {
  project_name = data.oxide_project.test.name
  name         = oxide_snapshot.test.name
  timeouts = {
    read = "1m"
  }
}
`

func TestAccCloudDataSourceSnapshot_full(t *testing.T) {
	diskName := sharedtest.NewResourceName()
	snapshotName := sharedtest.NewResourceName()
	config := sharedtest.ParsedAccConfig(t,
		dataSourceConfig{
			DiskName:     diskName,
			SnapshotName: snapshotName,
		},
		dataSourceConfigTpl,
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  checkDataSource("data.oxide_snapshot.test", snapshotName),
			},
		},
	})
}

func checkDataSource(dataName, snapshotName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc([]resource.TestCheckFunc{
		resource.TestCheckResourceAttrSet(dataName, "id"),
		resource.TestCheckResourceAttr(dataName, "name", snapshotName),
		resource.TestCheckResourceAttr(dataName, "description", "a test snapshot for data source"),
		resource.TestCheckResourceAttrSet(dataName, "disk_id"),
		resource.TestCheckResourceAttrSet(dataName, "project_id"),
		resource.TestCheckResourceAttr(dataName, "project_name", "tf-acc-test"),
		resource.TestCheckResourceAttr(dataName, "size", "1073741824"),
		resource.TestCheckResourceAttrSet(dataName, "state"),
		resource.TestCheckResourceAttrSet(dataName, "time_created"),
		resource.TestCheckResourceAttrSet(dataName, "time_modified"),
		resource.TestCheckResourceAttr(dataName, "timeouts.read", "1m"),
	}...)
}
