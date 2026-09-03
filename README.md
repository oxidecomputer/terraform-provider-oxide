# Oxide Terraform Provider

The Oxide Terraform provider declaratively manages
[Oxide](https://oxide.computer) resources with Terraform or OpenTofu.

The provider uses the [Oxide Go SDK](https://github.com/oxidecomputer/oxide.go)
to create, read, update, and delete Oxide resources.

## Build Status

| Branch      | Status |
| ----------- | ------ |
| `main`      | [![main](https://github.com/oxidecomputer/terraform-provider-oxide/actions/workflows/build-test.yml/badge.svg?branch=main)](https://github.com/oxidecomputer/terraform-provider-oxide/actions/workflows/build-test.yml?query=branch%3Amain) |
| `rel/v0.21` | [![0.21](https://github.com/oxidecomputer/terraform-provider-oxide/actions/workflows/build-test.yml/badge.svg?branch=rel%2Fv0.21)](https://github.com/oxidecomputer/terraform-provider-oxide/actions/workflows/build-test.yml?query=branch%3Arel%2Fv0.21) |

## Version Policy

This project adheres to [Semantic Versioning](https://semver.org/). It is
currently at major version zero (e.g., v0.Y.Z). Please note the following
semantics:

- The minimum supported Oxide version may change across minor version releases.
- Configuration changes may be required across minor version releases.

Read the [upgrade guide](./docs/guides/upgrade.md) before upgrading.

### Minimum Supported Terraform and OpenTofu Versions

Because this provider uses write-only attributes, it requires the following
minimum versions:

- Terraform 1.11
- OpenTofu 1.11

## Usage

Create a `main.tf` with the following configuration.

```hcl
terraform {
  required_version = ">= 1.11"

  required_providers {
    oxide = {
      source = "oxidecomputer/oxide"
    }
  }
}

# The provider defaults to using environment variables for authentication.
#
# - OXIDE_HOST: Oxide API host (e.g., https://oxide.sys.example.com)
# - OXIDE_TOKEN: Oxide API token (e.g., oxide-token-abc123)
provider "oxide" {}

resource "oxide_project" "example" {
  name        = "my-project"
  description = "A project managed by the Oxide provider."
}
```

For other authentication methods, see the
[provider documentation](https://registry.terraform.io/providers/oxidecomputer/oxide/latest/docs).

OpenTofu users can run the equivalent commands below by replacing `terraform`
with `tofu`.

Initialize the configuration and install the provider:

```shell
terraform init
```

Create the Oxide project:

```shell
terraform apply
```

Delete the Oxide project:

```shell
terraform destroy
```

## Documentation

Refer to the [Oxide Terraform Provider](https://registry.terraform.io/providers/oxidecomputer/oxide/latest)
documentation to learn which resources, data sources, and functions are
supported by this provider.

The registry documentation is rendered from [docs](./docs), which is generated
from [templates](./templates), [examples](./examples), and the provider source
code.

## Contributing

Read [CONTRIBUTING.md](./CONTRIBUTING.md) before contributing to this project.
