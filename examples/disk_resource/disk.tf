terraform {
  required_version = ">= 1.11"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = "0.12.0"
    }
  }
}

provider "oxide" {}

data "oxide_images" "image_example" {}

data "oxide_project" "example" {
  name = "{YOUR-PROJECT-NAME}"
}

resource "oxide_disk" "example" {
  project_id  = data.oxide_project.example.id
  description = "a test disk"
  name        = "mydisk"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_disk" "example2" {
  project_id      = data.oxide_project.example.id
  description     = "a test disk"
  name            = "mydisk2"
  size            = 1073741824
  source_image_id = element(tolist(data.oxide_images.image_example.global_images[*].id), 0)
}
