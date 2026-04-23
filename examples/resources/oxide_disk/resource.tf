resource "oxide_disk" "example" {
  project_id  = data.oxide_project.my_project.id
  description = "a test disk"
  name        = "mydisk"
  size        = 1073741824
  block_size  = 512
}

resource "oxide_disk" "example2" {
  project_id      = data.oxide_project.my_project.id
  description     = "a test disk"
  name            = "mydisk2"
  size            = 1073741824
  source_image_id = "49118786-ca55-49b1-ae9a-e03f7ce41d8c"
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
