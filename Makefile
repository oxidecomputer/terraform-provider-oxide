# Build variables
VERSION ?= $(shell cat $(CURDIR)/VERSION)
BINARY ?= terraform-provider-oxide_$(VERSION)
BINARY_LOCATION ?= bin/$(BINARY)
OS_ARCH ?= $(shell go env GOOS)_$(shell go env GOARCH)
RELEASE_VERSION ?= $(shell cat $(CURDIR)/VERSION | sed s/-dev//g)
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
TEST_ACC ?= github.com/oxidecomputer/terraform-provider-oxide/internal/provider
TEST_ACC_NAME ?= TestAccCloudResourceInstance
TEST_ACC_PARALLEL = 6

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

## Lints all of the source files
.PHONY: lint
lint: golangci-lint tfproviderdocs terrafmt tfproviderlint # configfmt

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

configfmt:
	@ echo "-> Running terraform linters on .tf files"
	@ terraform fmt -write=false -recursive -check

.PHONY: testacc
## Runs the Terraform acceptance tests. Use TEST_ACC_NAME, TEST_ACC_ARGS, TEST_ACC_COUNT and TEST_ACC_PARALLEL for acceptance testing settings.
testacc:
	@ echo "-> Running terraform acceptance tests"
	@ TF_ACC=1 go test $(TEST_ACC) -v -count $(TEST_ACC_COUNT) -parallel $(TEST_ACC_PARALLEL) $(TEST_ACC_ARGS) -timeout 20m -run $(TEST_ACC_NAME)

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
VERSION_GOLANGCILINT:=v1.55.2
VERSION_TFPROVIDERDOCS:=v0.9.1
VERSION_TERRAFMT:=v0.5.2
VERSION_TFPROVIDERLINT:=v0.29.0
VERSION_WHATSIT:=7fd2b385f

tools: $(GOBIN)/golangci-lint $(GOBIN)/tfproviderdocs $(GOBIN)/terrafmt $(GOBIN)/tfproviderlint 

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

