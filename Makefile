# Build variables
SHELL := /usr/bin/env bash
VERSION ?= $(shell cat $(CURDIR)/VERSION)
BINARY ?= terraform-provider-oxide_$(VERSION)
BINARY_LOCATION ?= bin/$(BINARY)
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
OS_ARCH ?= $(OS)_$(ARCH)
RELEASE_VERSION ?= $(shell cat $(CURDIR)/VERSION | sed s/-dev//g)
export GOBIN = $(shell pwd)/bin

# Terraform currently does not have a binary for Illumos.
# The one for Solaris works fine with Illumos, so we'll need
# to make sure we install the plugin in the solaris directory.
ifeq ($(OS), illumos)
    OS_ARCH = solaris_$(ARCH)
endif

PROVIDER_PATH ?= registry.terraform.io/oxidecomputer/oxide/$(VERSION)/$(OS_ARCH)/
PLUGIN_LOCATION ?= ~/.terraform.d/plugins/$(PROVIDER_PATH)

# Acceptance test variables
TEST_ACC_COUNT ?= 1
TEST_ACC ?= github.com/oxidecomputer/terraform-provider-oxide/internal/provider
TEST_ACC_NAME ?= TestAcc
TEST_ACC_PARALLEL = 6
TEST_ACC_OMICRON_BRANCH ?=
TEST_ACC_DOCKER_COMPOSE_FLAGS = --project-directory ./acctest --file ./acctest/docker-compose.yaml

# Unit test variables
TEST_ARGS ?= -timeout 10m -race -cover
TEST_PACKAGE ?= ./internal/...

# Helpers
GO_TOOL = go tool -modfile=tools/go.mod

### Build targets

## Builds the source code and saves the binary to bin/.
.PHONY: build
build:
	@ echo "-> Building binary in $(BINARY_LOCATION)"
	@ go build -o $(BINARY_LOCATION) .

## Builds the source code and moves the binary to the local terraform plugin directory.
.PHONY: install
install: build
	@ echo "-> Installing plugin in $(PLUGIN_LOCATION)"
	@ mkdir -p $(PLUGIN_LOCATION)
	@ cp $(BINARY_LOCATION) $(PLUGIN_LOCATION)

## Run unit tests. Use TEST_ARGS to set `go test` CLI arguments, and TEST_UNIT_DIR to set packages to be tested
.PHONY: test
test:
	@ echo "-> Running unit tests for $(BINARY)"
	@ $(GO_TOOL) gotestsum --format=testname --hide-summary=skipped -- $(TEST_PACKAGE) $(TEST_ARGS) $(TESTUNITARGS)

.PHONY: docs
docs:
	@ $(GO_TOOL) tfplugindocs generate

.PHONY: check-docs
check-docs:
	@ $(GO_TOOL) tfplugindocs generate
	@ if ! git diff --exit-code docs; then echo 'Generated docs have changed. Re-generate with `make docs`.'; fi

## Lints all of the source files
.PHONY: lint
lint: golangci-lint tfproviderdocs terrafmt tfproviderlint check-docs # configfmt

.PHONY: tfproviderlint
tfproviderlint:
	@ echo "-> Running Terraform static analysis linter"
	@ $(GO_TOOL) tfproviderlint ./...

.PHONY: tfproviderdocs
tfproviderdocs:
	@ echo "-> Running terraform provider documentation linter"
	@ $(GO_TOOL) tfproviderdocs check -provider-name $(BINARY) .

.PHONY: golangci-lint
golangci-lint:
	@ echo "-> Running Go linters"
	@ $(GO_TOOL) golangci-lint run

.PHONY: fmt
fmt: golangci-fmt terrafmt-fmt docs

.PHONY: golangci-fmt
golangci-fmt:
	@ echo "-> Formatting Go code"
	@ $(GO_TOOL) golangci-lint fmt

.PHONY: terrafmt
terrafmt:
	@ echo "-> Running terraform docs codeblocks linter"
	@ find ./docs -type f -name "*.md" -exec $(GO_TOOL) terrafmt diff -f {} \;

.PHONY: terrafmt-fmt
terrafmt-fmt:
	@ echo "-> Running terraform docs codeblocks linter"
	@ find ./docs -type f -name "*.md" -exec $(GO_TOOL) terrafmt fmt -f {} \;

configfmt:
	@ echo "-> Running terraform linters on .tf files"
	@ terraform fmt -write=false -recursive -check

.PHONY: fmt-md
fmt-md: ## Formats markdown files with prettier.
	@ echo "-> Formatting markdown files"
	@ npx prettier --write "**/*.md"

.PHONY: testacc-sim
## Starts a simulated omicron environment suitable to run the acceptance test suite.
testacc-sim: testacc-sim-docker testacc-sim-setup

.PHONY: testacc-sim-down
## Stops the containers of the simulated acceptance test suite environment.
testacc-sim-down:
	@ docker compose $(TEST_ACC_DOCKER_COMPOSE_FLAGS) down

.PHONY: testacc-sim-docker
## Starts the containers for the simulated acceptance test suite environment.
testacc-sim-docker: export TEST_ACC_DOCKER_TAG = $(shell ./acctest/omicron-version.sh $(TEST_ACC_OMICRON_BRANCH))
testacc-sim-docker:
	@ ./acctest/ensure-image.sh $(TEST_ACC_DOCKER_TAG)
	@ docker compose $(TEST_ACC_DOCKER_COMPOSE_FLAGS) up --wait --wait-timeout 1500

.PHONY: testacc-sim-token
## Generates an auth token for the simulated acceptance test suite environment.
testacc-sim-token:
	@ uv run ./acctest/auth.py > ./acctest/oxide-token

.PHONY: testacc-sim-oxide-cli
## Installs the version of the oxide CLI matching the current omicron version.
testacc-sim-oxide-cli:
	@ ./acctest/ensure-oxide-cli.sh

.PHONY: testacc-sim-setup
## Configures the simulated acceptance test suite environment.
testacc-sim-setup: testacc-sim-token testacc-sim-oxide-cli
	@ PATH=$(GOBIN):$$PATH OXIDE_TOKEN=$(shell cat ./acctest/oxide-token) OXIDE_HOST=http://localhost:12220 ./scripts/acc-test-setup.sh

.PHONY: testacc
## Runs the Terraform acceptance tests. Use TEST_ACC_NAME, TEST_ACC_ARGS, TEST_ACC_COUNT and TEST_ACC_PARALLEL for acceptance testing settings.
testacc:
	@ echo "-> Running terraform acceptance tests"
	@ TF_ACC=1 $(GO_TOOL) gotestsum \
	    --format=testname --rerun-fails=3 --packages=$(TEST_ACC)/... \
	    -- $(TEST_ACC) -v -count $(TEST_ACC_COUNT) -parallel $(TEST_ACC_PARALLEL) $(TEST_ACC_ARGS) -timeout 20m -run $(TEST_ACC_NAME)

.PHONY: testacc-local
## Runs the Terraform acceptance tests locally using the simulated acceptance test suite environment.
## Use TEST_ACC_NAME, TEST_ACC_ARGS, TEST_ACC_COUNT and TEST_ACC_PARALLEL for acceptance testing settings.
testacc-local: export OXIDE_HOST=http://localhost:12220
testacc-local: export OXIDE_TOKEN=$(shell cat ./acctest/oxide-token)
testacc-local: testacc

.PHONY: local-api
## Use local Go SDK.
local-api:
	@ echo "Initializing go.work"
	@ go work init 2> /dev/null || true
	@ go work use . ../oxide.go

.PHONY: unset-local-api
## Stop using local Go SDK.
unset-local-api:
	rm -f go.work go.work.sum

.PHONY: changelog
## Creates a changelog prior to a release
changelog: tools-private
	@ echo "-> Creating changelog"
	@ $(GOBIN)/whatsit changelog create --repository oxidecomputer/terraform-provider-oxide -n $(RELEASE_VERSION) -c ./.changelog/$(RELEASE_VERSION).toml

.PHONY: tag
tag: ## Create a new git tag to prepare to build a release.
	git tag v$(RELEASE_VERSION)
	@echo "Run 'git push origin v$(RELEASE_VERSION)' to push your new tag to GitHub and trigger a release."

.PHONY: sdk-version
## Sets Oxide Go SDK to a specified version
sdk-version:
	@ echo "-> Setting Oxide Go SDK to oxide.go@$(SDK_V)"
	@ go get github.com/oxidecomputer/oxide.go@$(SDK_V)
	@ go mod tidy

# whatsit is a Rust tool used for changelog generation, installed via cargo.
VERSION_WHATSIT:=053446d

tools-private: $(GOBIN)/whatsit

$(GOBIN):
	@ mkdir -p $(GOBIN)

# TODO: actually release a version of whatsit to use the tag flag
$(GOBIN)/whatsit: | $(GOBIN)
	@ echo "-> Installing whatsit..."
	@ cargo install --git ssh://git@github.com/oxidecomputer/whatsit.git#$(VERSION_WHATSIT) --branch main --root ./
