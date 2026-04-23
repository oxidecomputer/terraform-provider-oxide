data "oxide_project" "example" {
  name = "my-project"
  timeouts = {
    read = "1m"
  }
}
