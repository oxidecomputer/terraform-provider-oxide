resource "oxide_vpc_router_route" "example" {
  vpc_router_id = data.oxide_vpc_router.system.id
  description   = "a sample VPC router route"
  name          = "myroute"
  destination = {
    type  = "ip_net"
    value = "::/0"
  }
  target = {
    type  = "ip"
    value = "::1"
  }
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}

# Prerequisites for the example.
data "oxide_vpc_router" "system" {
  project_name = "my-project"
  vpc_name     = "default"
  name         = "system"
}
