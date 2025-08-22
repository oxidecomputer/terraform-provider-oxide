# Basic Example
resource "oxide_vpc_firewall_rules" "example" {
  vpc_id = "6556fc6a-63c0-420b-bb23-c3205410f5cc"
  rules = [
    {
      action      = "allow"
      description = "Allow HTTPS."
      name        = "allow-https"
      direction   = "inbound"
      priority    = 50
      status      = "enabled"
      filters = {
        hosts = [
          {
            type  = "vpc"
            value = "default"
          }
        ]
        ports     = ["443"]
        protocols = ["tcp"]
      },
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    }
  ]
}

# ICMP Example
resource "oxide_vpc_firewall_rules" "example" {
  vpc_id = "6556fc6a-63c0-420b-bb23-c3205410f5cc"
  rules = [
    {
      action      = "allow"
      description = "Allow ICMP"
      name        = "allow-icmp"
      direction   = "inbound"
      priority    = 50
      status      = "enabled"
      filters = {
        protocols = [
          # All ICMP.
          {
            type = "icmp",
          },
          # Echo Reply types.
          {
            type      = "icmp",
            icmp_type = 0
          },
          # Echo Reply types with codes 1-3.
          {
            type      = "icmp",
            icmp_type = 0
            icmp_code = "1-3"
          },
        ]
      },
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    }
  ]
}
