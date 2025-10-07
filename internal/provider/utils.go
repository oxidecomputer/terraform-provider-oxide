// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oxidecomputer/oxide.go/oxide"
)

// replaceBackticks replaces '' with `. It can be used to defined codeblocks in
// markdown raw strings.
//
//	var mdString = replaceBackticks(`this is a ''code'' block`)
func replaceBackticks(s string) string {
	return strings.ReplaceAll(s, "''", "`")
}

func is404(err error) bool {
	return strings.Contains(err.Error(), "Status: 404")
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

// newNameOrIdList takes a terraform set and converts is into a slice NameOrIds.
func newNameOrIdList(nameOrIDs types.Set) ([]oxide.NameOrId, diag.Diagnostics) {
	var diags diag.Diagnostics
	var list = []oxide.NameOrId{}
	for _, item := range nameOrIDs.Elements() {
		id, err := strconv.Unquote(item.String())
		if err != nil {
			diags.AddError(
				"Error retrieving name or ID information",
				"name or ID parse error: "+err.Error(),
			)
			return []oxide.NameOrId{}, diags
		}

		n := oxide.NameOrId(id)
		list = append(list, n)
	}

	return list, diags
}
