resource "oxide_vpc_subnet" "example" {
  vpc_id      = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a sample vpc subnet"
  name        = "mysubnet"
  ipv4_block  = "192.168.0.0/16"
  ipv6_block  = "fdfe:f6a5:5f06:a643::/64"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}
