resource "oxide_anti_affinity_group" "example" {
  project_id  = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a test anti-affinity group"
  name        = "my-anti-affinty-group"
  policy      = "allow"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}
