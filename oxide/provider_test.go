// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"os"
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

const accPrefix = "terraform_acc_"

var testAccProviderFactory = map[string]func() (*schema.Provider, error){
	"oxide": providerFactory,
}

func providerFactory() (*schema.Provider, error) {
	return Provider(), nil
}

func testAccPreCheck(t *testing.T) {
	//TODO: Unsure if I should only keep the tests tokens, but will leave like this
	//for now
	var host, token string

	if k := os.Getenv("OXIDE_HOST"); k != "" {
		host = k
	}
	if k := os.Getenv("OXIDE_TEST_HOST"); k != "" {
		host = k
	}

	if k := os.Getenv("OXIDE_TOKEN"); k != "" {
		token = k
	}
	if k := os.Getenv("OXIDE_TEST_TOKEN"); k != "" {
		token = k
	}

	if host == "" || token == "" {
		t.Fatal("No host and token found to execute acceptance tests")
	}
}
