// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package shared

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func Test_RequiresReplaceUnlessNonUUIDToUUID(t *testing.T) {
	const (
		uuidA = "8f1ce7f8-0c5b-4e4e-9b1a-6f6d8c1f2a3b"
		uuidB = "1b2c3d4e-5f60-7182-93a4-b5c6d7e8f900"
	)

	tests := []struct {
		name        string
		stateValue  types.String
		configValue types.String
		wantReplace bool
	}{
		{
			name:        "name to UUID is non-destructive",
			stateValue:  types.StringValue("my-silo"),
			configValue: types.StringValue(uuidA),
			wantReplace: false,
		},
		{
			name:        "distinct UUIDs still replace",
			stateValue:  types.StringValue(uuidA),
			configValue: types.StringValue(uuidB),
			wantReplace: true,
		},
		{
			name:        "UUID to name still replace",
			stateValue:  types.StringValue(uuidA),
			configValue: types.StringValue("my-silo"),
			wantReplace: true,
		},
		{
			name:        "name to name still replace",
			stateValue:  types.StringValue("my-silo"),
			configValue: types.StringValue("other-silo"),
			wantReplace: true,
		},
		{
			name:        "null state replaces",
			stateValue:  types.StringNull(),
			configValue: types.StringValue(uuidA),
			wantReplace: true,
		},
		{
			name:        "unknown config replaces",
			stateValue:  types.StringValue("my-silo"),
			configValue: types.StringUnknown(),
			wantReplace: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				StateValue:  tt.stateValue,
				ConfigValue: tt.configValue,
				PlanValue:   tt.configValue,
			}
			resp := &stringplanmodifier.RequiresReplaceIfFuncResponse{}

			RequiresReplaceUnlessNonUUIDToUUID()(context.Background(), req, resp)

			assert.Equal(t, tt.wantReplace, resp.RequiresReplace)
		})
	}
}
