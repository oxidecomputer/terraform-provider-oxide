// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package vpcs_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

func TestDataSourceVPCs(t *testing.T) {
	var firstPageRequests atomic.Int32
	var secondPageRequests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v1/vpcs" {
			t.Errorf("request path = %q, want /v1/vpcs", req.URL.Path)
		}
		if got := req.URL.Query().Get("project"); got != "project-name" {
			t.Errorf("project query = %q, want project-name", got)
		}

		w.Header().Set("Content-Type", "application/json")
		if req.URL.Query().Get("page_token") == "" {
			firstPageRequests.Add(1)
			_, _ = w.Write([]byte(`{
  "items": [{
    "description": "first description",
    "dns_name": "first-dns",
    "id": "00000000-0000-0000-0000-000000000001",
    "ipv6_prefix": "fd00:1::/48",
    "name": "first",
    "project_id": "00000000-0000-0000-0000-000000000010",
    "system_router_id": "00000000-0000-0000-0000-000000000011",
    "time_created": "2026-01-01T01:02:03Z",
    "time_modified": "2026-01-02T01:02:03Z"
  }],
  "next_page": "second-page"
}`))
			return
		}
		if got := req.URL.Query().Get("page_token"); got != "second-page" {
			t.Errorf("page token = %q, want second-page", got)
		}
		secondPageRequests.Add(1)

		_, _ = w.Write([]byte(`{
  "items": [{
    "description": "second description",
    "dns_name": "second-dns",
    "id": "00000000-0000-0000-0000-000000000002",
    "ipv6_prefix": "fd00:2::/48",
    "name": "second",
    "project_id": "00000000-0000-0000-0000-000000000010",
    "system_router_id": "00000000-0000-0000-0000-000000000012",
    "time_created": "2026-02-01T01:02:03Z",
    "time_modified": "2026-02-02T01:02:03Z"
  }]
}`))
	}))
	t.Cleanup(server.Close)

	//lintignore:AT004 // Provider must connect to test server.
	config := fmt.Sprintf(`
provider "oxide" {
  host  = %q
  token = "fake"
}

data "oxide_vpcs" "test" {
  project = "project-name"
}
`, server.URL)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.oxide_vpcs.test", "vpcs.#", "2"),
					resource.TestCheckResourceAttr("data.oxide_vpcs.test", "vpcs.0.name", "first"),
					resource.TestCheckResourceAttr(
						"data.oxide_vpcs.test",
						"vpcs.0.description",
						"first description",
					),
					resource.TestCheckResourceAttr(
						"data.oxide_vpcs.test",
						"vpcs.0.dns_name",
						"first-dns",
					),
					resource.TestCheckResourceAttr(
						"data.oxide_vpcs.test",
						"vpcs.0.ipv6_prefix",
						"fd00:1::/48",
					),
					resource.TestCheckResourceAttr("data.oxide_vpcs.test", "vpcs.1.name", "second"),
					resource.TestCheckResourceAttr(
						"data.oxide_vpcs.test",
						"vpcs.1.system_router_id",
						"00000000-0000-0000-0000-000000000012",
					),
				),
			},
		},
	})

	if firstPageRequests.Load() == 0 || firstPageRequests.Load() != secondPageRequests.Load() {
		t.Fatalf(
			"VPC list page requests = (%d, %d), want equal non-zero counts",
			firstPageRequests.Load(),
			secondPageRequests.Load(),
		)
	}
}
