// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolmember

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseImportID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		id         string
		wantPool   string
		wantSubnet string
		wantError  string
	}{
		{
			name:       "IPv4",
			id:         "my-pool/192.0.2.0/24",
			wantPool:   "my-pool",
			wantSubnet: "192.0.2.0/24",
		},
		{
			name:       "canonicalizes IPv6",
			id:         "my-pool/2001:0DB8:0000:0000::/64",
			wantPool:   "my-pool",
			wantSubnet: "2001:db8::/64",
		},
		{
			name:      "missing subnet",
			id:        "my-pool",
			wantError: "expected import ID format",
		},
		{
			name:      "invalid subnet",
			id:        "my-pool/not-a-subnet",
			wantError: "invalid subnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pool, subnet, err := parseImportID(tt.id)
			if tt.wantError != "" {
				require.ErrorContains(t, err, tt.wantError)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantPool, pool)
			require.Equal(t, tt.wantSubnet, subnet)
		})
	}
}
