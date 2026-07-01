package validator

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// Compile-time interface assertion.
var _ validator.String = isUUID{}

// isUUID validates that a configured string is a valid UUID.
type isUUID struct{}

// Description returns a plain text description of the validator's behavior.
func (v isUUID) Description(_ context.Context) string {
	return "Value must be a valid UUID"
}

// MarkdownDescription returns a markdown description of the validator's
// behavior.
func (v isUUID) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString validates that a configured string is a valid UUID. Null
// and unknown values are skipped so that this validator can be composed with
// others.
func (v isUUID) ValidateString(
	_ context.Context,
	req validator.StringRequest,
	resp *validator.StringResponse,
) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if _, err := uuid.Parse(value); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid UUID",
			fmt.Sprintf(
				"Attribute %s value must be a valid UUID, got: %s",
				req.Path,
				value,
			),
		)
	}
}

// IsUUID returns a string validator which ensures that a configured value is a
// valid UUID.
func IsUUID() validator.String {
	return isUUID{}
}
