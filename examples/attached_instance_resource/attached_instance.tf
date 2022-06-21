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

resource "oxide_disk" "example" {
  organization_name = "corp"
  project_name      = "test"
  description       = "a test disk"
  name              = "myattacheddisk1"
  size              = 1024
  disk_source       = { blank = 512 }
}

resource "oxide_disk" "example2" {
  organization_name = "corp"
  project_name      = "test"
  description       = "a test disk"
  name              = "myattacheddisk2"
  size              = 1024
  disk_source       = { blank = 512 }
}

resource "oxide_instance" "example3" {
  organization_name = "corp"
  project_name      = "test"
  description       = "a test instance"
  name              = "myinstance2"
  host_name         = "myhost"
  memory            = 512
  ncpus             = 1
  attach_to_disks   = ["myattacheddisk1", "myattacheddisk2"]
}