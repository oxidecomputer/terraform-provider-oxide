resource "oxide_switch_port_settings" "example" {
  name        = "example"
  description = "Example switch port settings."

  port_config = {
    geometry = "qsfp28x1"
  }

  addresses = [
    {
      link_name = "phy0"
      addresses = [
        {
          address        = "0.0.0.0/32"
          address_lot_id = "bc9d4ae4-8403-41db-8d08-b2380d6b898f"
        },
      ]
    },
  ]

  bgp_peers = [
    {
      link_name = "phy0"
      peers = [
        {
          address = "192.168.1.1"
          allowed_export = {
            type = "no_filtering"
          }
          allowed_import = {
            type = "no_filtering"
          }
          bgp_config       = "0a30b48d-f726-40dd-87e0-7174c7bee84a"
          communities      = []
          connect_retry    = 10
          delay_open       = 10
          enforce_first_as = false
          hold_time        = 10
          idle_hold_time   = 10
          interface_name   = "phy0"
          keepalive        = 10
        }
      ]
    }
  ]

  links = [
    {
      link_name = "phy0"
      autoneg   = false
      mtu       = 1500
      speed     = "speed1_g"
      lldp = {
        enabled = true
      }
    },
    {
      link_name = "phy1"
      autoneg   = false
      mtu       = 1500
      speed     = "speed10_g"
      lldp = {
        enabled = true
      }
    },
  ]

  routes = [
    {
      link_name = "phy0"
      routes = [
        {
          dst = "0.0.0.0/0"
          gw  = "0.0.0.0"
        },
      ]
    },
  ]
}
