data "oxide_vpc_internet_gateway" "example" {
  project_name = "my-project"
  name         = "default"
  vpc_name     = "default"
  timeouts = {
    read = "1m"
  }
}
