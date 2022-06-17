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

resource "oxide_instance" "example" {
  organization_name = "corp"
  project_name      = "test"
  description       = "a test instance"
  name              = "myinstance"
  host_name         = "myhost"
  memory            = 512
  ncpus             = 1
}
