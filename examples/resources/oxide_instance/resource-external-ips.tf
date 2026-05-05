resource "oxide_instance" "example" {
  project_id       = data.oxide_project.my_project.id
  description      = "Example instance."
  name             = "myinstance"
  hostname         = "myhostname"
  memory           = 10737418240
  ncpus            = 1
  disk_attachments = [data.oxide_disk.my_disk.id]
  start_on_create  = false

  external_ips = {
    ephemeral = [
      {
        ip_version = "v4"
      },
      {
        pool_id = data.oxide_ip_pool.default.id
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

# Prerequisites for the example.
data "oxide_project" "my_project" {
  name = "my-project"
}

data "oxide_ip_pool" "default" {
  name = "default"
}

data "oxide_disk" "my_disk" {
  project_name = "my-project"
  name         = "my-disk"
}
