package oxide

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// RequiresReplaceUnlessEmptyStringToNull returns a resource.RequiresReplaceIfFunc that
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
