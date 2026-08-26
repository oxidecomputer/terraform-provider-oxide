resource "oxide_vpc_internet_gateway_ip_pool_attachment" "example" {
  name        = "my-ip-pool-attachment"
  description = "an IP pool attachment"
  gateway_id  = data.oxide_vpc_internet_gateway.default.id
  ip_pool_id  = data.oxide_ip_pool.default.id
}

data "oxide_vpc_internet_gateway" "default" {
  project_name = "my-project"
  vpc_name     = "default"
  name         = "default"
}

data "oxide_ip_pool" "default" {
  name = "default"
}
