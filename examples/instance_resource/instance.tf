terraform {
  required_version = ">= 1.0"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = ">= 0.1.0"
    }
  }
}

provider "oxide" {}

data "oxide_project" "example" {
  name = "terraform-test"
}

data "oxide_vpc" "example" {
  project_name = "terraform-test"
  name         = "default"
}

data "oxide_vpc_subnet" "example" {
  project_name = "terraform-test"
  vpc_name     = "default"
  name         = "default"
}

resource "oxide_disk" "example" {
  project_id  = data.oxide_project.example.id
  description = "a test disk"
  name        = "terraform-disk-test"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_instance" "test" {
  project_id       = data.oxide_project.example.id
  description      = "a test instance"
  name             = "terraform-instance-test"
  host_name        = "terraform-acc-myhost"
  memory           = 1073741824
  ncpus            = 1
  start_on_create  = false
  disk_attachments = [oxide_disk.example.id]
  network_interfaces = [
    {
      subnet_id   = data.oxide_vpc_subnet.example.id
      vpc_id      = data.oxide_vpc.example.id
      description = "a sample nic"
      name        = "mynic"
    },
  ]
}
