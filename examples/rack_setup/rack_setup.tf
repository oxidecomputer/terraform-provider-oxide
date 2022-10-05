terraform {
  required_version = ">= 1.0"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = "0.1.0-dev"
    }
  }
}

provider "oxide" {}

resource "oxide_organization" "setup_org" {
  description       = "a test org"
  name              = "myorg"
}

resource "oxide_project" "setup_project" {
  description       = "a test project"
  name              = "myproj"
  organization_name = oxide_organization.setup_org.name
}

resource "oxide_ip_pool" "ip_pool_ranges" {
  description = "a test IP pool"
  name        = "mypool"
  ranges {
    first_address = "172.20.15.227"
    last_address  = "172.20.15.239"
  }
}

resource "oxide_global_image" "test" {
  description          = "a test global_image"
  name                 = "alpine"
  image_source         = { you_can_boot_anything_as_long_as_its_alpine = "noop" }
  block_size           = 512
  distribution         = "alpine"
  distribution_version = "propolis_blob"
}