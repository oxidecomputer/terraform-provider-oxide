// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"strings"
)

func is404(err error) bool {
	return strings.Contains(err.Error(), "404 Not Found")
}
