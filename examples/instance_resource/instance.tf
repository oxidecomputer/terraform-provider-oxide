terraform {
  required_version = ">= 1.0"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = "0.2.0"
    }
  }
}

provider "oxide" {}

data "oxide_project" "example" {
  name = "{YOUR-PROJECT-NAME}"
}

data "oxide_vpc_subnet" "example" {
  project_name = data.oxide_project.example.name
  vpc_name     = "default"
  name         = "default"
}

resource "oxide_disk" "example" {
  project_id  = data.oxide_project.example.id
  description = "a test disk"
  name        = "my-disk"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_ssh_key" "example" {
  name        = "example"
  description = "Example SSH key."
  public_key  = "ssh-ed25519 {MY_PUBLIC_KEY}"
}

resource "oxide_instance" "test" {
  project_id       = data.oxide_project.example.id
  description      = "a test instance"
  name             = "my-instance"
  host_name        = "my-host"
  memory           = 1073741824
  ncpus            = 1
  start_on_create  = true
  disk_attachments = [oxide_disk.example.id]
  ssh_public_keys  = [oxide_ssh_key.example.id]
  external_ips      = [
    {
      type = "ephemeral"
    }
  ]
  network_interfaces = [
    {
      subnet_id   = data.oxide_vpc_subnet.example.id
      vpc_id      = data.oxide_vpc_subnet.example.vpc_id
      description = "a sample nic"
      name        = "mynic"
    }
  ]
}
