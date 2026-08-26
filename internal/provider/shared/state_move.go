// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package shared

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// MoveState returns a state mover for resource types whose source and target
// state models are identical.
func MoveState[T any](
	sourceTypeName string,
	sourceSchemaVersion int64,
	sourceSchema schema.Schema,
) resource.StateMover {
	return resource.StateMover{
		SourceSchema: &sourceSchema,
		StateMover: func(
			ctx context.Context,
			req resource.MoveStateRequest,
			resp *resource.MoveStateResponse,
		) {
			if req.SourceTypeName != sourceTypeName ||
				req.SourceSchemaVersion != sourceSchemaVersion ||
				!strings.HasSuffix(req.SourceProviderAddress, "/oxidecomputer/oxide") {
				return
			}
			if req.SourceState == nil {
				resp.Diagnostics.AddError(
					"Unable to Move Resource State",
					"The source resource state could not be decoded.",
				)
				return
			}

			var sourceState T
			resp.Diagnostics.Append(req.SourceState.Get(ctx, &sourceState)...)
			if resp.Diagnostics.HasError() {
				return
			}

			resp.Diagnostics.Append(resp.TargetState.Set(ctx, &sourceState)...)
		},
	}
}
