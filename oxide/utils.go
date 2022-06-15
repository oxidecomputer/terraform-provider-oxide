// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"fmt"
	"strings"
)

func is404(err error) bool {
	if strings.Contains(err.Error(), fmt.Sprintf("404 Not Found")) {
		return true
	}

	return false
}
