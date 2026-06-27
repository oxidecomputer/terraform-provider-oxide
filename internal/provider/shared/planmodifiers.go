// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package shared

import (
	"context"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// RequiresReplaceUnlessEmptyStringOrNull returns a resource.RequiresReplaceIfFunc that
// returns true unless the change is from an empty string or null.
//
// This is particularly helpful for creating new nested objects that cannot be modified
// themselves, but it is possible to add or remove them.
func RequiresReplaceUnlessEmptyStringOrNull() stringplanmodifier.RequiresReplaceIfFunc {
	return func(ctx context.Context, req planmodifier.StringRequest,
		resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
		// If the configuration is unknown, this cannot be sure what to do yet.
		if req.ConfigValue.IsUnknown() {
			resp.RequiresReplace = false
			return
		}

		if req.StateValue.IsNull() || req.StateValue.ValueString() == "" {
			resp.RequiresReplace = false
			return
		}

		resp.RequiresReplace = true
	}
}

// RequiresReplaceUnlessNonUUIDToUUID returns a
// [stringplanmodifier.RequiresReplaceIfFunc] that requires resource replacement
// when an attribute's value changes in all cases except when going from
// a non-UUID to a UUID. This allows attributes that previously took an
// [oxide.NameOrId] to safely migrate to a UUID without replacement.
func RequiresReplaceUnlessNonUUIDToUUID() stringplanmodifier.RequiresReplaceIfFunc {
	return func(ctx context.Context, req planmodifier.StringRequest,
		resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {

		// If the new value isn't known yet, fall back to replacement.
		if req.ConfigValue.IsUnknown() {
			resp.RequiresReplace = true
			return
		}

		// Without a prior state value this cannot be a name to UUID correction.
		if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
			resp.RequiresReplace = true
			return
		}

		_, stateErr := uuid.Parse(req.StateValue.ValueString())
		_, configErr := uuid.Parse(req.ConfigValue.ValueString())

		// This is a valid name to UUID change. Do not replace the resource.
		if stateErr != nil && configErr == nil {
			resp.RequiresReplace = false
			return
		}

		resp.RequiresReplace = true
	}
}
