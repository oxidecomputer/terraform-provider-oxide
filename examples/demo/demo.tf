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

data "oxide_organizations" "org_list" {}

data "oxide_projects" "project_list" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
}

data "oxide_global_images" "image_list" {}

resource "oxide_disk" "web_disk_1" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
  project_name      = data.oxide_projects.project_list.projects.0.name
  description       = "Disk for a web instance"
  name              = "web-disk-1"
  size              = 107374182400
  disk_source       = { global_image = data.oxide_global_images.image_list.global_images.0.id }
}

resource "oxide_disk" "web_disk_2" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
  project_name      = data.oxide_projects.project_list.projects.0.name
  description       = "Disk for a web instance"
  name              = "web-disk-2"
  size              = 107374182400
  disk_source       = { global_image = data.oxide_global_images.image_list.global_images.0.id }
}

resource "oxide_disk" "web_disk_3" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
  project_name      = data.oxide_projects.project_list.projects.0.name
  description       = "Disk for a web instance"
  name              = "web-disk-3"
  size              = 107374182400
  disk_source       = { global_image = data.oxide_global_images.image_list.global_images.0.id }
}

resource "oxide_disk" "db_disk_1" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
  project_name      = data.oxide_projects.project_list.projects.0.name
  description       = "Disk for a DB instance"
  name              = "db-disk-1"
  size              = 536870912000
  disk_source       = { global_image = data.oxide_global_images.image_list.global_images.0.id }
}

resource "oxide_disk" "db_disk_2" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
  project_name      = data.oxide_projects.project_list.projects.0.name
  description       = "Disk for a DB instance"
  name              = "db-disk-2"
  size              = 536870912000
  disk_source       = { global_image = data.oxide_global_images.image_list.global_images.0.id }
}

resource "oxide_instance" "web_instance_1" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
  project_name      = data.oxide_projects.project_list.projects.0.name
  description       = "Web instance"
  name              = "web-instance-1"
  host_name         = "web-instance-1"
  memory            = 8589934592
  ncpus             = 2
  attach_to_disks   = [oxide_disk.web_disk_1.name]
  network_interface {
    description = "Network interface for web instance"
    name        = "web-interface-1"
    subnet_name = "default"
    vpc_name    = "default"
  }
}

resource "oxide_instance" "web_instance_2" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
  project_name      = data.oxide_projects.project_list.projects.0.name
  description       = "Web instance"
  name              = "web-instance-2"
  host_name         = "web-instance-2"
  memory            = 8589934592
  ncpus             = 2
  attach_to_disks   = [oxide_disk.web_disk_2.name]
  network_interface {
    description = "Network interface for web instance"
    name        = "web-interface-2"
    subnet_name = "default"
    vpc_name    = "default"
  }
}

resource "oxide_instance" "web_instance_3" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
  project_name      = data.oxide_projects.project_list.projects.0.name
  description       = "Web instance"
  name              = "web-instance-3"
  host_name         = "web-instance-3"
  memory            = 8589934592
  ncpus             = 2
  attach_to_disks   = [oxide_disk.web_disk_3.name]
  network_interface {
    description = "Network interface for web instance"
    name        = "web-interface-3"
    subnet_name = "default"
    vpc_name    = "default"
  }
}

resource "oxide_instance" "db_instance_1" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
  project_name      = data.oxide_projects.project_list.projects.0.name
  description       = "Web instance"
  name              = "db-instance-1"
  host_name         = "db-instance-1"
  memory            = 34359738368
  ncpus             = 8
  attach_to_disks   = [oxide_disk.db_disk_1.name]
  network_interface {
    description = "Network interface for DB instance"
    name        = "db-interface-1"
    subnet_name = "default"
    vpc_name    = "default"
  }
}

resource "oxide_instance" "db_instance_2" {
  organization_name = data.oxide_organizations.org_list.organizations.0.name
  project_name      = data.oxide_projects.project_list.projects.0.name
  description       = "Web instance"
  name              = "db-instance-2"
  host_name         = "db-instance-2"
  memory            = 34359738368
  ncpus             = 8
  attach_to_disks   = [oxide_disk.db_disk_2.name]
  network_interface {
    description = "Network interface for DB instance"
    name        = "db-interface-2"
    subnet_name = "default"
    vpc_name    = "default"
  }
}