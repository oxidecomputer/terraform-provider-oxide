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
  name              = "mydisk"
  size              = 1073741824
  disk_source       = { blank = 512 }
}

resource "oxide_disk" "example2" {
  organization_name = "corp"
  project_name      = "test"
  description       = "a test disk"
  name              = "mydisk2"
  size              = 1073741824
  disk_source       = { global_image = "611bb17d-6883-45be-b3aa-8a186fdeafe8" }
}