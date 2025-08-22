resource "oxide_snapshot" "example2" {
  project_id  = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description = "a test snapshot"
  name        = "mysnapshot"
  disk_id     = "49118786-ca55-49b1-ae9a-e03f7ce41d8c"
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}
