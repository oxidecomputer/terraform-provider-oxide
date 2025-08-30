data "oxide_vpc_router" "example" {
  project_name = "my-project"
  name         = "system"
  vpc_name     = "default"
  timeouts = {
    read = "1m"
  }
}
