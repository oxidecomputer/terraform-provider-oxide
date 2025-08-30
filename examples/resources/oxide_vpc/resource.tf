resource "oxide_vpc" "example" {
  project_id  = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a test vpc"
  name        = "myvpc"
  dns_name    = "my-vpc-dns"
  ipv6_prefix = "fd1e:4947:d4a1::/48"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}
