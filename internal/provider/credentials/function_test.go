package credentials_test

import (
	"os"
	"path"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"

	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/credentials"
	"github.com/oxidecomputer/terraform-provider-oxide/internal/provider/sharedtest"
)

var credentialsToml = `
[profile.alpha]
host = "https://alpha.example.com"
token = "oxide-token-alpha"
token_id = "f431a587-bb8b-41c7-8ccb-6a9e59240bc7"
user = "14e23a0b-5fbd-411c-9385-3a7d2462867e"

[profile.beta]
host = "https://beta.example.com"
token = "oxide-token-beta"
token_id = "3e70d706-5e18-47bb-8205-cd48552fd6da"
user = "85044b9a-7a89-411d-b634-fddba4713d42"
`

var functionTpl = `
locals {
  creds = provider::oxide::credentials({{.Path}})
}

output "alpha_token" {
  value = local.creds["alpha"].token
}

output "beta_host" {
  value = local.creds["beta"].host
}
`

type functionTplConfig struct {
	Path string
}

func TestFunction(t *testing.T) {
	// Temporarily override $HOME to test reading credentials from the default
	// file path.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Create temporary credential file.
	credsPath := path.Join(tmpHome, credentials.DefaultPath)
	if err := os.MkdirAll(path.Dir(credsPath), 0777); err != nil {
		t.Fatalf("error creating test credentials file: %v", err)
	}
	if err := os.WriteFile(credsPath, []byte(credentialsToml), 0666); err != nil {
		t.Fatalf("error creating test credentials file: %v", err)
	}

	config := func(path string) string {
		return sharedtest.ParsedAccConfig(t,
			functionTplConfig{Path: path},
			functionTpl,
		)
	}

	stateChecks := []statecheck.StateCheck{
		statecheck.ExpectKnownOutputValue(
			"alpha_token",
			knownvalue.StringExact("oxide-token-alpha"),
		),
		statecheck.ExpectKnownOutputValue(
			"beta_host",
			knownvalue.StringExact("https://beta.example.com"),
		),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { sharedtest.PreCheck(t) },
		ProtoV6ProviderFactories: sharedtest.ProviderFactories(),
		Steps: []resource.TestStep{
			{
				// Use default credential file if empty.
				Config:            config(`""`),
				ConfigStateChecks: stateChecks,
			},
			{
				// Use default credential file if null.
				Config:            config("null"),
				ConfigStateChecks: stateChecks,
			},
			{
				// Use specified file.
				Config:            config(`"` + credsPath + `"`),
				ConfigStateChecks: stateChecks,
			},
			{
				// Invalid file.
				Config:      config(`"doesnt-exist"`),
				ExpectError: regexp.MustCompile(`doesnt-exist`),
			},
		},
	})
}
