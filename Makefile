# Build variables
VERSION ?= $(shell cat $(CURDIR)/VERSION)
BINARY ?= terraform-provider-oxide_$(VERSION)
BINARY_LOCATION ?= bin/$(BINARY)
OS_ARCH ?= $(shell go env GOOS)_$(shell go env GOARCH)
export GOBIN = $(shell pwd)/bin

# Terraform currently does not have a binary for Illumos.
# The one for Solaris works fine with Illumos, so we'll need
# to make sure we install the plugin in the solaris directory.
ifeq ($(shell go env GOOS), illumos)
    OS_ARCH = solaris_$(shell go env GOARCH)
endif

PROVIDER_PATH ?= registry.terraform.io/oxidecomputer/oxide/$(VERSION)/$(OS_ARCH)/
PLUGIN_LOCATION ?= ~/.terraform.d/plugins/$(PROVIDER_PATH)

# Acceptance test variables
TEST_ACC_COUNT ?= 1
TEST_ACC ?= github.com/oxidecomputer/terraform-provider-oxide/oxide
TEST_ACC_NAME ?= TestAcc
TEST_ACC_PARALLEL = 6

# Unit test variables
TEST_ARGS ?= -timeout 10m -race -cover
TEST_PACKAGE ?= ./oxide

include Makefile.tools

### Build targets

## Builds the source code and saves the binary to bin/.
.PHONY: build
build:
	@ echo "-> Building binary in $(BINARY_LOCATION)..."
	@ go build -o $(BINARY_LOCATION) .

## Builds the source code and moves the binary to the local terraform plugin directory.
.PHONY: install
install: build
	@ echo "-> Installing plugin in $(PLUGIN_LOCATION)..."
	@ mkdir -p $(PLUGIN_LOCATION)
	@ cp $(BINARY_LOCATION) $(PLUGIN_LOCATION)

## Run unit tests. Use TEST_ARGS to set `go test` CLI arguments, and TEST_UNIT_DIR to set packages to be tested
.PHONY: test
test:
	@ echo "-> Running unit tests for $(BINARY)..."
	@ go test $(TEST_PACKAGE) $(TEST_ARGS) $(TESTUNITARGS)

## Lints all of the source files
.PHONY: lint
lint: golint golangci-lint tfproviderdocs terrafmt tfproviderlint # configfmt

.PHONY: tfproviderlint
tfproviderlint: tools
	@ echo "-> Checking source code against terraform provider linters..."
	@ $(GOBIN)/tfproviderlint ./...

.PHONY: tfproviderdocs
tfproviderdocs: tools
	@ echo "-> Running terraform provider docs check..."
	@ $(GOBIN)/tfproviderdocs check -provider-name $(BINARY) .

.PHONY: golangci-lint
golangci-lint: tools
	@ echo "-> Running golangci-lint..."
	@ $(GOBIN)/golangci-lint run

.PHONY: golint
golint: tools
	@ echo "-> Running golint..."
	@ $(GOBIN)/golint -set_exit_status ./...

.PHONY: terrafmt
terrafmt: tools
	@ echo "-> Checking that the terraform docs codeblocks are formatted..."
	@ find ./docs -type f -name "*.md" -exec $(GOBIN)/terrafmt diff -f {} \;

configfmt:
	@ echo "-> Checking that the terraform .tf files are formatted..."
	@ terraform fmt -write=false -recursive -check

.PHONY: testacc
## Runs the Terraform acceptance tests. Use TEST_ACC_NAME, TEST_ACC_ARGS, TEST_ACC_COUNT and TEST_ACC_PARALLEL for acceptance testing settings.
testacc:
	@ echo "-> Running terraform acceptance tests..."
	@ TF_ACC=1 go test $(TEST_ACC) -v -count $(TEST_ACC_COUNT) -parallel $(TEST_ACC_PARALLEL) $(TEST_ACC_ARGS) -timeout 20m -run $(TEST_ACC_NAME)

.PHONY: local-api
## Use local API language client
local-api:
	@ go mod edit -replace=github.com/oxidecomputer/oxide.go=../oxide.go

.PHONY: unset-local-api
## Removes local API language client
unset-local-api:
	@ go mod edit -dropreplace=github.com/oxidecomputer/oxide.go
