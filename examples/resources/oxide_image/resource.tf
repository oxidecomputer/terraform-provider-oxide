resource "oxide_image" "example2" {
  project_id         = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description        = "a test image"
  name               = "myimage2"
  source_snapshot_id = "eb65d5cb-d8c5-4eae-bcf3-a0e89a633042"
  os                 = "ubuntu"
  version            = "20.04"
  timeouts = {
    read   = "1m"
    create = "3m"
  }
}
