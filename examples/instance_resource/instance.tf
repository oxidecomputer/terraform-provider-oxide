terraform {
  required_version = ">= 1.0"

  required_providers {
    oxide = {
      source  = "oxidecomputer/oxide"
      version = "0.4.0-dev"
    }
  }
}

provider "oxide" {}

data "oxide_project" "example" {
  name = "{MY_PROJECT_NAME}"
}

resource "oxide_vpc" "example" {
  project_id  = data.oxide_project.example.id
  description = "a test vpc"
  name        = "myvpc"
  dns_name    = "my-vpc-dns"
}

resource "oxide_vpc_firewall_rules" "example" {
  vpc_id = oxide_vpc.example.id
  rules = [
    {
      action      = "allow"
      description = "allow inbound TCP connections on ports 22 and 80 from anywhere"
      name        = "allow-ssh-http"
      direction   = "inbound"
      priority    = 65534
      status      = "enabled"
      filters = {
        ports     = ["22", "80"]
        protocols = ["TCP"]
      },
      targets = [
        {
          type  = "vpc"
          value = oxide_vpc.example.name
        }
      ]
    },
    {
      action      = "allow"
      description = "allow inbound traffic to all instances within the VPC if originated within the VPC"
      name        = "allow-internal-inbound"
      direction   = "inbound"
      priority    = 65534
      status      = "enabled"
      filters = {
        hosts = [
          {
            type  = "vpc"
            value = oxide_vpc.example.name
          }
        ]
      },
      targets = [
        {
          type  = "vpc"
          value = oxide_vpc.example.name
        }
      ]
    },
    {
      action      = "allow"
      description = "allow inbound ICMP traffic from anywhere"
      name        = "allow-icmp"
      direction   = "inbound"
      priority    = 65534
      status      = "enabled"
      filters = {
        protocols = ["ICMP"]
      },
      targets = [
        {
          type  = "vpc"
          value = oxide_vpc.example.name
        }
      ]
    }
  ]
}

data "oxide_vpc_subnet" "example" {
  project_name = data.oxide_project.example.name
  vpc_name     = oxide_vpc.example.name
  name         = "default"
}

resource "oxide_disk" "example" {
  project_id      = data.oxide_project.example.id
  description     = "a test disk"
  name            = "my-disk"
  size            = 21474836480
  source_image_id = "{MY_IMAGE_ID}"
}

resource "oxide_ssh_key" "example" {
  name        = "example"
  description = "Example SSH key."
  public_key  = "ssh-ed25519 {MY_PUBLIC_KEY}"
}

resource "oxide_instance" "test" {
  project_id       = data.oxide_project.example.id
  description      = "a test instance"
  name             = "my-instance"
  host_name        = "my-host"
  memory           = 21474836480
  ncpus            = 1
  start_on_create  = true
  disk_attachments = [oxide_disk.example.id]
  ssh_public_keys  = [oxide_ssh_key.example.id]
  external_ips      = [
    {
      type = "ephemeral"
    }
  ]
  network_interfaces = [
    {
      subnet_id   = data.oxide_vpc_subnet.example.id
      vpc_id      = data.oxide_vpc_subnet.example.vpc_id
      description = "a sample nic"
      name        = "mynic"
    }
  ]
  user_data = filebase64("./init.sh")
}
