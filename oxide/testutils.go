// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oxide

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	oxideSDK "github.com/oxidecomputer/oxide.go/oxide"
)

// TODO: Use this prefix + random string for all resource names for uniqueness
// const accPrefix = "terraform-acc-"

// TODO: Remove once migration is done
var testAccProviderFactory = map[string]func() (*schema.Provider, error){
	"oxide": providerFactory,
}

// TODO: Remove once migration is done
func providerFactory() (*schema.Provider, error) {
	return &schema.Provider{}, nil
	//return Provider(), nil
}

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"oxide": providerserver.NewProtocol6WithError(New()),
	}
}

func testAccPreCheck(t *testing.T) {
	host, token := setAccFromEnvVar()

	if host == "" || token == "" {
		t.Fatal("Both host and token need to be set to execute acceptance tests")
	}
}

func newTestClient() (*oxideSDK.Client, error) {
	host, token := setAccFromEnvVar()

	client, err := oxideSDK.NewClient(token, "terraform-provider-oxide-test", host)
	if err != nil {
		return nil, err
	}

	return client, nil

}

func setAccFromEnvVar() (string, string) {
	// TODO: Unsure if I should only keep the tests tokens,
	// but will leave like this for now
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

	return host, token
}
