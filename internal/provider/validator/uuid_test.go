// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package validator

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_IsUUID(t *testing.T) {
	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{
			name:      "valid UUID",
			value:     types.StringValue("8f1ce7f8-0c5b-4e4e-9b1a-6f6d8c1f2a3b"),
			wantError: false,
		},
		{
			name:      "invalid UUID (name)",
			value:     types.StringValue("my-silo"),
			wantError: true,
		},
		{
			name:      "empty string",
			value:     types.StringValue(""),
			wantError: true,
		},
		{
			name:      "null is skipped",
			value:     types.StringNull(),
			wantError: false,
		},
		{
			name:      "unknown is skipped",
			value:     types.StringUnknown(),
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("test"),
				ConfigValue: tt.value,
			}
			resp := &validator.StringResponse{}

			IsUUID().ValidateString(context.Background(), req, resp)

			assert.Equal(t, tt.wantError, resp.Diagnostics.HasError())
		})
	}
}
