---
page_title: "oxide_switch_port_settings Resource - terraform-provider-oxide"
---

# oxide_switch_port_settings (Resource)

This resource manages switch port settings for configuring network switch ports
including addresses, BGP peers, interfaces, links, and routes.

!> Switch port settings defined by this resource are considered exhaustive and
will overwrite any other switch port settings upon apply.

## Example Usage

### Basic Switch Port Settings

```hcl
resource "oxide_switch_port_settings" "example" {
  name        = "example"
  description = "Switch port settings."

  port_config = {
    geometry = "qsfp28x1"
  }

  addresses = [
    {
      link_name = "phy0"
      addresses = [
        {
          address        = "192.168.1.123/24"
          address_lot_id = "38223e3a-76da-400d-a1e2-8cb4d242095a"
        },
      ]
    },
  ]

  links = [
    {
      link_name = "phy0"
      autoneg   = false
      mtu       = 1500
      speed     = "speed1_g"
      lldp = {
        enabled = false
      }
    },
  ]

  routes = [
    {
      link_name = "phy0"
      routes = [
        {
          dst = "0.0.0.0/0"
          gw  = "192.168.1.1"
        },
      ]
    },
  ]
}
```

## Schema

### Required

- `addresses` (Set of Object) Address configuration for the switch port. See [below for nested schema](#nestedatt--addresses).
- `description` (String) Human-readable description of the switch port settings.
- `links` (Set of Object) Link configuration for the switch port. See [below for nested schema](#nestedatt--links).
- `name` (String) Name of the switch port settings.
- `port_config` (Object) Physical port configuration. See [below for nested schema](#nestedatt--port_config).

### Optional

- `bgp_peers` (Set of Object) BGP peer configuration for the switch port. See [below for nested schema](#nestedatt--bgp_peers).
- `groups` (Set of String) Set of port settings group IDs to include in these settings.
- `interfaces` (Set of Object) Interface configuration for the switch port. See [below for nested schema](#nestedatt--interfaces).
- `routes` (Set of Object) Static route configuration. See [below for nested schema](#nestedatt--routes).
- `timeouts` (Attribute) Timeouts for performing API operations. See [below for nested schema](#nestedatt--timeouts).

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the switch port settings.
- `time_created` (String) Timestamp of when the switch port settings were created.
- `time_modified` (String) Timestamp of when the switch port settings were last modified.

<a id="nestedatt--addresses"></a>

### Nested Schema for `addresses`

#### Required

- `addresses` (Set of Object) Set of addresses to assign to the link. See [below for nested schema](#nestedatt--addresses--addresses).
- `link_name` (String) Name of the link for the address configuration.

<a id="nestedatt--addresses--addresses"></a>

### Nested Schema for `addresses.addresses`

#### Required

- `address` (String) IPv4 or IPv6 address, including the subnet mask.
- `address_lot_id` (String) Address lot the address is allocated from.

#### Optional

- `vlan_id` (Number) VLAN ID for the address.

<a id="nestedatt--bgp_peers"></a>

### Nested Schema for `bgp_peers`

#### Required

- `link_name` (String) Name of the link for the BGP peers configuration.
- `peers` (Set of Object) Set of BGP peers configuration to assign to the link. See [below for nested schema](#nestedatt--bgp_peers--peers).

<a id="nestedatt--bgp_peers--peers"></a>

### Nested Schema for `bgp_peers.peers`

#### Required

- `addr` (String) Address of the host to peer with.
- `allowed_export` (Object) Export policy for the peer. See [below for nested schema](#nestedatt--bgp_peers--peers--allowed_export).
- `allowed_import` (Object) Import policy for the peer. See [below for nested schema](#nestedatt--bgp_peers--peers--allowed_import).
- `bgp_config` (String) Name or ID of the global BGP configuration used for establishing a session with this peer.
- `communities` (Set of Number) BGP communities to apply to this peer's routes.
- `connect_retry` (Number) Number of seconds to wait before retrying a TCP connection.
- `delay_open` (Number) Number of seconds to delay sending an open request after establishing a TCP session.
- `enforce_first_as` (Boolean) Enforce that the first autonomous system in paths received from this peer is the peer's autonomous system.
- `hold_time` (Number) Number of seconds to hold peer connections between keepalives.
- `idle_hold_time` (Number) Number of seconds to hold a peer in idle before attempting a new session.
- `interface_name` (String) Name of the interface to use for this BGP peer session.
- `keepalive` (Number) Number of seconds between sending BGP keepalive requests.

#### Optional

- `local_pref` (Number) BGP local preference value for routes received from this peer.
- `md5_auth_key` (String) MD5 authentication key for this BGP session.
- `min_ttl` (Number) Minimum acceptable TTL for BGP packets from this peer.
- `multi_exit_discriminator` (Number) Multi-exit discriminator (MED) to advertise to this peer.
- `remote_asn` (Number) Remote autonomous system number for this BGP peer.
- `vlan_id` (Number) VLAN ID for this BGP peer session.

<a id="nestedatt--bgp_peers--peers--allowed_export"></a>

### Nested Schema for `bgp_peers.peers.allowed_export`

#### Required

- `type` (String) Type of filter to apply.

#### Optional

- `value` (Set of String) IPv4 or IPv6 address to apply the filter to, including the subnet mask.

<a id="nestedatt--bgp_peers--peers--allowed_import"></a>

### Nested Schema for `bgp_peers.peers.allowed_import`

#### Required

- `type` (String) Type of filter to apply.

#### Optional

- `value` (Set of String) IPv4 or IPv6 address to apply the filter to, including the subnet mask.

<a id="nestedatt--interfaces"></a>

### Nested Schema for `interfaces`

#### Required

- `kind` (Object) The kind of interface this configuration represents. See [below for nested schema](#nestedatt--interfaces--kind).
- `link_name` (String) Name of the link this interface is associated with.

#### Optional

- `v6_enabled` (Boolean) Enable IPv6 on this interface.

<a id="nestedatt--interfaces--kind"></a>

### Nested Schema for `interfaces.kind`

#### Required

- `type` (String) Type of the interface.

#### Optional

- `vid` (Number) VLAN ID for the interfaces.

<a id="nestedatt--links"></a>

### Nested Schema for `links`

#### Required

- `autoneg` (Boolean) Whether to enable auto-negotiation for this link.
- `link_name` (String) Name of the link.
- `lldp` (Object) Link Layer Discovery Protocol (LLDP) configuration. See [below for nested schema](#nestedatt--links--lldp).
- `mtu` (Number) Maximum Transmission Unit (MTU) for this link.
- `speed` (String) Link speed.

#### Optional

- `fec` (String) Forward error correction (FEC) type.
- `tx_eq` (Object) Transceiver equalization settings. See [below for nested schema](#nestedatt--links--tx_eq).

<a id="nestedatt--links--lldp"></a>

### Nested Schema for `links.lldp`

#### Required

- `enabled` (Boolean) Whether to enable LLDP on this link.

#### Optional

- `chassis_id` (String) LLDP chassis ID.
- `link_description` (String) LLDP link description.
- `link_name` (String) LLDP link name.
- `management_ip` (String) LLDP management IP address.
- `system_description` (String) LLDP system description.
- `system_name` (String) LLDP system name.

<a id="nestedatt--links--tx_eq"></a>

### Nested Schema for `links.tx_eq`

#### Optional

- `main` (Number) Main tap equalization value.
- `post1` (Number) Post-cursor tap1 equalization value.
- `post2` (Number) Post-cursor tap2 equalization value.
- `pre1` (Number) Pre-cursor tap1 equalization value.
- `pre2` (Number) Pre-cursor tap2 equalization value.

<a id="nestedatt--port_config"></a>

### Nested Schema for `port_config`

#### Required

- `geometry` (String) Port geometry.

<a id="nestedatt--routes"></a>

### Nested Schema for `routes`

#### Required

- `link_name` (String) Name of the link for these routes.
- `routes` (Set of Object) Set of static routes for this link. See [below for nested schema](#nestedatt--routes--routes).

<a id="nestedatt--routes--routes"></a>

### Nested Schema for `routes.routes`

#### Required

- `dst` (String) Destination network in CIDR notation.
- `gw` (String) Gateway IP address for this route.

#### Optional

- `rib_priority` (Number) Routing Information Base (RIB) priority for this route.
- `vid` (Number) VLAN ID for this route.

<a id="nestedatt--timeouts"></a>

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)
