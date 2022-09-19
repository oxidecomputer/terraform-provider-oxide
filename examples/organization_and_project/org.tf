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

resource "oxide_organization" "example" {
  description       = "a test org"
  name              = "anorg"
}

resource "oxide_project" "example2" {
  description       = "a test project"
  name              = "aproject"
  organization_name = oxide_organization.example.name
}
