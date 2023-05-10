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

data "oxide_images" "image_list" {
  project_id = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
}

resource "oxide_disk" "web_disk_1" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "Disk for a web instance"
  name              = "web-disk-1"
  size              = var.ten_gib
  disk_source       = { global_image = data.oxide_images.image_list.images.0.id }
}

resource "oxide_disk" "web_disk_2" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "Disk for a web instance"
  name              = "web-disk-2"
  size              = var.ten_gib
  disk_source       = { global_image = data.oxide_images.image_list.images.0.id }
}

resource "oxide_disk" "web_disk_3" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "Disk for a web instance"
  name              = "web-disk-3"
  size              = var.ten_gib
  disk_source       = { global_image = data.oxide_images.image_list.images.0.id }
}

resource "oxide_disk" "db_disk_1" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "Disk for a DB instance"
  name              = "db-disk-1"
  size              = var.twenty_gib
  disk_source       = { global_image = data.oxide_images.image_list.images.0.id }
}

resource "oxide_disk" "db_disk_2" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "Disk for a DB instance"
  name              = "db-disk-2"
  size              = var.twenty_gib
  disk_source       = { global_image = data.oxide_images.image_list.images.0.id }
}

resource "oxide_instance" "web_instance_1" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "Web instance"
  name              = "web-instance-1"
  host_name         = "web-instance-1"
  memory            = var.one_gib
  ncpus             = 2
  attach_to_disks   = [oxide_disk.web_disk_1.name]
}

resource "oxide_instance_disk_attachment" "web_instance_attach_1" {
  disk_id     = oxide_disk.web_disk_1.id
  instance_id = oxide_instance.web_instance_1.id
}

resource "oxide_instance" "web_instance_2" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "Web instance"
  name              = "web-instance-2"
  host_name         = "web-instance-2"
  memory            = var.one_gib
  ncpus             = 2
  attach_to_disks   = [oxide_disk.web_disk_2.name]
}

resource "oxide_instance_disk_attachment" "web_instance_attach_2" {
  disk_id     = oxide_disk.web_disk_2.id
  instance_id = oxide_instance.web_instance_2.id
}

resource "oxide_instance" "web_instance_3" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "Web instance"
  name              = "web-instance-3"
  host_name         = "web-instance-3"
  memory            = var.one_gib
  ncpus             = 2
  start_on_create   = false
}

resource "oxide_instance_disk_attachment" "web_instance_attach_3" {
  disk_id     = oxide_disk.web_disk_3.id
  instance_id = oxide_instance.web_instance_3.id
}

resource "oxide_instance" "db_instance_1" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "Web instance"
  name              = "db-instance-1"
  host_name         = "db-instance-1"
  memory            = var.two_gib
  ncpus             = 4
  start_on_create   = false
}

resource "oxide_instance_disk_attachment" "db_instance_attach_1" {
  disk_id     = oxide_disk.db_disk_1.id
  instance_id = oxide_instance.db_instance_1.id
}

resource "oxide_instance" "db_instance_2" {
  project_id        = element(tolist(data.oxide_projects.project_list.projects[*].id), 0)
  description       = "Web instance"
  name              = "db-instance-2"
  host_name         = "db-instance-2"
  memory            = var.two_gib
  ncpus             = 4
  start_on_create   = false
}

resource "oxide_instance_disk_attachment" "db_instance_attach_2" {
  disk_id     = oxide_disk.db_disk_2.id
  instance_id = oxide_instance.db_instance_2.id
}