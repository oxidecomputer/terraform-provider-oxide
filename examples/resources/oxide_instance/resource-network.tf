resource "oxide_instance" "example" {
  project_id       = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description      = "Example instance."
  name             = "myinstance"
  hostname         = "myhostname"
  memory           = 10737418240
  ncpus            = 1
  disk_attachments = ["611bb17d-6883-45be-b3aa-8a186fdeafe8"]
  network_interfaces = [
    {
      subnet_id   = "066cab1b-c550-4aea-8a80-8422fd3bfc40"
      vpc_id      = "9b9f9be1-96bf-44ad-864a-0dedae3b3999"
      description = "Example network interface."
      name        = "mynic"
    },
  ]
  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}
