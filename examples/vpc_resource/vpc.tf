terraform {
  required_version = ">= 1.11"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = "0.18.0"
    }
  }
}

provider "oxide" {}

data "oxide_project" "example" {
  name = "{YOUR-PROJECT-NAME}"
}

resource "oxide_vpc" "example" {
  project_id        = data.oxide_project.example.id
  description       = "a test vpc"
  name              = "myvpc"
  dns_name          = "my-vpc-dnssd"
  ipv6_prefix       = "fd1e:4947:d4a1::/48"
}
