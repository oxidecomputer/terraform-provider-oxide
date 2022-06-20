// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

var (
	testProvider  *schema.Provider
	testProviders map[string]*schema.Provider
)

func init() {
	testProvider = Provider()
	testProviders = map[string]*schema.Provider{
		"oxide": testProvider,
	}
}
