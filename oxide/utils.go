// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

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

func newBoolPointer(b bool) *bool {
	return &b
}
