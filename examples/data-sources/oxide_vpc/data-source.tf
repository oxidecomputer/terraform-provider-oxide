data "oxide_vpc" "example" {
  project_name = "my-project"
  name         = "default"
  timeouts = {
    read = "1m"
  }
}
