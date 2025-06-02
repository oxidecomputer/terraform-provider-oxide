---
page_title: "Oxide Provider"
---

# Oxide Provider

The Oxide Terraform provider can be used to manage an Oxide rack.

## Authentication

As a preferred method of authentication, export the `OXIDE_HOST` and `OXIDE_TOKEN` environment variables with their corresponding values.

There are a two alternatives to this:

### Host and Token Arguments
It is possible to authenticate via the optional `host` and `token` arguments. In most cases this method of authentication is not recommended. It is generally preferable to keep credential information out of the configuration.

### Profile Argument
Another option is to use profile-based authentication by passing the `profile` argument. If you have authenticated using the Oxide CLI with `oxide auth login --host https://$YourSiloDnsName`, a `profile` will be created in your `credentials.toml`. You can reference this `profile` directly in your Terraform provider block.

Note: Cannot use `profile` with `host` and `token` arguments and vice versa.

## Example Usage

```hcl
terraform {
  required_version = ">= 1.11"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = "0.9.0"
    }
  }
}

provider "oxide" {
  # The provider will default to use $OXIDE_HOST and $OXIDE_TOKEN.
  # If necessary they can be set explicitly (not recommended).
  # host = "<host address>"
  # token = "<token value>"

  # Can pass in a existing profile that exists in the credentials.toml
  # profile = "<profile name>"
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

## Schema

### Optional

- `host` (String) URL of the root of the target server
- `token` (String, Sensitive) Token used to authenticate
- `profile` (String, Sensitive) Profile used to authenticate from `credentials.toml`
