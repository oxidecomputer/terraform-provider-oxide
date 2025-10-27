resource "oxide_vpc_firewall_rules" "example" {
  vpc_id = oxide_vpc.example.id

  rules = provider::oxide::to_vpc_firewall_rules_map(jsonencode([
    {
      name        = "allow-https"
      description = "Allow HTTPS."
      action      = "allow"
      direction   = "inbound"
      priority    = 50
      status      = "enabled"
      filters = {
        hosts = [
          {
            type  = "vpc"
            value = oxide_vpc.example.name
          }
        ]
        ports = [443]
        protocols = [
          { type = "tcp" },
          { type = "udp" },
        ]
      },
      targets = [
        {
          type  = "subnet"
          value = "default"
        }
      ]
    },
  ]))

  # Converted rules map:
  #   rules = {
  #     allow-https = {
  #       description = "Allow HTTPS."
  #       action      = "allow"
  #       direction   = "inbound"
  #       priority    = 50
  #       status      = "enabled"
  #       filters = {
  #         hosts = [
  #           {
  #             type  = "vpc"
  #             value = oxide_vpc.example.name
  #           }
  #         ]
  #         ports = [443]
  #         protocols = [
  #           { type = "tcp" },
  #           { type = "udp" },
  #         ]
  #       },
  #       targets = [
  #         {
  #           type  = "subnet"
  #           value = "default"
  #         }
  #       ]
  #     },
  #   }
}
