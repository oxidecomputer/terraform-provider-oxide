# Build variables
VERSION ?= 0.1.0-dev
BINARY ?= terraform-provider-oxide_$(VERSION)
BINARY_LOCATION ?= bin/$(BINARY)
OS_ARCH ?= $(shell go env GOOS)_$(shell go env GOARCH)

# Terraform currently does not have a binary for Illumos.
# The one for Solaris works fine with Illumos, so we'll need
# to make sure we install the plugin in the solaris directory.
ifeq ($(shell go env GOOS), illumos)
    OS_ARCH = solaris_$(shell go env GOARCH)
endif

PROVIDER_PATH ?= registry.terraform.io/oxidecomputer/oxide/$(VERSION)/$(OS_ARCH)/
PLUGIN_LOCATION ?= ~/.terraform.d/plugins/$(PROVIDER_PATH)

# Acceptance test variables
TEST_COUNT ?= 1
TEST_ACC ?= github.com/oxidecomputer/terraform-provider-oxide/oxide
TEST_NAME ?= TestAcc
TEST_ACC_PARALLEL = 6

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

## Lints all of the source files
.PHONY: lint
lint: golint golangci-lint tfproviderdocs terrafmt configfmt # tfproviderlint

# tf providerlint currently has a bug with the latest stable go version (1.18), will uncomment 
# when this issue is solved https://github.com/bflad/tfproviderlint/issues/255
# .PHONY: tfproviderlint
# tfproviderlint: tools
# 	@ echo "-> Checking source code against terraform provider linters..."
# 	@ $(GOBIN)/tfproviderlint ./...

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
	@ find ./docs -type f -name "*.md" -exec $(GOBIN)/terrafmt diff -c -q {} \;

configfmt:
	@ echo "-> Checking that the terraform .tf files are formatted..."
	@ terraform fmt -write=false -recursive -check

.PHONY: testacc
## Runs the Terraform acceptance tests. Use TEST_NAME, TESTARGS, TEST_COUNT and TEST_ACC_PARALLEL to control execution.
testacc:
	@ echo "-> Running terraform acceptance tests..."
	@ TF_ACC=1 go test $(TEST_ACC) -v -count $(TEST_COUNT) -parallel $(TEST_ACC_PARALLEL) $(TESTARGS) -timeout 20m -run $(TEST_NAME)

.PHONY: local-api
## Use local API language client
local-api:
	@ go mod edit -replace=github.com/oxidecomputer/oxide.go=../oxide.go

.PHONY: unset-local-api
## Removes local API language client
unset-local-api:
	@ go mod edit -dropreplace=github.com/oxidecomputer/oxide.go
