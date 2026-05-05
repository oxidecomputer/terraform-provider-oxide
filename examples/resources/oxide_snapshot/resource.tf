resource "oxide_snapshot" "example2" {
  project_id  = data.oxide_project.my_project.id
  description = "a test snapshot"
  name        = "mysnapshot"
  disk_id     = data.oxide_disk.my_disk.id
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}

# Prerequisites for the example.
data "oxide_project" "my_project" {
  name = "my-project"
}

data "oxide_disk" "my_disk" {
  project_name = "my-project"
  name         = "my-disk"
}
