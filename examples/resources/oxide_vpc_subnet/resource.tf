resource "oxide_vpc_subnet" "example" {
  vpc_id      = data.oxide_vpc.default.id
  description = "a sample vpc subnet"
  name        = "mysubnet"
  ipv4_block  = "172.30.4.0/22"
  ipv6_block  = cidrsubnet(data.oxide_vpc.default.ipv6_prefix, 16, 1)
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}

# Prerequisites for the example.
data "oxide_vpc" "default" {
  project_name = "my-project"
  name         = "default"
}
