// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package systemsubnetpoolmember

import (
	"errors"
	"net/http"
	"testing"

	"github.com/oxidecomputer/oxide.go/oxide"
)

func TestIsSubnetPoolMemberNotFound(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		err  error
		want bool
	}{
		"matching invalid request": {
			err: newHTTPError(
				http.StatusBadRequest,
				"InvalidRequest",
				"A provided subnet pool member with subnet 1.1.1.1/32 does not exist",
			),
			want: true,
		},
		"different invalid request": {
			err: newHTTPError(
				http.StatusBadRequest,
				"InvalidRequest",
				"Cannot delete external subnet pool member while it contains external subnets.",
			),
			want: false,
		},
		"different not found message": {
			err: newHTTPError(
				http.StatusBadRequest,
				"InvalidRequest",
				"The subnet pool member does not exist in this pool",
			),
			want: true,
		},
		"http not found": {
			err:  newHTTPError(http.StatusNotFound, "ObjectNotFound", "not found"),
			want: true,
		},
		"other error": {
			err:  errors.New("connection failed"),
			want: false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if got := isSubnetPoolMemberNotFound(test.err); got != test.want {
				t.Fatalf("isSubnetPoolMemberNotFound() = %t, want %t", got, test.want)
			}
		})
	}
}

func newHTTPError(status int, code, message string) error {
	return &oxide.HTTPError{
		HTTPResponse: &http.Response{StatusCode: status},
		ErrorResponse: &oxide.ErrorResponse{
			ErrorCode: code,
			Message:   message,
		},
	}
}
