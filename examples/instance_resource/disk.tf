terraform {
  required_version = ">= 0.12"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = "0.1.0-dev"
    }
  }
}

provider "oxide" {
  host = "http://127.0.0.1:12220"
  token = "oxide-spoof-001de000-05e4-4000-8000-000000004007"
}

resource "oxide_instance" "example" {
  organization_name = "corp"
  project_name = "test"
  description = "a test instance"
  name = "myinstance"
  host_name = "myhost"
  memory = 512
  ncpus = 1
}
