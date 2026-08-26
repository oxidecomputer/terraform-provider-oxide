// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package shared

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type stateMoveModel struct {
	Value types.String `tfsdk:"value"`
}

func TestMoveState(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testSchema := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"value": schema.StringAttribute{Optional: true},
		},
	}
	sourceState := newState(ctx, t, testSchema)
	diags := sourceState.Set(ctx, &stateMoveModel{
		Value: types.StringValue("moved"),
	})
	if diags.HasError() {
		t.Fatalf("setting source state: %v", diags)
	}

	mover := MoveState[stateMoveModel]("oxide_source", 0, testSchema)
	resp := resource.MoveStateResponse{
		TargetState: newState(ctx, t, testSchema),
	}
	mover.StateMover(ctx, resource.MoveStateRequest{
		SourceProviderAddress: "registry.terraform.io/oxidecomputer/oxide",
		SourceSchemaVersion:   0,
		SourceState:           &sourceState,
		SourceTypeName:        "oxide_source",
	}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("moving state: %v", resp.Diagnostics)
	}

	var got stateMoveModel
	resp.Diagnostics.Append(resp.TargetState.Get(ctx, &got)...)
	if resp.Diagnostics.HasError() {
		t.Fatalf("reading target state: %v", resp.Diagnostics)
	}
	if got.Value.ValueString() != "moved" {
		t.Fatalf("value = %q, want %q", got.Value.ValueString(), "moved")
	}
}

func newState(
	ctx context.Context,
	t *testing.T,
	testSchema schema.Schema,
) tfsdk.State {
	t.Helper()
	return tfsdk.State{
		Schema: testSchema,
		Raw: tftypes.NewValue(
			testSchema.Type().TerraformType(ctx),
			nil,
		),
	}
}
