resource "oxide_anti_affinity_group" "example" {
  project_id  = data.oxide_project.my_project.id
  description = "a test anti-affinity group"
  name        = "my-anti-affinity-group"
  policy      = "allow"
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
