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

resource "oxide_disk" "sample_disk" {
  project_id  = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description = "a test disk"
  name        = "disk-test-1"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_instance" "sample_instance" {
  project_id      = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description     = "a test instance"
  name            = "int-test-1"
  host_name       = "myhost"
  memory          = 1073741824
  ncpus           = 1
  start_on_create = false
}

resource "oxide_instance_disk_attachment" "sample_attachment" {
  disk_id     = oxide_disk.sample_disk.id
  instance_id = oxide_instance.sample_instance.id
}