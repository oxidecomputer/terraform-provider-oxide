resource "oxide_vpc" "example" {
  project_id  = data.oxide_project.my_project.id
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

# Prerequisites for the example.
data "oxide_project" "my_project" {
  name = "my-project"
}
