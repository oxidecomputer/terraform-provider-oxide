resource "oxide_project" "example" {
  description = "a test project"
  name        = "myproject"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
    update = "2m"
  }
}
