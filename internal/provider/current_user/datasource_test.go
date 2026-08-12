// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package currentuser_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestAccDataSourceCurrentUser_full(t *testing.T) {
	const dataSourceName = "data.oxide_current_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "oxide_current_user" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "display_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "silo_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "silo_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "time_created"),
					resource.TestCheckResourceAttrSet(dataSourceName, "time_modified"),
				),
			},
		},
	})
}
