terraform {
  required_version = ">= 1.0"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = "0.1.0"
    }
  }
}

provider "oxide" {}

data "oxide_projects" "project_list" {}

data "oxide_ip_pool" "pool" {
  name = "default"
}

resource "oxide_instance" "example" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "a test instance"
  name              = "myinstance"
  host_name         = "myhost"
  memory            = 1073741824
  ncpus             = 1
  external_ips      = [
    {
      id = data.oxide_ip_pool.pool.id
      type = "ephemeral"
    }
  ]
}
