// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package antiaffinitygroupmembers_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccCloudDataSourceAntiAffinityGroupMembers_full(t *testing.T) {
	const dataSourceName = "data.oxide_anti_affinity_group_members.test"
	groupName := sharedtest.NewResourceName()
	instanceBaseName := sharedtest.NewResourceName()
	// The instance declared first gets the name that sorts last so the test
	// verifies members are ordered by name rather than by creation.
	firstInstanceName := instanceBaseName + "-b"
	secondInstanceName := instanceBaseName + "-a"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testDataSourceConfigEmpty(groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "project_name", "tf-acc-test"),
					resource.TestCheckResourceAttr(dataSourceName, "name", groupName),
					resource.TestCheckResourceAttr(dataSourceName, "timeouts.read", "1m"),
					resource.TestCheckResourceAttr(dataSourceName, "members.#", "0"),
				),
			},
			{
				Config: testDataSourceConfigMembers(
					groupName,
					firstInstanceName,
					secondInstanceName,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "members.#", "2"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "members.0.id",
						"oxide_instance.second", "id",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"members.0.name",
						secondInstanceName,
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"members.0.run_state",
						"stopped",
					),
					resource.TestCheckResourceAttr(dataSourceName, "members.0.type", "instance"),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "members.1.id",
						"oxide_instance.first", "id",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"members.1.name",
						firstInstanceName,
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"members.1.run_state",
						"stopped",
					),
					resource.TestCheckResourceAttr(dataSourceName, "members.1.type", "instance"),
				),
			},
		},
	})
}

func testDataSourceConfigEmpty(groupName string) string {
	return fmt.Sprintf(`
data "oxide_project" "test" {
  name = "tf-acc-test"
}

resource "oxide_anti_affinity_group" "test" {
  project_id  = data.oxide_project.test.id
  name        = %q
  description = "a group"
  policy      = "allow"
}

data "oxide_anti_affinity_group_members" "test" {
  project_name = "tf-acc-test"
  name         = oxide_anti_affinity_group.test.name
  timeouts = {
    read = "1m"
  }
}
`, groupName)
}

func testDataSourceConfigMembers(groupName, firstInstanceName, secondInstanceName string) string {
	return fmt.Sprintf(`
data "oxide_project" "test" {
  name = "tf-acc-test"
}

resource "oxide_anti_affinity_group" "test" {
  project_id  = data.oxide_project.test.id
  name        = %q
  description = "a group"
  policy      = "allow"
}

resource "oxide_instance" "first" {
  anti_affinity_groups = [oxide_anti_affinity_group.test.id]
  project_id           = data.oxide_project.test.id
  description          = "a test instance"
  name                 = %q
  hostname             = "terraform-acc-myhost"
  memory               = 1073741824
  ncpus                = 1
  start_on_create      = false
}

resource "oxide_instance" "second" {
  anti_affinity_groups = [oxide_anti_affinity_group.test.id]
  project_id           = data.oxide_project.test.id
  description          = "a test instance"
  name                 = %q
  hostname             = "terraform-acc-myhost"
  memory               = 1073741824
  ncpus                = 1
  start_on_create      = false
}

data "oxide_anti_affinity_group_members" "test" {
  depends_on = [oxide_instance.first, oxide_instance.second]

  project_name = "tf-acc-test"
  name         = oxide_anti_affinity_group.test.name
  timeouts = {
    read = "1m"
  }
}
`, groupName, firstInstanceName, secondInstanceName)
}
