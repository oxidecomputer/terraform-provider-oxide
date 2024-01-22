// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/oxidecomputer/oxide.go/oxide"
)

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

func newTestClient() (*oxide.Client, error) {
	host, token := setAccFromEnvVar()

	config := oxide.Config{
		Token:     token,
		UserAgent: "terraform-provider-oxide-test",
		Host:      host,
	}
	client, err := oxide.NewClient(&config)
	if err != nil {
		return nil, err
	}

	return client, nil

}

func parsedAccConfig(config any, tpl string) (string, error) {
	var buf bytes.Buffer
	tmpl, _ := template.New("test").Parse(tpl)
	err := tmpl.Execute(&buf, config)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
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

func newResourceName() string {
	return fmt.Sprintf("acc-terraform-%s", uuid.New())
}

func newBlockName(resource string) string {
	return fmt.Sprintf("acc-%s-%s", resource, uuid.New())
}
