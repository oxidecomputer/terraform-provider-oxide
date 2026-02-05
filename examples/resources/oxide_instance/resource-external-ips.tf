resource "oxide_instance" "example" {
  project_id       = "c1dee930-a8e4-11ed-afa1-0242ac120002"
  description      = "Example instance."
  name             = "myinstance"
  hostname         = "myhostname"
  memory           = 10737418240
  ncpus            = 1
  disk_attachments = ["611bb17d-6883-45be-b3aa-8a186fdeafe8"]
  start_on_create  = false

  external_ips = {
    ephemeral = [
      {
        ip_version = "v4"
      },
      {
        pool_id = "f6f65759-2510-45f5-a3b8-e5090ac56993"
      },
    ]

    floating = [
      {
        id = "43b30584-1580-446d-a213-cad9e298df44",
      }
    ]
  }

  network_interfaces = [
    {
      subnet_id   = "066cab1b-c550-4aea-8a80-8422fd3bfc40"
      vpc_id      = "9b9f9be1-96bf-44ad-864a-0dedae3b3999"
      description = "Example network interface."
      name        = "mynic"

      ip_config = {
        v4 = {
          ip = "auto"
        }
      }
    },
  ]
}
