// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolsilolink

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseID(t *testing.T) {
	t.Parallel()

	pool, silo, err := parseID("pool-id/silo-id")
	require.NoError(t, err)
	require.Equal(t, "pool-id", pool)
	require.Equal(t, "silo-id", silo)

	for _, id := range []string{"", "pool-id", "/silo-id", "pool-id/", "pool-id/silo-id/extra"} {
		t.Run(id, func(t *testing.T) {
			t.Parallel()
			_, _, err := parseID(id)
			require.Error(t, err)
		})
	}
}
