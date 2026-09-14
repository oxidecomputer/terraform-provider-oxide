// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package siloutilization_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestDataSourceSiloUtilization(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/utilization" {
			http.NotFound(w, r)
			return
		}

		fmt.Fprint(w, `{
  "capacity": {"cpus": 2, "memory": 3, "storage": 5},
  "provisioned": {"cpus": 7, "memory": 11, "storage": 13}
}`)
	}))
	t.Cleanup(server.Close)
	t.Setenv("OXIDE_HOST", server.URL)
	t.Setenv("OXIDE_TOKEN", "test-token")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "oxide_silo_utilization" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.oxide_silo_utilization.test",
						"capacity.cpus",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.oxide_silo_utilization.test",
						"capacity.memory",
						"3",
					),
					resource.TestCheckResourceAttr(
						"data.oxide_silo_utilization.test",
						"capacity.storage",
						"5",
					),
					resource.TestCheckResourceAttr(
						"data.oxide_silo_utilization.test",
						"provisioned.cpus",
						"7",
					),
					resource.TestCheckResourceAttr(
						"data.oxide_silo_utilization.test",
						"provisioned.memory",
						"11",
					),
					resource.TestCheckResourceAttr(
						"data.oxide_silo_utilization.test",
						"provisioned.storage",
						"13",
					),
				),
			},
		},
	})
}

func TestAccDataSourceSiloUtilization_full(t *testing.T) {
	const dataSourceName = "data.oxide_silo_utilization.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "oxide_silo_utilization" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "capacity.cpus"),
					resource.TestCheckResourceAttrSet(dataSourceName, "capacity.memory"),
					resource.TestCheckResourceAttrSet(dataSourceName, "capacity.storage"),
					resource.TestCheckResourceAttrSet(dataSourceName, "provisioned.cpus"),
					resource.TestCheckResourceAttrSet(dataSourceName, "provisioned.memory"),
					resource.TestCheckResourceAttrSet(dataSourceName, "provisioned.storage"),
				),
			},
		},
	})
}
