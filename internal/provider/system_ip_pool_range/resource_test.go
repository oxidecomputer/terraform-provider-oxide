// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippoolrange_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
	"github.com/stretchr/testify/require"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
	systemippoolrange "github.com/oxidecomputer/terraform-provider-oxide/internal/provider/system_ip_pool_range"
)

func TestImportStateRejectsNonUUIDs(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"pool name":  "my-pool/3e2c6e84-bed8-4c94-afc3-1032082d6a90",
		"range name": "4f0e69ad-66b6-41c0-b727-7b0285b0c384/my-range",
	}
	res := systemippoolrange.NewResource().(frameworkresource.ResourceWithImportState)
	for name, importID := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var response frameworkresource.ImportStateResponse
			res.ImportState(
				context.Background(),
				frameworkresource.ImportStateRequest{ID: importID},
				&response,
			)
			require.True(t, response.Diagnostics.HasError(), response.Diagnostics)
		})
	}
}

func TestAccResourceSystemIPPoolRange_full(t *testing.T) {
	poolName := sharedtest.NewResourceName()
	resourceName := "oxide_system_ip_pool_range.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		CheckDestroy:             testAccResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig(
					poolName,
					"172.20.30.10",
					"172.20.30.19",
					"1m",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"ip_pool_id",
						"oxide_ip_pool.test",
						"id",
					),
					resource.TestCheckResourceAttr(resourceName, "first_address", "172.20.30.10"),
					resource.TestCheckResourceAttr(resourceName, "last_address", "172.20.30.19"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.read", "1m"),
				),
			},
			{
				Config: testResourceConfig(
					poolName,
					"172.20.30.10",
					"172.20.30.19",
					"2m",
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							resourceName,
							plancheck.ResourceActionUpdate,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "first_address", "172.20.30.10"),
					resource.TestCheckResourceAttr(resourceName, "last_address", "172.20.30.19"),
					resource.TestCheckResourceAttr(resourceName, "timeouts.read", "2m"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					rangeResource, ok := state.RootModule().Resources[resourceName]
					if !ok {
						return "", fmt.Errorf("resource not found: %s", resourceName)
					}
					return fmt.Sprintf(
						"%s/%s",
						rangeResource.Primary.Attributes["ip_pool_id"],
						rangeResource.Primary.ID,
					), nil
				},
			},
			{
				Config: testResourceWithUpdatedPoolConfig(
					poolName,
					"172.20.30.10",
					"172.20.30.19",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "first_address", "172.20.30.10"),
					resource.TestCheckResourceAttr(resourceName, "last_address", "172.20.30.19"),
				),
			},
			{
				Config: testResourceConfig(
					poolName,
					"172.20.30.20",
					"172.20.30.29",
					"2m",
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "first_address", "172.20.30.20"),
					resource.TestCheckResourceAttr(resourceName, "last_address", "172.20.30.29"),
				),
			},
		},
	})
}

func testResourceWithUpdatedPoolConfig(poolName, first, last string) string {
	return fmt.Sprintf(`
resource "oxide_ip_pool" "test" {
	name        = %q
	description = "updated system IP pool range acceptance test"
}

resource "oxide_system_ip_pool_range" "test" {
	ip_pool_id    = oxide_ip_pool.test.id
	first_address = %q
	last_address  = %q
}
`, poolName, first, last)
}

func testResourceConfig(poolName, first, last, readTimeout string) string {
	return fmt.Sprintf(`
resource "oxide_ip_pool" "test" {
	name        = %q
	description = "system IP pool range acceptance test"
}

resource "oxide_system_ip_pool_range" "test" {
	ip_pool_id    = oxide_ip_pool.test.id
	first_address = %q
	last_address  = %q
	timeouts = {
		read = %q
	}
}
`, poolName, first, last, readTimeout)
}

func testAccResourceDestroy(state *terraform.State) error {
	client, err := sharedtest.NewTestClient()
	if err != nil {
		return err
	}

	for _, resourceState := range state.RootModule().Resources {
		if resourceState.Type != "oxide_system_ip_pool_range" {
			continue
		}

		ranges, err := client.SystemIpPoolRangeListAllPages(
			context.Background(),
			oxide.SystemIpPoolRangeListParams{
				Pool: oxide.NameOrId(resourceState.Primary.Attributes["ip_pool_id"]),
			},
		)
		if err != nil && errors.Is(err, oxide.ErrHTTP404) {
			continue
		}
		if err != nil {
			return err
		}

		for _, ipRange := range ranges {
			if ipRange.Id == resourceState.Primary.ID {
				return fmt.Errorf("system_ip_pool_range (%s) still exists", ipRange.Id)
			}
		}
	}

	return nil
}
