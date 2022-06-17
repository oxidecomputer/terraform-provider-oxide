VERSION ?= 0.1.0-dev
BINARY ?= terraform-provider-oxide_$(VERSION)
BINARY_LOCATION ?= bin/$(BINARY)
OS_ARCH ?= $(shell go env GOOS)_$(shell go env GOARCH)
PROVIDER_PATH ?= registry.terraform.io/oxidecomputer/oxide/$(VERSION)/$(OS_ARCH)/
PLUGIN_LOCATION ?= ~/.terraform.d/plugins/$(PROVIDER_PATH)

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