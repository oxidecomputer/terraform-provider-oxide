# Terraform Provider Oxide

## Code Style

After making changes, run the linter:

```
make lint
```

If this target fails, run `make fmt` to auto-format the code and docs, then verify with `make lint`.
If the check continues to fail, investigate any errors and update the code to fix.

Prefer to use make targets rather than ad-hoc `go run` or `go build` commands. If the Makefile is
missing a useful target, propose adding it.

## Testing

Run unit tests with:

```
make test
```

### Acceptance Tests

Acceptance tests run against a simulated omicron environment in Docker. While working on an
individual resource, prefer to run just the relevant tests, but always run the full suite before
committing changes.

Each resource and data source should include acceptance test suites. Minimally, the test suite
should create, update, and delete a resource, and import it if supported.

#### Setup

Start the simulated environment. This resolves the omicron version from go.mod, pulls the docker
image from GHCR (or builds locally if unavailable), installs the matching oxide CLI to `./bin/`, and
configures test resources:

```
make testacc-sim
```

#### Running tests

```
make testacc-local
```

To run a specific test:

```
make testacc-local TEST_ACC_NAME=TestAccCloudResourceInstance_full
```

#### Teardown

```
make testacc-sim-down
```
