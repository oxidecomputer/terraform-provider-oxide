resource "oxide_vpc_router_route" "example" {
  vpc_router_id = "c1dee930-a8e4-11ed-afa1-0242ac120002"
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
