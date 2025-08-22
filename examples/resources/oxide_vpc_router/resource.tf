resource "oxide_vpc_router" "example" {
  vpc_id      = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a sample vpc router"
  name        = "myrouter"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}
