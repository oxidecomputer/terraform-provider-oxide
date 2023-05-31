// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newResourceName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "results are always different and contain prefix",
			want: "acc-terraform-",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name1 := newResourceName()
			name2 := newResourceName()
			name3 := newResourceName()
			name4 := newResourceName()
			assert.NotEqual(t, name1, name2, name3, name4)
			assert.Contains(t, name1, tt.want)
			assert.Contains(t, name2, tt.want)
			assert.Contains(t, name3, tt.want)
			assert.Contains(t, name4, tt.want)
		})
	}
}

func Test_newBlockName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "results are always different and contain prefix",
			want: "acc-vpc-",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name1 := newBlockName("vpc")
			name2 := newBlockName("vpc")
			name3 := newBlockName("vpc")
			name4 := newBlockName("vpc")
			assert.NotEqual(t, name1, name2, name3, name4)
			assert.Contains(t, name1, tt.want)
			assert.Contains(t, name2, tt.want)
			assert.Contains(t, name3, tt.want)
			assert.Contains(t, name4, tt.want)
		})
	}
}
