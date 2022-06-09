VERSION ?= 0.1.0
BINARY ?= terraform-provider-oxide_$(VERSION)
BINARY_LOCATION ?= bin/$(BINARY)
OS_ARCH ?= $(shell go env GOOS)_$(shell go env GOARCH)
# Switch to use OS_ARCH instead of darwin_arm64 for other OSs
PROVIDER_PATH ?= registry.terraform.io/oxidecomputer/oxide/$(VERSION)/darwin_arm64/
PLUGIN_LOCATION ?= ~/.terraform.d/plugins/$(PROVIDER_PATH)

### Build targets

## Builds the source code and saves the binary to bin/terraform-provider-oxide-demo.
.PHONY: build
build:
	@ echo "-> Building binary in $(BINARY_LOCATION)..."
	@ go build -o $(BINARY_LOCATION) .

## Builds the source code and moves the binary to the user's terraform plugin location.
.PHONY: install
install: build
	@ mkdir -p $(PLUGIN_LOCATION)
	@ cp $(BINARY_LOCATION) $(PLUGIN_LOCATION)
