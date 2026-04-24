resource "oxide_instance" "example" {
  project_id       = data.oxide_project.my_project.id
  description      = "Example instance."
  name             = "myinstance"
  hostname         = "myhostname"
  memory           = 10737418240
  ncpus            = 1
  disk_attachments = [data.oxide_disk.my_disk.id]
}

# Prerequisites for the example.
data "oxide_project" "my_project" {
  name = "my-project"
}

data "oxide_disk" "my_disk" {
  project_name = "my-project"
  name         = "my-disk"
}
