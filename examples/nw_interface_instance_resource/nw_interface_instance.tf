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

resource "oxide_instance" "example" {
  project_id        = data.oxide_projects.project_list.projects.0.id
  description       = "a test instance"
  name              = "myinstance3"
  host_name         = "myhost"
  memory            = 1073741824
  ncpus             = 1
  network_interface {
    description = "a network interface"
    name        = "mynetworkinterface"
    subnet_name = "default"
    vpc_name    = "default"
  }
}
