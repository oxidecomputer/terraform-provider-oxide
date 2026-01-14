package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"
)

// lintignore:AT004
func TestProvider(t *testing.T) {
	ts := newTestProviderServer(t)

	configDir := t.TempDir()
	creds, err := os.Create(filepath.Join(configDir, "credentials.toml"))
	require.NoError(t, err)

	testProfile := "test"
	profileToken := "profile-token"
	fmt.Fprintf(creds, `
[profile.%s]
host = "%s"
token = "%s"
token_id = "f5660a9f-962e-4c00-a6dc-638256ae1d4e"
user = "4dc4ba10-ab9e-403f-b08a-a7388df64e3a"
`, testProfile, ts.URL(), profileToken)

	renderConfig := func(provider string, a ...any) string {
		return fmt.Sprintf(`
%s

data "oxide_project" "test" {
  name = "test-project"
}`, fmt.Sprintf(provider, a...))
	}

	testCleanEnv(t, []string{
		"OXIDE_HOST",
		"OXIDE_TOKEN",
		"OXIDE_PROFILE",
	})

	testCases := []struct {
		name        string
		preConfig   func(*testing.T)
		config      string
		checkFunc   func(*testing.T)
		expectError *regexp.Regexp
	}{
		{
			name: "host and token env var",
			preConfig: func(t *testing.T) {
				t.Setenv("OXIDE_HOST", ts.URL())
				t.Setenv("OXIDE_TOKEN", "env-token")
			},
			config: renderConfig(`
provider "oxide" {}
`),
			checkFunc: func(t *testing.T) {
				req := ts.LastRequest()
				requireRequestToken(t, req, "env-token")
			},
		},
		{
			name: "host and token config",
			preConfig: func(t *testing.T) {
				t.Setenv("OXIDE_HOST", "https://example.com")
				t.Setenv("OXIDE_TOKEN", "env-token")
			},
			config: renderConfig(`
provider "oxide" {
  host  = "%s"
  token = "config-token"
}`, ts.URL()),
			checkFunc: func(t *testing.T) {
				req := ts.LastRequest()
				requireRequestToken(t, req, "config-token")
			},
		},
		{
			name: "profile env var",
			preConfig: func(t *testing.T) {
				t.Setenv("OXIDE_PROFILE", testProfile)
			},
			config: renderConfig(`
provider "oxide" {
  config_dir = "%s"
}
`, configDir),
			checkFunc: func(t *testing.T) {
				req := ts.LastRequest()
				requireRequestToken(t, req, profileToken)
			},
		},
		{
			name: "profile config",
			preConfig: func(t *testing.T) {
				t.Setenv("OXIDE_PROFILE", "other-profile")
			},
			config: renderConfig(`
provider "oxide" {
  config_dir = "%s"
  profile    = "%s"
}
`, configDir, testProfile),
			checkFunc: func(t *testing.T) {
				req := ts.LastRequest()
				requireRequestToken(t, req, profileToken)
			},
		},
		{
			name: "skip tls",
			preConfig: func(t *testing.T) {
				t.Setenv("OXIDE_HOST", ts.URLTLS())
				t.Setenv("OXIDE_TOKEN", "env-token")
			},
			config: renderConfig(`
provider "oxide" {
  insecure_skip_verify = true
}
`),
		},
		{
			name: "fail without skip tls",
			preConfig: func(t *testing.T) {
				t.Setenv("OXIDE_HOST", ts.URLTLS())
				t.Setenv("OXIDE_TOKEN", "env-token")
			},
			config: renderConfig(`
provider "oxide" {}
`),
			expectError: regexp.MustCompile("tls: failed to verify"),
		},
		{
			name: "profile conflicts with host and token",
			preConfig: func(t *testing.T) {
				t.Setenv("OXIDE_HOST", ts.URL())
			},
			config: renderConfig(`
provider "oxide" {
  profile = "%s"
  token   = "config-token"
}
`, testProfile),
			expectError: regexp.MustCompile("Invalid Attribute Combination"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
				Steps: []resource.TestStep{
					{
						PreConfig: func() {
							if tc.preConfig != nil {
								tc.preConfig(t)
							}
						},
						Config: tc.config,
						PostApplyFunc: func() {
							if tc.checkFunc != nil {
								tc.checkFunc(t)
							}
						},
						ExpectError: tc.expectError,
					},
				},
			})
		})
	}
}

type testProviderServer struct {
	lock sync.Mutex

	testServer    *httptest.Server
	testServerTLS *httptest.Server

	lastRequest *http.Request
}

func newTestProviderServer(t *testing.T) *testProviderServer {

	tps := &testProviderServer{}

	ts := httptest.NewServer(http.HandlerFunc(tps.handleRequest))
	t.Cleanup(ts.Close)

	tsTLS := httptest.NewTLSServer(http.HandlerFunc(tps.handleRequest))
	t.Cleanup(tsTLS.Close)

	tps.testServer = ts
	tps.testServerTLS = tsTLS

	return tps
}

func (t *testProviderServer) LastRequest() *http.Request {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.lastRequest
}

func (t *testProviderServer) URL() string {
	return t.testServer.URL
}

func (t *testProviderServer) URLTLS() string {
	return t.testServerTLS.URL
}

func (t *testProviderServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.lastRequest = r

	fmt.Fprintln(w, `
{
  "description": "test project",
  "id": "3fa85f64-5717-4562-b3fc-2c963f66afa6",
  "name": "test-project",
  "time_created": "2026-01-13T19:10:21.227Z",
  "time_modified": "2026-01-13T19:10:21.227Z"
}`)
}

func testCleanEnv(t *testing.T, envars []string) {
	for _, envar := range envars {
		if val := os.Getenv(envar); val != "" {
			os.Setenv(envar, "")
			t.Cleanup(func() {
				os.Setenv(envar, val)
			})
		}
	}
}

func requireRequestToken(t *testing.T, req *http.Request, token string) {
	t.Helper()
	requireRequestHeader(t, req, "Authorization", fmt.Sprintf("Bearer %s", token))
}

func requireRequestHeader(t *testing.T, req *http.Request, header string, expected string) {
	t.Helper()
	val := req.Header.Get(header)
	require.Equal(t, expected, val)
}
