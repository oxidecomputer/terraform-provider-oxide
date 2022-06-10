VERSION ?= 0.1.0-dev
BINARY ?= terraform-provider-oxide_$(VERSION)
BINARY_LOCATION ?= bin/$(BINARY)
OS_ARCH ?= $(shell go env GOOS)_$(shell go env GOARCH)
PROVIDER_PATH ?= registry.terraform.io/oxidecomputer/oxide/$(VERSION)/$(OS_ARCH)/
PLUGIN_LOCATION ?= ~/.terraform.d/plugins/$(PROVIDER_PATH)

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
