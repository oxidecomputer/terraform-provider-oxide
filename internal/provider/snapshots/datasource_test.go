// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package snapshots_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccCloudDataSourceSnapshots_full(t *testing.T) {
	const dataSourceName = "data.oxide_snapshots.test"
	snapshotName := sharedtest.NewResourceName()
	diskName := sharedtest.NewResourceName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceConfig(snapshotName, diskName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "project", "tf-acc-test"),
					resource.TestCheckResourceAttr(dataSourceName, "timeouts.read", "1m"),
					checkSnapshotInCollection(dataSourceName, snapshotName),
				),
			},
		},
	})
}

func testDataSourceConfig(snapshotName, diskName string) string {
	return fmt.Sprintf(`
data "oxide_project" "support" {
  name = "tf-acc-test"
}

resource "oxide_disk" "test" {
  project_id  = data.oxide_project.support.id
  description = "a test disk for the snapshots data source"
  name        = %q
  size        = 1073741824
  block_size  = 512
}

resource "oxide_snapshot" "test" {
  project_id  = data.oxide_project.support.id
  description = "a test snapshot for the collection data source"
  name        = %q
  disk_id     = oxide_disk.test.id
}

data "oxide_snapshots" "test" {
  project    = data.oxide_project.support.name
  depends_on = [oxide_snapshot.test]
  timeouts = {
    read = "1m"
  }
}
`, diskName, snapshotName)
}

func checkSnapshotInCollection(dataSourceName, snapshotName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("data source %s not found in state", dataSourceName)
		}

		attributes := resourceState.Primary.Attributes
		count, err := strconv.Atoi(attributes["snapshots.#"])
		if err != nil {
			return fmt.Errorf("invalid snapshot count: %w", err)
		}

		for i := range count {
			prefix := fmt.Sprintf("snapshots.%d.", i)
			if attributes[prefix+"name"] != snapshotName {
				continue
			}

			for _, attribute := range []string{
				"disk_id", "id", "project_id", "size", "state", "time_created", "time_modified",
			} {
				if attributes[prefix+attribute] == "" {
					return fmt.Errorf("snapshot %q has empty %s", snapshotName, attribute)
				}
			}
			if got := attributes[prefix+"description"]; got != "a test snapshot for the collection data source" {
				return fmt.Errorf("snapshot %q has description %q", snapshotName, got)
			}

			return nil
		}

		return fmt.Errorf("snapshot %q not found in collection", snapshotName)
	}
}
