data "oxide_vpc_subnet" "example" {
  project_name = "my-project"
  name         = "default"
  vpc_name     = "default"
  timeouts = {
    read = "1m"
  }
}
