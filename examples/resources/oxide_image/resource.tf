resource "oxide_image" "example2" {
  project_id         = data.oxide_project.my_project.id
  description        = "a test image"
  name               = "myimage2"
  source_snapshot_id = "eb65d5cb-d8c5-4eae-bcf3-a0e89a633042"
  os                 = "ubuntu"
  version            = "20.04"
  timeouts = {
    read   = "1m"
    create = "3m"
  }
}

# Prerequisites for the example.
data "oxide_project" "my_project" {
  name = "my-project"
}
