resource "oxide_instance" "example" {
  project_id           = data.oxide_project.my_project.id
  description          = "Example instance."
  name                 = "myinstance"
  hostname             = "myhostname"
  memory               = 10737418240
  ncpus                = 1
  anti_affinity_groups = [data.oxide_anti_affinity_group.my_group.id]
  disk_attachments     = [data.oxide_disk.my_disk.id]
  ssh_public_keys      = [data.oxide_ssh_key.my_key.id]
  user_data            = base64encode("#!/bin/sh\necho hello from cloud-init\n")
}

# Prerequisites for the example.
data "oxide_project" "my_project" {
  name = "my-project"
}

data "oxide_disk" "my_disk" {
  project_name = "my-project"
  name         = "my-disk"
}

data "oxide_anti_affinity_group" "my_group" {
  project_name = "my-project"
  name         = "my-group"
}

data "oxide_ssh_key" "my_key" {
  name = "my-key"
}
