// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"net"
	"strings"
	"time"
)

func is404(err error) bool {
	if strings.Contains(err.Error(), "HTTP 404") ||
		strings.Contains(err.Error(), "404 Not Found") {
		return true
	}
	return false
}

// Original function from https://pkg.go.dev/github.com/asaskevich/govalidator#IsIPv4
// Shamelessly copied here to avoid importing the entire package
//
// isIPv4 checks if the string is an IP version 4.
func isIPv4(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ".")
}

// Original function from https://pkg.go.dev/github.com/asaskevich/govalidator#IsIPv6
// Shamelessly copied here to avoid importing the entire package
//
// isIPv6 checks if the string is an IP version 6.
func isIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}

func defaultTimeout() time.Duration {
	return 10 * time.Minute
}

//nolint:golint,unused
func newBoolPointer(b bool) *bool {
	return &b
}

// sliceDiff returns a string slice of the elements in `a` that aren't in `b`.
// This function is a bit expensive, but given the fact that
// the expected number of elements is relatively slow
// it's not a big deal.
func sliceDiff[S []E, E any](a, b S) S {
	mb := make(map[any]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}

	var diff S
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
