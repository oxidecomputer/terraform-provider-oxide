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

DOCKER_COMPOSE_FLAGS = --project-directory ./acctest --file ./acctest/docker-compose.yaml
ifneq ($(ARCH), amd64)
	DOCKER_COMPOSE_FLAGS += --file ./acctest/docker-compose-amd64.yaml
endif

PROVIDER_PATH ?= registry.terraform.io/oxidecomputer/oxide/$(VERSION)/$(OS_ARCH)/
PLUGIN_LOCATION ?= ~/.terraform.d/plugins/$(PROVIDER_PATH)

# Acceptance test variables
TEST_ACC_COUNT ?= 1
TEST_ACC ?= github.com/oxidecomputer/terraform-provider-oxide/internal/provider
TEST_ACC_NAME ?= TestAcc
TEST_ACC_PARALLEL = 6
TEST_ACC_OMICRON_BRANCH ?= main

# Unit test variables
TEST_ARGS ?= -timeout 10m -race -cover
TEST_PACKAGE ?= ./internal/...

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
	@ go test $(TEST_PACKAGE) $(TEST_ARGS) $(TESTUNITARGS)

.PHONY: docs
docs: tools
	@ $(GOBIN)/tfplugindocs generate

.PHONY: check-docs
check-docs: tools
	@ $(GOBIN)/tfplugindocs generate
	@ if ! git diff --exit-code docs; then echo 'Generated docs have changed. Re-generate with `make docs`.'; fi

## Lints all of the source files
.PHONY: lint
lint: golangci-lint tfproviderdocs terrafmt tfproviderlint check-docs # configfmt

.PHONY: tfproviderlint
tfproviderlint: tools
	@ echo "-> Running Terraform static analysis linter"
	@ $(GOBIN)/tfproviderlint ./...

.PHONY: tfproviderdocs
tfproviderdocs: tools
	@ echo "-> Running terraform provider documentation linter"
	@ $(GOBIN)/tfproviderdocs check -provider-name $(BINARY) .

.PHONY: golangci-lint
golangci-lint: tools
	@ echo "-> Running Go linters"
	@ $(GOBIN)/golangci-lint run -E gofmt

.PHONY: terrafmt
terrafmt: tools
	@ echo "-> Running terraform docs codeblocks linter"
	@ find ./docs -type f -name "*.md" -exec $(GOBIN)/terrafmt diff -f {} \;

.PHONY: terrafmt-fmt
terrafmt-fmt: tools
	@ echo "-> Running terraform docs codeblocks linter"
	@ find ./docs -type f -name "*.md" -exec $(GOBIN)/terrafmt fmt -f {} \;

configfmt:
	@ echo "-> Running terraform linters on .tf files"
	@ terraform fmt -write=false -recursive -check

.PHONY: testacc-sim
## Starts a simulated omicron environment suitable to run the acceptance test suite.
testacc-sim: testacc-sim-docker testacc-sim-setup

.PHONY: testacc-sim-down
## Stops the containers of the simulated acceptance test suite environment.
testacc-sim-down:
	@ docker compose $(DOCKER_COMPOSE_FLAGS) down

.PHONY: testacc-sim-docker
## Starts the containers for the simulated acceptance test suite environment.
testacc-sim-docker: export TEST_ACC_DOCKER_TAG = $(shell echo '$(TEST_ACC_OMICRON_BRANCH)' | sed 's/[^[:alnum:]]/_/g')
testacc-sim-docker:
	@ docker compose $(DOCKER_COMPOSE_FLAGS) build \
		--build-arg 'OMICRON_BRANCH=$(TEST_ACC_OMICRON_BRANCH)'
	@ docker compose $(DOCKER_COMPOSE_FLAGS) up --wait --wait-timeout 1500

.PHONY: testacc-sim-token
## Generates an auth token for the simulated acceptance test suite environment.
testacc-sim-token:
	@ uv run ./acctest/auth.py > ./acctest/oxide-token

.PHONY: testacc-sim-setup
## Configures the simulated acceptance test suite environment.
testacc-sim-setup: testacc-sim-token
	@ OXIDE_TOKEN=$(shell cat ./acctest/oxide-token) OXIDE_HOST=http://localhost:12220 ./scripts/acc-test-setup.sh

.PHONY: testacc
## Runs the Terraform acceptance tests. Use TEST_ACC_NAME, TEST_ACC_ARGS, TEST_ACC_COUNT and TEST_ACC_PARALLEL for acceptance testing settings.
testacc:
	@ echo "-> Running terraform acceptance tests"
	@ TF_ACC=1 go test $(TEST_ACC) -v -count $(TEST_ACC_COUNT) -parallel $(TEST_ACC_PARALLEL) $(TEST_ACC_ARGS) -timeout 20m -run $(TEST_ACC_NAME)

.PHONY: testacc-local
## Runs the Terraform acceptance tests locally using the simulated acceptance test suite environment.
## Use TEST_ACC_NAME, TEST_ACC_ARGS, TEST_ACC_COUNT and TEST_ACC_PARALLEL for acceptance testing settings.
testacc-local: export OXIDE_HOST=http://localhost:12220
testacc-local: export OXIDE_TOKEN=$(shell cat ./acctest/oxide-token)
testacc-local: testacc

.PHONY: local-api
## Use local API language client
local-api:
	@ go mod edit -replace=github.com/oxidecomputer/oxide.go=../oxide.go

.PHONY: unset-local-api
## Removes local API language client
unset-local-api:
	@ go mod edit -dropreplace=github.com/oxidecomputer/oxide.go

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

# The following installs the necessary tools within the local /bin directory.
# This way linting tools don't need to be downloaded/installed every time you
# want to run the linters.
VERSION_DIR:=$(GOBIN)/versions
VERSION_GOLANGCILINT:=v1.64.8
VERSION_TFPROVIDERDOCS:=v0.12.1
VERSION_TERRAFMT:=v0.5.4
VERSION_TFPROVIDERLINT:=v0.31.0
VERSION_TFPLUGINDOCS:=v0.24.0
VERSION_WHATSIT:=053446d

tools: $(GOBIN)/golangci-lint $(GOBIN)/tfproviderdocs $(GOBIN)/terrafmt $(GOBIN)/tfproviderlint $(GOBIN)/tfplugindocs

tools-private: $(GOBIN)/whatsit

$(GOBIN):
	@ mkdir -p $(GOBIN)

$(VERSION_DIR): | $(GOBIN)
	@ mkdir -p $(GOBIN)/versions

$(VERSION_DIR)/.version-golangci-lint-$(VERSION_GOLANGCILINT): | $(VERSION_DIR)
	@ rm -f $(VERSION_DIR)/.version-golangci-lint-*
	@ echo $(VERSION_GOLANGCILINT) > $(VERSION_DIR)/.version-golangci-lint-$(VERSION_GOLANGCILINT)

$(GOBIN)/golangci-lint: $(VERSION_DIR)/.version-golangci-lint-$(VERSION_GOLANGCILINT) | $(GOBIN)
	@ echo "-> Installing golangci-lint..."
	@ curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOBIN) $(VERSION_GOLANGCILINT)

$(VERSION_DIR)/.version-tfproviderdocs-$(VERSION_TFPROVIDERDOCS): | $(VERSION_DIR)
	@ rm -f $(VERSION_DIR)/.version-tfproviderdocs-*
	@ echo $(VERSION_TFPROVIDERDOCS) > $(VERSION_DIR)/.version-tfproviderdocs-$(VERSION_TFPROVIDERDOCS)

$(GOBIN)/tfproviderdocs: $(VERSION_DIR)/.version-tfproviderdocs-$(VERSION_TFPROVIDERDOCS) | $(GOBIN)
	@ echo "-> Installing tfproviderdocs..."
	@ go install github.com/bflad/tfproviderdocs@$(VERSION_TFPROVIDERDOCS)

$(VERSION_DIR)/.version-terrafmt-$(VERSION_TERRAFMT): | $(VERSION_DIR)
	@ rm -f $(VERSION_DIR)/.version-terrafmt-*
	@ echo $(VERSION_TERRAFMT) > $(VERSION_DIR)/.version-terrafmt-$(VERSION_TERRAFMT)

$(GOBIN)/terrafmt: $(VERSION_DIR)/.version-terrafmt-$(VERSION_TERRAFMT) | $(GOBIN)
	@ echo "-> Installing terrafmt..."
	@ go install github.com/katbyte/terrafmt@$(VERSION_TERRAFMT)

$(VERSION_DIR)/.version-tfproviderlint-$(VERSION_TFPROVIDERLINT): | $(VERSION_DIR)
	@ rm -f $(VERSION_DIR)/.version-tfproviderlint-*
	@ echo $(VERSION_TFPROVIDERLINT) > $(VERSION_DIR)/.version-tfproviderlint-$(VERSION_TFPROVIDERLINT)

$(GOBIN)/tfproviderlint: $(VERSION_DIR)/.version-tfproviderlint-$(VERSION_TFPROVIDERLINT) | $(GOBIN)
	@ echo "-> Installing tfproviderlint..."
	@ go install github.com/bflad/tfproviderlint/cmd/tfproviderlint@$(VERSION_TFPROVIDERLINT)

$(VERSION_DIR)/.version-whatsit-$(VERSION_WHATSIT): | $(VERSION_DIR)
	@ rm -f $(VERSION_DIR)/.version-whatsit-*
	@ echo $(VERSION_WHATSIT) > $(VERSION_DIR)/.version-whatsit-$(VERSION_WHATSIT)

# TODO: actually release a version of whatsit to use the tag flag
$(GOBIN)/whatsit: $(VERSION_DIR)/.version-whatsit-$(VERSION_WHATSIT) | $(GOBIN)
	@ echo "-> Installing whatsit..."
	@ cargo install --git ssh://git@github.com/oxidecomputer/whatsit.git#$(VERSION_WHATSIT) --branch main --root ./

$(GOBIN)/tfplugindocs:
	@ echo "-> Installing tfplugindocs..."
	@ go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@$(VERSION_TFPLUGINDOCS)
