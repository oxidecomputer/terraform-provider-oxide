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

data "oxide_project" "example" {
  name = "{YOUR-PROJECT-NAME}"
}

data "oxide_images" "image_list" {}

resource "oxide_disk" "web_disk_1" {
  project_id      = data.oxide_project.example.id
  description     = "Disk for a web instance"
  name            = "web-disk-1"
  size            = var.ten_gib
  source_image_id = data.oxide_images.image_list.images.0.id
}

resource "oxide_disk" "web_disk_2" {
  project_id      = data.oxide_project.example.id
  description     = "Disk for a web instance"
  name            = "web-disk-2"
  size            = var.ten_gib
  source_image_id = data.oxide_images.image_list.images.0.id
}

resource "oxide_disk" "web_disk_3" {
  project_id      = data.oxide_project.example.id
  description     = "Disk for a web instance"
  name            = "web-disk-3"
  size            = var.ten_gib
  source_image_id = data.oxide_images.image_list.images.0.id
}

resource "oxide_disk" "db_disk_1" {
  project_id      = data.oxide_project.example.id
  description     = "Disk for a DB instance"
  name            = "db-disk-1"
  size            = var.twenty_gib
  source_image_id = data.oxide_images.image_list.images.0.id
}

resource "oxide_disk" "db_disk_2" {
  project_id      = data.oxide_project.example.id
  description     = "Disk for a DB instance"
  name            = "db-disk-2"
  size            = var.twenty_gib
  source_image_id = data.oxide_images.image_list.images.0.id
}

resource "oxide_instance" "web_instance_1" {
  project_id        = data.oxide_project.example.id
  description       = "Web instance"
  name              = "web-instance-1"
  host_name         = "web-instance-1"
  memory            = var.one_gib
  ncpus             = 2
  disk_attachments  = [oxide_disk.web_disk_1.id]
}

resource "oxide_instance" "web_instance_2" {
  project_id        = data.oxide_project.example.id
  description       = "Web instance"
  name              = "web-instance-2"
  host_name         = "web-instance-2"
  memory            = var.one_gib
  ncpus             = 2
  disk_attachments  = [oxide_disk.web_disk_2.id]
}

resource "oxide_instance" "web_instance_3" {
  project_id        = data.oxide_project.example.id
  description       = "Web instance"
  name              = "web-instance-3"
  host_name         = "web-instance-3"
  memory            = var.one_gib
  ncpus             = 2
  disk_attachments  = [oxide_disk.web_disk_3.id]
}

resource "oxide_instance" "db_instance_1" {
  project_id        = data.oxide_project.example.id
  description       = "Web instance"
  name              = "db-instance-1"
  host_name         = "db-instance-1"
  memory            = var.two_gib
  ncpus             = 4
  disk_attachments  = [oxide_disk.db_disk_1.id]
}

resource "oxide_instance" "db_instance_2" {
  project_id        = data.oxide_project.example.id
  description       = "Web instance"
  name              = "db-instance-2"
  host_name         = "db-instance-2"
  memory            = var.two_gib
  ncpus             = 4
  disk_attachments  = [oxide_disk.db_disk_2.id]
}
