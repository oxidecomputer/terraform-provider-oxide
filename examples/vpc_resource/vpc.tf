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

data "oxide_organizations" "org_list" {}

data "oxide_projects" "project_list" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
}

resource "oxide_vpc" "example" {
  project_id        = data.oxide_projects.project_list.projects.0.id
  description       = "a test vpc"
  name              = "myvpc"
  dns_name          = "my-vpc-dnssd"
  ipv6_prefix       = "fd1e:4947:d4a1::/48"
}
