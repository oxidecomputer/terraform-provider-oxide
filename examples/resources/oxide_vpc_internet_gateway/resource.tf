resource "oxide_vpc_internet_gateway" "example" {
  vpc_id      = data.oxide_vpc.default.id
  description = "a sample VPC internet gateway"
  name        = "myinternetgateway"
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
