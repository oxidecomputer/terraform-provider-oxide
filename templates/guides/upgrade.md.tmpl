---
page_title: "Oxide Provider - Version Upgrade Guide"
description: |-
  Instructions on how to upgrade the Oxide provider to versions that contain
  deprecations and breaking changes.
---

# Oxide Provider Version Upgrade Guide

This page documents additional steps and instructions that may be necessary
when upgrading the Oxide Terraform provider to a given version.

It may be necessary to apply all version changes when upgrading multiple
versions at a time.

Refer to the
[changelog](https://github.com/oxidecomputer/terraform-provider-oxide/blob/main/CHANGELOG.md)
for a full list of changes.

## Upgrading to `0.20.0`

Release `0.20.0` contains breaking changes that require updates to Terraform
configuration files.

### Minimum Supported Oxide Version

The minimum supported Oxide version is now v20 (API Version [2026060800.0.0](https://github.com/oxidecomputer/omicron/blob/rel/v20/rc1/openapi/nexus/nexus-2026060800.0.0-f1db6e.json)).

This provider will not work with previous versions of Oxide.

### Resource `oxide_switch_port_settings`

#### Breaking change: deprecated resource

The `oxide_switch_port_settings` resource is deprecated and will be removed in
version v0.22.0 of the provider. Please remove this resource from your Terraform
state to stop managing it via Terraform. A future version of the provider will
release a new resource with the functionality provided by this resource once the
Oxide API has been updated accordingly.

#### Breaking change: state removal

The `bgp_peers` attribute is no longer tracked in Terraform state. It is
considered, but not enforced, to be write only. Remove any usage of this
attribute from your Terraform configuration.

#### Breaking change: attribute removal

The `bgp_peers[].peers[].interface_name` attribute has been removed. Remove any
usage of this attribute from your Terraform configuration.

#### Breaking change: attribute schema change

The `bgp_peers[].peers[].address` attribute has been removed in favor of
`bgp_peers[].peers[].addr`, which uses a new schema. Here's a diff showing how
to migrate your Terrafrom configuration.

```diff
--- main.tf
+++ main.tf
     {
       peers = [
         {
-          address = "192.168.1.1"
+          addr = {
+            type = "numbered"
+            ip   = "192.168.1.1"
+          }
         }
       ]
     }
```

## Upgrading to `0.19.0`

Release `0.19.0` contains breaking changes that require updates to Terraform
configuration files.

### Resource `oxide_instance`

#### Breaking change: removed deprecated attributes

The deprecated attributes `host_name`, `network_interfaces.ip_address`,
`network_intefaces.mac_address`, `network_interfaces.id`,
`network_interfaces.primary`, `network_interfaces.time_created`, and
`network_interfaces.time_modified` have been removed. Use `hostname` and the
`attached_network_interfaces` instead.

Refer to previous deprecation notes for more details on to how to update
configuration files.

### Function `to_vpc_firewall_rules_map`

#### Breaking change: removed `to_vpc_firewall_rules_map` function

The `to_vpc_firewall_rules_map` function has been removed.

Refer to previous breaking change notes of `vpc_firewall_rules` for more
details on to how to update configuration files.

## Upgrading to `0.18.0`

Release `0.18.0` contains breaking changes that require updates to Terraform
configuration files and deprecations that may require changes when upgrading to
future versions.

### Resource `oxide_instance`

#### Breaking change: changed schema of `external_ips`

The `external_ips` attribute of the `oxide_instance` resource is now an object
instead of a list. The new schema adds support for ephemeral IPv6.
`oxide_instance` resources need to be updated to match the new schema.

Before:

```terraform
resource "oxide_instance" "instance" {
  # ...
  external_ips = [
    # External IP from the default pool.
    {
      type = "ephemeral"
    },

    # External IP from a specific IP pool.
    {
      type = "ephemeral"
      id   = "4f0e69ad-66b6-41c0-b727-7b0285b0c384"
    },

    # Floating IP.
    {
      type = "floating"
      id   = "eb65d5cb-d8c5-4eae-bcf3-a0e89a633042"
    }
  ]
  # ...
}
```

After:

```terraform
resource "oxide_instance" "instance" {
  # ...
  external_ips = {
    ephemeral = [
      # External IPv4 from the default pool.
      {
        ip_version = "v4"
      },

      # External IP from a specific IP pool.
      {
        ip_version = "v4"
        pool_id    = "4f0e69ad-66b6-41c0-b727-7b0285b0c384"
      },
    ]

    floating = [
      # Floating IP.
      {
        id = "eb65d5cb-d8c5-4eae-bcf3-a0e89a633042"
      },
    ]
  }
  # ...
}
```

#### Deprecation: `network_interfaces.ip_address`

The `ip_address` attribute of `network_interfaces` elements has been deprecated
and will removed in a future release. Use `ip_config` instead.

The new `ip_config` attribute allows using IPv6 addresses.

Before:

```terraform
resource "oxide_instance" "instance" {
  # ...
  network_interfaces = [
    # Network interface with auto-assigned IP.
    {
      name        = "nic0"
      description = "nic0"
      vpc_id      = data.oxide_vpc_subnet.default.vpc_id
      subnet_id   = data.oxide_vpc_subnet.default.id
    },

    # Network interface with specific IP.
    {
      name        = "nic1"
      description = "nic1"
      vpc_id      = data.oxide_vpc_subnet.default.vpc_id
      subnet_id   = data.oxide_vpc_subnet.default.id
      ip_addres   = "172.30.0.5"
    },
  ]
  # ...
}
```

After:

```terraform
resource "oxide_instance" "instance" {
  # ...
  network_interfaces = [
    # Network interface with auto-assigned IPv4.
    {
      name        = "nic0"
      description = "nic0"
      vpc_id      = data.oxide_vpc_subnet.default.vpc_id
      subnet_id   = data.oxide_vpc_subnet.default.id

      ip_config = {
        v4 = {
          ip = "auto"
        }
      }
    },

    # Network interface with specific IPv4.
    {
      name        = "nic1"
      description = "nic1"
      vpc_id      = data.oxide_vpc_subnet.default.vpc_id
      subnet_id   = data.oxide_vpc_subnet.default.id

      ip_config = {
        v4 = {
          ip = "172.30.0.5"
        }
      }
    },
  ]
  # ...
}
```

#### Deprecation: computed attributes of `network_interfaces` elements

The computed attributes `id`, `mac_address`, `primary`, `time_created`, and
`time_modified` attributes of `network_interfaces` elements have been
deprecated and will be removed in a future release. Use the new
`attached_network_interfaces` attribute to access these values.

Network interfaces in the new `attached_network_interfaces` attribute can be
referenced by name instead of having to iterate over a set.

Before:

```terraform
resource "oxide_instance" "instance" {
  # ...
  network_interfaces = [
    {
      name        = "nic0"
      description = "nic0"
      vpc_id      = data.oxide_vpc_subnet.default.vpc_id
      subnet_id   = data.oxide_vpc_subnet.default.id
    },
  ]
  # ...
}

output "nic0" {
  value = [for nic in oxide_instance.test.network_interfaces : {
    id            = nic.id
    mac_address   = nic.mac_address
    primary       = nic.primary
    time_created  = nic.time_created
    time_modified = nic.time_modified
  }][0]
}
```

After:

```terraform
resource "oxide_instance" "instance" {
  # ...
  network_interfaces = [
    {
      name        = "nic0"
      description = "nic0"
      vpc_id      = data.oxide_vpc_subnet.default.vpc_id
      subnet_id   = data.oxide_vpc_subnet.default.id

      ip_config = {
        v4 = {
          ip = "auto"
        }
      }
    },
  ]
  # ...
}

output "nic0" {
  value = {
    id            = oxide_instance.test.attached_network_interfaces["nic0"].id
    mac_address   = oxide_instance.test.attached_network_interfaces["nic0"].mac_address
    primary       = oxide_instance.test.attached_network_interfaces["nic0"].primary
    time_created  = oxide_instance.test.attached_network_interfaces["nic0"].time_created
    time_modified = oxide_instance.test.attached_network_interfaces["nic0"].time_modified
  }
}
```

#### Deprecation: `host_name`

This `host_name` attribute has been deprecated and will be removed in a future
release. Use `hostname` instead.

Before:

```terraform
resource "oxide_instance" "instance" {
  # ...
  host_name = "my-instance"
  # ...
}
```

After:

```terraform
resource "oxide_instance" "instance" {
  # ...
  hostname = "my-instance"
  # ...
}
```

## Upgrading to `0.16.0`

Release `0.16.0` contains breaking changes that require updates to Terraform
configuration files.

### Resource `vpc_firewall_rules`

#### Breaking change: changed schema of `rules`

The `rules` attribute has been changed from a set to a hashmap for better
performance. Terraform configuration files need to be updated to use the new
schema.

1. Update the `rules` attribute from a set to a map.
2. Define the `rules` map keys as the VPC firewall rule name. Note that this
   key must then comply with the [Oxide
   API](https://docs.oxide.computer/api/vpc_firewall_rules_update) requirements
   for VPC firewall rule names.
3. Remove the `name` attribute from all entries of the `rules` map.

Before:

```terraform
resource "oxide_vpc_firewall_rules" "example" {
  vpc_id = "6556fc6a-63c0-420b-bb23-c3205410f5cc"
  rules = [
    {
      name        = "allow-https"
      action      = "allow"
      description = "Allow HTTPS."
      # ...
    }
  ]
}
```

After:

```terraform
resource "oxide_vpc_firewall_rules" "example" {
  vpc_id = "6556fc6a-63c0-420b-bb23-c3205410f5cc"
  rules = {
    allow-https = {
      action      = "allow"
      description = "Allow HTTPS."
      # ...
    }
  }
}
```

The new `provider::oxide::to_vpc_firewall_rules_map` provider function can also
be used to reduce the amount of changes necessary, but note that this function
is provided as a temporary solution. You should update your configuration files
to use the new schema as soon as possible.

```terraform
resource "oxide_vpc_firewall_rules" "example" {
  vpc_id = "6556fc6a-63c0-420b-bb23-c3205410f5cc"
  rules = provider::oxide::to_vpc_firewall_rules_map(jsonencode([
    {
      name        = "allow-https"
      action      = "allow"
      description = "Allow HTTPS."
      # ...
    }
  ]))
}
```
