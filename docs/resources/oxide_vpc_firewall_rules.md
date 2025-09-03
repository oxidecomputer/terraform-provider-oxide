---
page_title: "oxide_vpc_firewall_rules Resource - terraform-provider-oxide"
---

# oxide_vpc_firewall_rules (Resource)

This resource manages VPC firewall rules.

!> Firewall rules defined by this resource are considered exhaustive and will
overwrite any other firewall rules for the VPC once applied.

!> Setting the `rules` attribute to `[]` will delete all firewall rules for the
VPC which may cause undesired network traffic. Please double check the firewall
rules when updating this resource.

## Example Usage

### Basic Example

```hcl
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
```

### ICMP Example

```hcl
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
```

## Schema

### Required

- `vpc_id` (String) ID of the VPC that will have the firewall rules applied to.
- `rules` (Set) Associated firewall rules. Set to `[]` to delete all firewall rules. (see [below for nested schema](#nestedatt--rules))

### Optional

- `timeouts` (Attribute, Optional) (see [below for nested schema](#nestedatt--timeouts))

### Read-Only

- `id` (String) Unique, immutable, system-controlled identifier of the firewall rules. Specific only to Terraform.
- `time_created` (String) Timestamp of when the VPC firewall rules were created.
- `time_modified` (String) Timestamp of when the VPC firewall rules were last modified.

<a id="nestedatt--rules"></a>

### Nested Schema for `rules`

Required:

- `action` (String) Whether traffic matching the rule should be allowed or dropped. Possible values are: allow or deny.
- `description` (String) Description for the VPC firewall rule.
- `direction` (String) Whether this rule is for incoming or outgoing traffic. Possible values are: inbound or outbound.
- `filters` (Single Nested Attribute) Reductions on the scope of the rule. (see [below for nested schema](#nestedatt--filters))
- `name` (String) Name of the firewall rule.
- `priority` (Number) The relative priority of this rule.
- `status` (String) Whether this rule is in effect. Possible values are: enabled or disabled.
- `targets` (Set) Sets of instances that the rule applies to. (see [below for nested schema](#nestedatt--targets))

<a id="nestedatt--filters"></a>

### Nested Schema for `filters`

Optional:

- `hosts` (Set) If present, the sources (if incoming) or destinations (if outgoing) this rule applies to. (see [below for nested schema](#nestedatt--hosts))
- `protocols` (Set) If present, the networking protocols this rule applies to. (see [below for nested schema](#nestedatt--protocols))
- `ports` (Array of Strings) If present, the destination ports this rule applies to. Can be a mix of single ports (e.g., `"443"`) and port ranges (e.g., `"30000-32768"`).

<a id="nestedatt--hosts"></a>

### Nested Schema for `hosts`

Required:

- `type` (String) Used to filter traffic on the basis of its source or destination host. Possible values: vpc, subnet, instance, ip and ip_net.
- `value` (String) Depending on the type, it will be one of the following:
	- For type vpc: Name of the VPC
	- For type subnet: Name of the VPC subnet
	- For type instance: Name of the instance
	- For type ip: IP address
	- For type ip_net: IPv4 or IPv6 subnet

<a id="nestedatt--protocols"></a>

### Nested Schema for `protocols`

Required:

- `type` (String) The protocol type. Must be one of `tcp`, `udp`, or `icmp`.

Optional:

- `icmp_type` (Number) ICMP type (e.g., 0 for Echo Reply). Only valid when `type` is `icmp`.
- `icmp_code` (String) ICMP code (e.g., 0) or range (e.g., 1-3). Omit to filter all traffic of the specified `icmp_type`. Only valid when type is `icmp` and `icmp_type` is provided.

<a id="nestedatt--targets"></a>

### Nested Schema for `targets`

Required:

- `type` (String) The rule applies to a single or all instances of this type, or specific IPs. Possible values: vpc, subnet, instance, ip, ip_net.
- `value` (String) Depending on the type, it will be one of the following:
	- For type vpc: Name of the VPC
	- For type subnet: Name of the VPC subnet
	- For type instance: Name of the instance
	- For type ip: IP address
	- For type ip_net: IPv4 or IPv6 subnet

### Nested Schema for `timeouts`

Optional:

- `create` (String, Default `10m`)
- `delete` (String, Default `10m`)
- `read` (String, Default `10m`)
- `update` (String, Default `10m`)
