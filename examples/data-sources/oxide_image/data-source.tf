data "oxide_image" "example" {
  project_name = "my-project"
  name         = "my-image"
  timeouts = {
    read = "1m"
  }
}
