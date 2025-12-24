resource "oxide_instance" "example" {
  project_id       = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description      = "Example instance."
  name             = "myinstance"
  hostname         = "myhostname"
  memory           = 10737418240
  ncpus            = 1
  disk_attachments = ["611bb17d-6883-45be-b3aa-8a186fdeafe8"]
}
