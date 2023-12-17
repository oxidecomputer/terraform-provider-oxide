# 0.1.0 (2023/Dec/18)

### Breaking changes

- **Image resource:** Image creation via URL is no longer supported [#228](https://github.com/oxidecomputer/terraform-provider-oxide/pull/228)
- **Instance resource:** Support floating IPs as external IP addresses provided to the instance. The `ip_pool_name` attribute within the `external_ips` block has been modified to `name`. [#230](https://github.com/oxidecomputer/terraform-provider-oxide/pull/230)
- **Image data sources:** All image data sources no longer retrieve an image source URL [#234](https://github.com/oxidecomputer/terraform-provider-oxide/pull/234)

### New features

- **New resource:** `oxide_ssh_key` [#211](https://github.com/oxidecomputer/terraform-provider-oxide/pull/211)
- **New data source:** `oxide_ssh_key` [#211](https://github.com/oxidecomputer/terraform-provider-oxide/pull/211)
- **New resource:** `oxide_vpc_firewall_rules` [#220](https://github.com/oxidecomputer/terraform-provider-oxide/pull/220)

### Enhancements

- **Documentation:** Various documentation clarifications [#185](https://github.com/oxidecomputer/terraform-provider-oxide/pull/185), [#218](https://github.com/oxidecomputer/terraform-provider-oxide/pull/218)
- **Image resource:** Deletes are now supported [#228](https://github.com/oxidecomputer/terraform-provider-oxide/pull/228)
- **VPC and project resources:** System created default VPCs and subnets are now removed automatically when deleting a VPC or project [#229](https://github.com/oxidecomputer/terraform-provider-oxide/pull/229)

### List of commits

- [cad68a7](https://github.com/oxidecomputer/terraform-provider-oxide/commit/cad68a7) Update to SDK v0.1.0-beta2 (#237)
- [a4ec337](https://github.com/oxidecomputer/terraform-provider-oxide/commit/a4ec337) Update repo docs to reflect new 0.1.0 version (#236)
- [5b14882](https://github.com/oxidecomputer/terraform-provider-oxide/commit/5b14882) Update to latest SDK commit (#234)
- [b3d1294](https://github.com/oxidecomputer/terraform-provider-oxide/commit/b3d1294) Update README.md (#233)
- [b26fd28](https://github.com/oxidecomputer/terraform-provider-oxide/commit/b26fd28) Bump actions/setup-go from 4 to 5 (#232)
- [116d2cb](https://github.com/oxidecomputer/terraform-provider-oxide/commit/116d2cb) Bump all terraform modules (#231)
- [d587225](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d587225) [instance resource] Support floating IPs (#230)
- [3eb287c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/3eb287c) Automatically remove system created VPC and subnet when deleting projects and VPCs (#229)
- [62c5824](https://github.com/oxidecomputer/terraform-provider-oxide/commit/62c5824) [image resource] Remove ability to create via URL and enable deletes (#228)
- [9fca9ab](https://github.com/oxidecomputer/terraform-provider-oxide/commit/9fca9ab) Bump release version to 0.1.0 (#225)
- [5a5e3e7](https://github.com/oxidecomputer/terraform-provider-oxide/commit/5a5e3e7) New resource for VPC firewall rules (#220)
- [d3f196b](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d3f196b) Doc fixes and changelog entries (#218)
- [1fb7ec4](https://github.com/oxidecomputer/terraform-provider-oxide/commit/1fb7ec4) Various acceptance test fixes (#215)
- [29c22e7](https://github.com/oxidecomputer/terraform-provider-oxide/commit/29c22e7) Update to oxide SDK e20dc58 (#214)
- [5ca47f2](https://github.com/oxidecomputer/terraform-provider-oxide/commit/5ca47f2) Implement `oxide_ssh_key` resource and data source (#211)
- [0ad0c75](https://github.com/oxidecomputer/terraform-provider-oxide/commit/0ad0c75) Bump github.com/google/uuid from 1.3.1 to 1.4.0 (#212)
- [c343a0b](https://github.com/oxidecomputer/terraform-provider-oxide/commit/c343a0b) Bump golang.org/x/net from 0.11.0 to 0.17.0 (#207)
- [ea0ab26](https://github.com/oxidecomputer/terraform-provider-oxide/commit/ea0ab26) Bump google.golang.org/grpc from 1.56.1 to 1.56.3 (#209)
- [8ade868](https://github.com/oxidecomputer/terraform-provider-oxide/commit/8ade868) Bump goreleaser/goreleaser-action from 4.6.0 to 5.0.0 (#205)
- [9b7535c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/9b7535c) Bump crazy-max/ghaction-import-gpg from 5 to 6 (#204)
- [d65ba24](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d65ba24) Bump goreleaser/goreleaser-action from 4.4.0 to 4.6.0 (#203)
- [18ba480](https://github.com/oxidecomputer/terraform-provider-oxide/commit/18ba480) Bump actions/checkout from 3 to 4 (#197)
- [2ecc32a](https://github.com/oxidecomputer/terraform-provider-oxide/commit/2ecc32a) Bump github.com/hashicorp/terraform-plugin-testing from 1.4.0 to 1.5.1 (#198)
- [144a902](https://github.com/oxidecomputer/terraform-provider-oxide/commit/144a902) Bump github.com/hashicorp/terraform-plugin-framework-validators (#199)
- [0d9d00f](https://github.com/oxidecomputer/terraform-provider-oxide/commit/0d9d00f) Update Go SDK to v0.1.0-beta1 (#196)
- [58d3cbb](https://github.com/oxidecomputer/terraform-provider-oxide/commit/58d3cbb) Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.27.0 to 2.28.0 (#195)
- [564c53e](https://github.com/oxidecomputer/terraform-provider-oxide/commit/564c53e) Bump github.com/hashicorp/terraform-plugin-framework from 1.3.4 to 1.3.5 (#194)
- [bb031d1](https://github.com/oxidecomputer/terraform-provider-oxide/commit/bb031d1) Bump github.com/google/uuid from 1.3.0 to 1.3.1 (#193)
- [fc19990](https://github.com/oxidecomputer/terraform-provider-oxide/commit/fc19990) Bump goreleaser/goreleaser-action from 4.3.0 to 4.4.0 (#192)
- [01aaad4](https://github.com/oxidecomputer/terraform-provider-oxide/commit/01aaad4) Bump github.com/hashicorp/terraform-plugin-framework-validators (#191)
- [4006aab](https://github.com/oxidecomputer/terraform-provider-oxide/commit/4006aab) Bump github.com/hashicorp/terraform-plugin-framework from 1.3.3 to 1.3.4 (#190)
- [5586f8c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/5586f8c) Bump github.com/hashicorp/terraform-plugin-framework from 1.3.2 to 1.3.3 (#189)
- [0788136](https://github.com/oxidecomputer/terraform-provider-oxide/commit/0788136) Bump github.com/hashicorp/terraform-plugin-testing from 1.3.0 to 1.4.0 (#188)
- [d0d81c9](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d0d81c9) Post release version housekeeping (#186)
- [c67786e](https://github.com/oxidecomputer/terraform-provider-oxide/commit/c67786e) Small doc fix (#185)

