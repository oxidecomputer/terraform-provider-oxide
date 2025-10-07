data "oxide_vpc_router_route" "example" {
  project_name    = "my-project"
  name            = "default-v4"
  vpc_name        = "default"
  vpc_router_name = "system"
  timeouts = {
    read = "1m"
  }
}
