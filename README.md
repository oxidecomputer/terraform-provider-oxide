# Terraform Provider Oxide

## Build status

| Branch | |
| --- | --- |
| main | [![main](https://github.com/oxidecomputer/terraform-provider-oxide/actions/workflows/build-test.yml/badge.svg?branch=main)](https://github.com/oxidecomputer/terraform-provider-oxide/actions/workflows/build-test.yml) |
| 0.17 | [![0.17](https://github.com/oxidecomputer/terraform-provider-oxide/actions/workflows/build-test.yml/badge.svg?branch=0.17)](https://github.com/oxidecomputer/terraform-provider-oxide/actions/workflows/build-test.yml) |

## Requirements

- [Terraform](https://www.terraform.io/downloads) 1.11.x and above, we recommend using the latest stable release whenever possible. When installing on an Illumos machine use the Solaris binary.

## Using the provider

As a preferred method of authentication, export the `OXIDE_HOST` and `OXIDE_TOKEN` environment variables with their corresponding values.

Alternatively, it is possible to authenticate via the optional `host` and `token` arguments. In most cases this method of authentication is not recommended. It is generally preferable to keep credential information out of the configuration.

To generate a token, follow these steps:

- Make sure you have installed the Oxide CLI
- Log in via the Oxide console.
- Run `oxide auth login --host <host>`
- Retrieve the token associated with the host from `$HOME/.config/oxide/credentials.toml`.

### Example

```hcl
terraform {
  required_version = ">= 1.11"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = "0.18.0"
    }
  }
}

provider "oxide" {
  # The provider will default to use $OXIDE_HOST and $OXIDE_TOKEN.
  # If necessary they can be set explicitly (not recommended).
  # host = "<host address>"
  # token = "<token value>"
}

# Create a blank disk
resource "oxide_disk" "example" {
  project_id  = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a test disk"
  name        = "mydisk"
  size        = 1073741824
  block_size  = 512
}
```

There are several examples in the [examples/](./examples/) directory.

## Development guides and contributing information

Read [CONTRIBUTING.md](./CONTRIBUTING.md) to learn more.
