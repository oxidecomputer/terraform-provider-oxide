data "oxide_anti_affinity_group" "example" {
  project_name = "my-project"
  name         = "my-group"
  timeouts = {
    read = "1m"
  }
}
