// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemippoolsilolink_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	frameworkresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/oxidecomputer/oxide.go/oxide"
	"github.com/stretchr/testify/require"

	legacyippoolsilolink "github.com/oxidecomputer/terraform-provider-oxide/internal/provider/ip_pool_silo_link"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
	systemippoolsilolink "github.com/oxidecomputer/terraform-provider-oxide/internal/provider/system_ip_pool_silo_link"
)

type systemIPPoolSiloLinkResourceModel struct {
	ID        types.String   `tfsdk:"id"`
	IsDefault types.Bool     `tfsdk:"is_default"`
	IPPoolID  types.String   `tfsdk:"ip_pool_id"`
	SiloID    types.String   `tfsdk:"silo_id"`
	Timeouts  timeouts.Value `tfsdk:"timeouts"`
}

func TestMoveState(t *testing.T) {
	t.Parallel()
	const poolID = "4f0e69ad-66b6-41c0-b727-7b0285b0c384"
	const siloID = "9e199e45-01a6-43d3-8bc3-5b27726e67a6"

	ctx := context.Background()
	res := systemippoolsilolink.NewResource()
	movers := res.(frameworkresource.ResourceWithMoveState).MoveState(ctx)
	require.Len(t, movers, 1)

	sourceState := newMoveState(ctx, *movers[0].SourceSchema)
	timeoutValue := timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType,
		"read":   types.StringType,
		"update": types.StringType,
		"delete": types.StringType,
	})}
	diags := sourceState.Set(ctx, &legacyippoolsilolink.ResourceModel{
		ID:        types.StringValue("legacy-random-id"),
		IPPoolID:  types.StringValue(poolID),
		SiloID:    types.StringValue(siloID),
		IsDefault: types.BoolValue(true),
		Timeouts:  timeoutValue,
	})
	require.False(t, diags.HasError(), diags)

	var targetSchemaResponse frameworkresource.SchemaResponse
	res.Schema(ctx, frameworkresource.SchemaRequest{}, &targetSchemaResponse)
	response := frameworkresource.MoveStateResponse{
		TargetState: newMoveState(ctx, targetSchemaResponse.Schema),
	}
	movers[0].StateMover(ctx, frameworkresource.MoveStateRequest{
		SourceProviderAddress: "registry.terraform.io/oxidecomputer/oxide",
		SourceSchemaVersion:   0,
		SourceState:           &sourceState,
		SourceTypeName:        "oxide_ip_pool_silo_link",
	}, &response)
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)

	var got systemIPPoolSiloLinkResourceModel
	response.Diagnostics.Append(response.TargetState.Get(ctx, &got)...)
	require.False(t, response.Diagnostics.HasError(), response.Diagnostics)
	require.Equal(t, poolID+"/"+siloID, got.ID.ValueString())
	require.Equal(t, poolID, got.IPPoolID.ValueString())
	require.Equal(t, siloID, got.SiloID.ValueString())
	require.True(t, got.IsDefault.ValueBool())
	require.True(t, timeoutValue.Equal(got.Timeouts))
}

func TestImportStateRejectsNonUUIDs(t *testing.T) {
	t.Parallel()

	testCases := map[string]string{
		"pool name": "my-pool/9e199e45-01a6-43d3-8bc3-5b27726e67a6",
		"silo name": "4f0e69ad-66b6-41c0-b727-7b0285b0c384/my-silo",
	}
	res := systemippoolsilolink.NewResource().(frameworkresource.ResourceWithImportState)
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

func TestMoveStateRejectsPreV021NameState(t *testing.T) {
	t.Parallel()
	const poolID = "4f0e69ad-66b6-41c0-b727-7b0285b0c384"
	const siloID = "9e199e45-01a6-43d3-8bc3-5b27726e67a6"

	testCases := map[string]legacyippoolsilolink.ResourceModel{
		"pool name": {
			IPPoolID: types.StringValue("my-pool"),
			SiloID:   types.StringValue(siloID),
		},
		"silo name": {
			IPPoolID: types.StringValue(poolID),
			SiloID:   types.StringValue("my-silo"),
		},
	}
	for name, source := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mover := systemippoolsilolink.NewResource().(frameworkresource.ResourceWithMoveState).MoveState(ctx)[0]
			sourceState := newMoveState(ctx, *mover.SourceSchema)
			source.ID = types.StringValue("legacy-random-id")
			source.Timeouts = timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
				"create": types.StringType,
				"read":   types.StringType,
				"update": types.StringType,
				"delete": types.StringType,
			})}
			diags := sourceState.Set(ctx, &source)
			require.False(t, diags.HasError(), diags)

			var response frameworkresource.MoveStateResponse
			mover.StateMover(ctx, frameworkresource.MoveStateRequest{
				SourceProviderAddress: "registry.terraform.io/oxidecomputer/oxide",
				SourceSchemaVersion:   0,
				SourceState:           &sourceState,
				SourceTypeName:        "oxide_ip_pool_silo_link",
			}, &response)
			require.True(t, response.Diagnostics.HasError(), response.Diagnostics)
			require.Contains(t, response.Diagnostics.Errors()[0].Detail(), "v0.21.0")
		})
	}
}

func newMoveState(ctx context.Context, stateSchema schema.Schema) tfsdk.State {
	return tfsdk.State{
		Schema: stateSchema,
		Raw: tftypes.NewValue(
			stateSchema.Type().TerraformType(ctx),
			nil,
		),
	}
}

func TestAccResourceSystemIPPoolSiloLink_full(t *testing.T) {
	firstPoolName := sharedtest.NewResourceName()
	secondPoolName := sharedtest.NewResourceName()
	linkResourceName := "oxide_system_ip_pool_silo_link.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testResourceConfig(firstPoolName, secondPoolName, "first", "1m"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(linkResourceName, "id"),
					resource.TestCheckResourceAttr(linkResourceName, "is_default", "false"),
					resource.TestCheckResourceAttrPair(
						linkResourceName,
						"ip_pool_id",
						"oxide_ip_pool.first",
						"id",
					),
					resource.TestCheckResourceAttrPair(
						linkResourceName,
						"silo_id",
						"data.oxide_silo.test",
						"id",
					),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.read", "1m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.create", "3m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.update", "4m"),
					resource.TestCheckResourceAttr(linkResourceName, "timeouts.delete", "2m"),
					checkLinkIDComposite(linkResourceName),
				),
			},
			{
				Config: testResourceConfig(firstPoolName, secondPoolName, "first", "2m"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							linkResourceName,
							plancheck.ResourceActionUpdate,
						),
					},
				},
				Check: resource.TestCheckResourceAttr(
					linkResourceName,
					"timeouts.read",
					"2m",
				),
			},
			{
				Config: testResourceConfig(firstPoolName, secondPoolName, "second", "2m"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(
							linkResourceName,
							plancheck.ResourceActionReplace,
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						linkResourceName,
						"ip_pool_id",
						"oxide_ip_pool.second",
						"id",
					),
					resource.TestCheckResourceAttrPair(
						linkResourceName,
						"silo_id",
						"data.oxide_silo.test",
						"id",
					),
					checkLinkIDComposite(linkResourceName),
				),
			},
			{
				ResourceName:            linkResourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
			{
				Config: testResourceConfig(firstPoolName, secondPoolName, "second", "2m"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccResourceDisappears(linkResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testResourceConfig(
	firstPoolName, secondPoolName, selectedPool string,
	readTimeout string,
) string {
	return fmt.Sprintf(`
data "oxide_silo" "test" {
	name = "test-suite-silo"
}

resource "oxide_ip_pool" "first" {
	name        = %[1]q
	description = "first system IP pool silo link test pool"
}

resource "oxide_ip_pool" "second" {
	name        = %[2]q
	description = "second system IP pool silo link test pool"
}

resource "oxide_system_ip_pool_silo_link" "test" {
	ip_pool_id = oxide_ip_pool.%[3]s.id
	silo_id    = data.oxide_silo.test.id
	is_default = false
	timeouts = {
		read   = %[4]q
		create = "3m"
		delete = "2m"
		update = "4m"
	}
}
`, firstPoolName, secondPoolName, selectedPool, readTimeout)
}

func checkLinkIDComposite(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		parts := strings.Split(rs.Primary.ID, "/")
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("expected id in IP_POOL_ID/SILO_ID format, got %q", rs.Primary.ID)
		}

		return nil
	}
}

func testAccResourceDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		client, err := sharedtest.NewTestClient()
		if err != nil {
			return err
		}

		return client.SystemIpPoolSiloUnlink(
			context.Background(),
			oxide.SystemIpPoolSiloUnlinkParams{
				Pool: oxide.NameOrId(rs.Primary.Attributes["ip_pool_id"]),
				Silo: oxide.NameOrId(rs.Primary.Attributes["silo_id"]),
			},
		)
	}
}
