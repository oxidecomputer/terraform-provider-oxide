---
page_title: "Oxide Provider"
---

# Oxide Provider

The Oxide Terraform provider can be used to manage an Oxide rack.

## Authentication

As a preferred method of authentication, export the `OXIDE_HOST` and `OXIDE_TOKEN` environment variables with their corresponding values.

Alternatively, it is possible to authenticate via the optional `host` and `token` arguments. In most cases this method of authentication is not recommended. It is generally preferable to keep credential information out of the configuration.

## Example Usage

```hcl
terraform {
  required_version = ">= 1.0"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = ">= 0.1.0"
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

## Schema

### Optional

- `host` (String) URL of the root of the target server
- `token` (String, Sensitive) Token used to authenticate
