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

resource "oxide_project" "setup_project" {
  description       = "a test project"
  name              = "myproj"
}

resource "oxide_ip_pool" "ip_pool_ranges" {
  description = "a test IP pool"
  name        = "mypool"
  ranges = [
    {
      first_address = "172.20.15.227"
      last_address  = "172.20.15.239"
    }
  ]
}

resource "oxide_global_image" "debian" {
  description  = "a debian image"
  name         = "debian"
  image_source = { url = "http://${var.catacomb}/media/cloud/debian-11-genericcloud-amd64.raw" }
  block_size   = 512
  os           = "debian"
  version      = "11"
}

resource "oxide_global_image" "ubuntu" {
  description  = "an ubuntu image"
  name         = "ubuntu"
  image_source = { url = "http://${var.catacomb}/media/cloud/focal-server-cloudimg-amd64.raw" }
  block_size   = 512
  os           = "ubuntu"
  version      = "22.04"
}

resource "oxide_global_image" "fedora" {
  description  = "a fedora image"
  name         = "fedora"
  image_source = { url = "http://${var.catacomb}/media/cloud/Fedora-Cloud-Base-35-1.2.x86_64.raw" }
  block_size   = 512
  os           = "fedora"
  version      = "35-1.2"
}

resource "oxide_global_image" "debian-nocloud" {
  description  = "a debian-nocloud image"
  name         = "debian-nocloud"
  image_source = { url = "http://${var.catacomb}/media/debian/debian-11-nocloud-amd64-20220503-998.raw" }
  block_size   = 512
  os           = "debian-nocloud"
  version      = "nocloud 11"
}

resource "oxide_global_image" "ubuntu-nocloud-iso" {
  description  = "an ubuntu nocloud iso image"
  name         = "ubuntu-nocloud-iso"
  image_source = { url = "http://${var.catacomb}/media/ubuntu/ubuntu-22.04-live-server-amd64.iso" }
  block_size   = 512
  os           = "ubuntu-iso"
  version      = "iso 22.04"
}