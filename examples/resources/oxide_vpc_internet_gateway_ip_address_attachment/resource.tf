resource "oxide_vpc_internet_gateway_ip_address_attachment" "example" {
  gateway_id  = "f5660a9f-962e-4c00-a6dc-638256ae1d4e"
  address     = "198.51.100.47"
  name        = "my-address-attachment"
  description = "an IP address attached to my internet gateway"
  timeouts = {
    create = "1m"
    read   = "1m"
    delete = "1m"
  }
}
