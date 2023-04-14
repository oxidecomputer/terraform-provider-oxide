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

data "oxide_images" "example" {
  project_id = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
}