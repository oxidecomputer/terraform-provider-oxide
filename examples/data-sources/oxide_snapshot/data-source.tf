data "oxide_snapshot" "example" {
  project_name = "my-project"
  name         = "my-snapshot"
  timeouts = {
    read = "1m"
  }
}
