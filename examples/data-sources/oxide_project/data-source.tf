data "oxide_project" "example" {
  name = "test"
  timeouts = {
    read = "1m"
  }
}
