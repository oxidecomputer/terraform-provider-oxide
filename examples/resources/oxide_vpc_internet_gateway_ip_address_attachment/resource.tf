resource "oxide_vpc_internet_gateway_ip_address_attachment" "example" {
  gateway_id  = data.oxide_vpc_internet_gateway.default.id
  address     = "198.51.100.47"
  name        = "my-address-attachment"
  description = "an IP address attached to my internet gateway"
  timeouts = {
    create = "1m"
    read   = "1m"
    delete = "1m"
  }
}

data "oxide_vpc_internet_gateway" "default" {
  project_name = "my-project"
  vpc_name     = "default"
  name         = "default"
}
