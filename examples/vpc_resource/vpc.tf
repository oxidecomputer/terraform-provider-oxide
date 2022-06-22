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

resource "oxide_vpc" "example" {
  organization_name = "corp"
  project_name      = "test"
  description       = "a test vpc"
  name              = "myvpc"
  dns_name          = "my-vpc-dns"
}
