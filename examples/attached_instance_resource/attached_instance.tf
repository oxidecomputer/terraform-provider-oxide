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

resource "oxide_disk" "example" {
  project_id = data.oxide_projects.project_list.projects.0.id
  description       = "a test disk"
  name              = "myattacheddisk1"
  size              = 1073741824
  disk_source       = { blank = 512 }
}

resource "oxide_disk" "example2" {
  project_id = data.oxide_projects.project_list.projects.0.id
  description       = "a test disk"
  name              = "myattacheddisk2"
  size              = 1073741824
  disk_source       = { blank = 512 }
}

resource "oxide_instance" "example3" {
  project_id = data.oxide_projects.project_list.projects.0.id
  description       = "a test instance"
  name              = "myinstance2"
  host_name         = "myhost"
  memory            = 1073741824
  ncpus             = 1
  attach_to_disks   = ["myattacheddisk1", "myattacheddisk2"]
}