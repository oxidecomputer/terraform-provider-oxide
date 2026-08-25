resource "oxide_vpc_internet_gateway_ip_pool_attachment" "example" {
  name        = "my-ip-pool-attachment"
  description = "an IP pool attachment"
  gateway_id  = "f5660a9f-962e-4c00-a6dc-638256ae1d4e"
  ip_pool_id  = "fbec3335-9178-486f-90d2-315d9098ca6f"
}
