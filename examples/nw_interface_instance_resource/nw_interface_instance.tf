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

data "oxide_projects" "project_list" {}

resource "oxide_instance" "example" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "a test instance"
  name              = "myinstance3"
  host_name         = "myhost"
  memory            = 1073741824
  ncpus             = 1
  network_interface = [
    {
      description = "a network interface"
      name        = "mynetworkinterface"
      subnet_name = "default"
      vpc_name    = "default"
    }
  ]
}
