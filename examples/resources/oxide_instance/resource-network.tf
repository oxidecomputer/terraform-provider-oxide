resource "oxide_instance" "example" {
  project_id       = data.oxide_project.my_project.id
  description      = "Example instance."
  name             = "myinstance"
  hostname         = "myhostname"
  memory           = 10737418240
  ncpus            = 1
  disk_attachments = [data.oxide_disk.my_disk.id]

  network_interfaces = [
    {
      subnet_id   = data.oxide_vpc_subnet.default.id
      vpc_id      = data.oxide_vpc.default.id
      description = "Example network interface."
      name        = "mynic"

      ip_config = {
        v4 = {
          ip = "172.30.0.6"
        }

        v6 = {
          ip = "auto"
        }
      }
    },
  ]

  timeouts = {
    read   = "1m"
    create = "3m"
    delete = "2m"
  }
}

# Prerequisites for the example.
data "oxide_project" "my_project" {
  name = "my-project"
}

data "oxide_disk" "my_disk" {
  project_name = "my-project"
  name         = "my-disk"
}

data "oxide_vpc" "default" {
  project_name = "my-project"
  name         = "default"
}

data "oxide_vpc_subnet" "default" {
  project_name = "my-project"
  vpc_name     = "default"
  name         = "default"
}
