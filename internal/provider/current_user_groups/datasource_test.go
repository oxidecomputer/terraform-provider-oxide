// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package currentusergroups_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestDataSourceCurrentUserGroups(t *testing.T) {
	var firstPageRequests atomic.Int32
	var secondPageRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/me/groups" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("limit"); got != "100" {
			t.Errorf("unexpected limit: %q", got)
		}
		if got := r.URL.Query().Get("sort_by"); got != "id_ascending" {
			t.Errorf("unexpected sort_by: %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		switch pageToken := r.URL.Query().Get("page_token"); pageToken {
		case "":
			firstPageRequests.Add(1)
			fmt.Fprint(w, `{
  "items": [{
    "display_name": "Operators",
    "id": "00000000-0000-0000-0000-000000000001",
    "silo_id": "10000000-0000-0000-0000-000000000001",
    "time_created": "2026-01-02T03:04:05Z",
    "time_modified": "2026-02-03T04:05:06Z"
  }],
  "next_page": "second-page"
}`)
		case "second-page":
			secondPageRequests.Add(1)
			fmt.Fprint(w, `{
  "items": [{
    "display_name": "Developers",
    "id": "00000000-0000-0000-0000-000000000002",
    "silo_id": "10000000-0000-0000-0000-000000000002",
    "time_created": "2026-03-04T05:06:07Z",
    "time_modified": "2026-04-05T06:07:08Z"
  }]
}`)
		default:
			t.Errorf("unexpected page token: %q", pageToken)
		}
	}))
	t.Cleanup(server.Close)
	t.Setenv("OXIDE_HOST", server.URL)
	t.Setenv("OXIDE_TOKEN", "test-token")

	const dataSourceName = "data.oxide_current_user_groups.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `data "oxide_current_user_groups" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttr(dataSourceName, "groups.#", "2"),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"groups.0.display_name",
						"Operators",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"groups.0.id",
						"00000000-0000-0000-0000-000000000001",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"groups.0.silo_id",
						"10000000-0000-0000-0000-000000000001",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"groups.0.time_created",
						"2026-01-02 03:04:05 +0000 UTC",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"groups.0.time_modified",
						"2026-02-03 04:05:06 +0000 UTC",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"groups.1.display_name",
						"Developers",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"groups.1.id",
						"00000000-0000-0000-0000-000000000002",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"groups.1.silo_id",
						"10000000-0000-0000-0000-000000000002",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"groups.1.time_created",
						"2026-03-04 05:06:07 +0000 UTC",
					),
					resource.TestCheckResourceAttr(
						dataSourceName,
						"groups.1.time_modified",
						"2026-04-05 06:07:08 +0000 UTC",
					),
				),
			},
		},
	})

	if firstPageRequests.Load() == 0 || secondPageRequests.Load() == 0 {
		t.Fatalf(
			"pagination not exercised: first page requests %d, second page requests %d",
			firstPageRequests.Load(),
			secondPageRequests.Load(),
		)
	}
}

func TestAccDataSourceCurrentUserGroups_full(t *testing.T) {
	const dataSourceName = "data.oxide_current_user_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
data "oxide_current_user_groups" "test" {
  timeouts = {
    read = "1m"
  }
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "groups.#"),
					resource.TestCheckResourceAttr(dataSourceName, "timeouts.read", "1m"),
				),
			},
		},
	})
}
