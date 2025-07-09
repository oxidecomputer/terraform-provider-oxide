# 0.12.0 (2025/Jul/09)

### Bug fixes

- **`oxide_instance`:** Fixed the `inconsistent result after apply` error when applying a subsequent plan where the `external_ips` attribute contained an ephemeral IP with a non-empty ID. [#460](https://github.com/oxidecomputer/terraform-provider-oxide/pull/460)

### List of commits

- [7af9fa6](https://github.com/oxidecomputer/terraform-provider-oxide/commit/7af9fa6) resource(oxide_instance): fix inconsistent result after apply (#460)
- [f9b42d1](https://github.com/oxidecomputer/terraform-provider-oxide/commit/f9b42d1) misc: update release checklist (#458)
- [f148e51](https://github.com/oxidecomputer/terraform-provider-oxide/commit/f148e51) changelog: add missing date
- [d1a4d6d](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d1a4d6d) misc: bump to version v0.12.0 (#457)

# 0.11.0 (2025/Jun/26)

### Breaking changes

- **Removed computed attributes from `vpc_firewall_rules`:** Removed all computed attributes from the `rules` attribute within the `vpc_firewall_rules` resource to fix an invalid plan error while maintaining in-place update of VPC firewall rules. [#456](https://github.com/oxidecomputer/terraform-provider-oxide/pull/456)

### List of commits

- [d93050b](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d93050b) vpc_firewall_rules: fix provider produced invalid plan (#456)
- [4336e1e](https://github.com/oxidecomputer/terraform-provider-oxide/commit/4336e1e) revert: vpc_firewall_rules: fix provider produced invalid plan (#455)
- [83b931f](https://github.com/oxidecomputer/terraform-provider-oxide/commit/83b931f) misc: format json
- [788bf8c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/788bf8c) vpc_firewall_rules: fix provider produced invalid plan
- [26cfef7](https://github.com/oxidecomputer/terraform-provider-oxide/commit/26cfef7) Bump github.com/hashicorp/terraform-plugin-testing from 1.13.1 to 1.13.2 (#452)
- [b925ae6](https://github.com/oxidecomputer/terraform-provider-oxide/commit/b925ae6) Bump to version 0.11.0 (#451)

# 0.10.0 (2025/Jun/12)

### Breaking changes

- **Minimum Terraform version v1.11 required:** Due to the introduction of [write-only attributes](https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments) in the new `oxide_silo` resource, the minimum Terraform version is now v1.11 [#425](https://github.com/oxidecomputer/terraform-provider-oxide/pull/425).

### New features

- **New resource:** `oxide_silo` [#425](https://github.com/oxidecomputer/terraform-provider-oxide/pull/425).
- **New resource:** `oxide_vpc_router_route` [#423](https://github.com/oxidecomputer/terraform-provider-oxide/pull/423).
- **New data resource:** `oxide_vpc_router_route` [#423](https://github.com/oxidecomputer/terraform-provider-oxide/pull/423).

### Enhancements

- **VPC firewall rules resource:** In place updates are now supported [#432](https://github.com/oxidecomputer/terraform-provider-oxide/pull/432)
- **`oxide_instance` attribute validation:** Added validation to the `ssh_public_keys` attribute on the `oxide_instance` resource. [#443](https://github.com/oxidecomputer/terraform-provider-oxide/pull/443)

### List of commits

- [c004b55](https://github.com/oxidecomputer/terraform-provider-oxide/commit/c004b55) Ignore 404 when removing instance membership from anti-affinity group (#449)
- [bb71de5](https://github.com/oxidecomputer/terraform-provider-oxide/commit/bb71de5) Update docs to reference 0.10.0 (#448)
- [2674283](https://github.com/oxidecomputer/terraform-provider-oxide/commit/2674283) Update to Go SDK v0.5.0 (#447)
- [a10d6cf](https://github.com/oxidecomputer/terraform-provider-oxide/commit/a10d6cf) Bump github.com/cloudflare/circl from 1.6.0 to 1.6.1 (#445)
- [a5e3113](https://github.com/oxidecomputer/terraform-provider-oxide/commit/a5e3113) This filters the boot disk out of attachments (#444)
- [f803d8b](https://github.com/oxidecomputer/terraform-provider-oxide/commit/f803d8b) oxide_instance: add validation to ssh_public_keys attribute (#443)
- [7478329](https://github.com/oxidecomputer/terraform-provider-oxide/commit/7478329) In-place update for VPC firewall rules (#432)
- [3f15cea](https://github.com/oxidecomputer/terraform-provider-oxide/commit/3f15cea) Bump github.com/hashicorp/terraform-plugin-go from 0.27.0 to 0.28.0 (#439)
- [2eacede](https://github.com/oxidecomputer/terraform-provider-oxide/commit/2eacede) Bump github.com/hashicorp/terraform-plugin-testing from 1.13.0 to 1.13.1 (#440)
- [8860cdc](https://github.com/oxidecomputer/terraform-provider-oxide/commit/8860cdc) oxide_instance: allow in-place `external_ips` modification (#381)
- [ac7cc52](https://github.com/oxidecomputer/terraform-provider-oxide/commit/ac7cc52) Bump github.com/hashicorp/terraform-plugin-framework-validators from 0.17.0 to 0.18.0 (#433)
- [4744222](https://github.com/oxidecomputer/terraform-provider-oxide/commit/4744222) Bump github.com/hashicorp/terraform-plugin-testing from 1.12.0 to 1.13.0 (#437)
- [154598c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/154598c) Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.36.1 to 2.37.0 (#434)
- [b646f11](https://github.com/oxidecomputer/terraform-provider-oxide/commit/b646f11) Bump github.com/hashicorp/terraform-plugin-framework from 1.14.1 to 1.15.0 (#435)
- [e5644d0](https://github.com/oxidecomputer/terraform-provider-oxide/commit/e5644d0) Feature silo creation (#425)
- [9e5664d](https://github.com/oxidecomputer/terraform-provider-oxide/commit/9e5664d) Router route resource and datasource (#423)
- [9f0e000](https://github.com/oxidecomputer/terraform-provider-oxide/commit/9f0e000) version: bump to 0.10.0 (#429)

# 0.9.0 (2025/Apr/23)

### New features

- **New resource:** `oxide_floating_ip` [#427](https://github.com/oxidecomputer/terraform-provider-oxide/pull/427).
- **New data resource:** `oxide_floating_ip` [#427](https://github.com/oxidecomputer/terraform-provider-oxide/pull/427).

### List of commits

- [2bd3b33](https://github.com/oxidecomputer/terraform-provider-oxide/commit/2bd3b33) oxide_floating_ip: initial resource and data source (#427)
- [c54c243](https://github.com/oxidecomputer/terraform-provider-oxide/commit/c54c243) Bump golang.org/x/net from 0.37.0 to 0.38.0 (#424)
- [440abcb](https://github.com/oxidecomputer/terraform-provider-oxide/commit/440abcb) Version bump to 0.9.0 (#422)

# 0.8.0 (2025/Apr/15)

### New features

- **New instance argument:** It is now possible to manage anti-affinity group assigments on instance resources. [#414](https://github.com/oxidecomputer/terraform-provider-oxide/pull/414).
- **New resource:** `oxide_anti_affinity_group` [#415](https://github.com/oxidecomputer/terraform-provider-oxide/pull/415).
- **New data source:** `oxide_anti_affinity_group` [#415](https://github.com/oxidecomputer/terraform-provider-oxide/pull/415).

### List of commits

- [9a8fcec](https://github.com/oxidecomputer/terraform-provider-oxide/commit/9a8fcec) Update SDK to v0.4.0 and doc fixes (#420)
- [663403c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/663403c) Anti-affinity groups on instance create (#414)
- [021b0f8](https://github.com/oxidecomputer/terraform-provider-oxide/commit/021b0f8) New anti-affinity group resource and datasource (#415)
- [c55ad49](https://github.com/oxidecomputer/terraform-provider-oxide/commit/c55ad49) Bump goreleaser/goreleaser-action from 6.2.1 to 6.3.0 (#413)
- [cb246e7](https://github.com/oxidecomputer/terraform-provider-oxide/commit/cb246e7) Bump SDK version to c8be658 (#410)
- [ce6ac4c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/ce6ac4c) Update dependencies (#409)
- [fbc831d](https://github.com/oxidecomputer/terraform-provider-oxide/commit/fbc831d) vpc_firewall_rules: update documentation (#405)
- [31d1cca](https://github.com/oxidecomputer/terraform-provider-oxide/commit/31d1cca) Update GoReleaser deprecated fields (#402)
- [0fe7a4d](https://github.com/oxidecomputer/terraform-provider-oxide/commit/0fe7a4d) Version bump to 0.8.0 (#401)
- [6712875](https://github.com/oxidecomputer/terraform-provider-oxide/commit/6712875) Bump github.com/hashicorp/terraform-plugin-framework from 1.13.0 to 1.14.1 (#398)
- [da67e68](https://github.com/oxidecomputer/terraform-provider-oxide/commit/da67e68) Bump github.com/hashicorp/terraform-plugin-framework-validators from 0.16.0 to 0.17.0 (#397)
- [a3fdc71](https://github.com/oxidecomputer/terraform-provider-oxide/commit/a3fdc71) Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.36.0 to 2.36.1 (#396)

# 0.7.0 (2025/Feb/25)

### New features

- **Profile attribute in provider block:** Allows authentication via an authenticated `profile` in `credentials.toml` [#383](https://github.com/oxidecomputer/terraform-provider-oxide/pull/383).
- **New resource:** `oxide_vpc_router` [#388](https://github.com/oxidecomputer/terraform-provider-oxide/pull/388).
- **New datasource:** `oxide_vpc_router` [#388](https://github.com/oxidecomputer/terraform-provider-oxide/pull/388).
- **New resource:** `oxide_vpc_internet_gateway` [#389](https://github.com/oxidecomputer/terraform-provider-oxide/pull/389).
- **New datasource:** `oxide_vpc_internet_gateway` [#389](https://github.com/oxidecomputer/terraform-provider-oxide/pull/389).

### List of commits

- [360c215](https://github.com/oxidecomputer/terraform-provider-oxide/commit/360c215) Update SDK to v0.3.0 (#399)
- [e76a859](https://github.com/oxidecomputer/terraform-provider-oxide/commit/e76a859) Update docs to new version 0.7.0 (#395)
- [36f7bf0](https://github.com/oxidecomputer/terraform-provider-oxide/commit/36f7bf0) Bump goreleaser/goreleaser-action from 6.1.0 to 6.2.1 (#394)
- [44b6bdd](https://github.com/oxidecomputer/terraform-provider-oxide/commit/44b6bdd) docs: update oxide_instance resource (#393)
- [a85899a](https://github.com/oxidecomputer/terraform-provider-oxide/commit/a85899a) New `oxide_vpc_internet_gateway` resource and datasource (#389)
- [f471988](https://github.com/oxidecomputer/terraform-provider-oxide/commit/f471988) Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.35.0 to 2.36.0 (#391)
- [8540a1f](https://github.com/oxidecomputer/terraform-provider-oxide/commit/8540a1f) New oxide_vpc_router resource and datasource (#388)
- [accc74b](https://github.com/oxidecomputer/terraform-provider-oxide/commit/accc74b) Improve contributing docs (#387)
- [fc247e8](https://github.com/oxidecomputer/terraform-provider-oxide/commit/fc247e8) Bump github.com/hashicorp/terraform-plugin-go from 0.25.0 to 0.26.0 (#386)
- [c7d43e8](https://github.com/oxidecomputer/terraform-provider-oxide/commit/c7d43e8) Bump github.com/hashicorp/terraform-plugin-framework-timeouts from 0.4.1 to 0.5.0 (#384)
- [631a15c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/631a15c) Adding profile lookup for credentials.toml (#383)
- [9d23c27](https://github.com/oxidecomputer/terraform-provider-oxide/commit/9d23c27) Bump github.com/oxidecomputer/oxide.go from 0.1.0-beta9.0.20241125050113-eb153ea4db8c to 0.2.0 (#382)
- [90e035d](https://github.com/oxidecomputer/terraform-provider-oxide/commit/90e035d) Bump golang.org/x/net from 0.28.0 to 0.33.0 (#380)
- [e393a06](https://github.com/oxidecomputer/terraform-provider-oxide/commit/e393a06) Bump golang.org/x/crypto from 0.29.0 to 0.31.0 (#379)
- [f140866](https://github.com/oxidecomputer/terraform-provider-oxide/commit/f140866) Bump github.com/hashicorp/terraform-plugin-framework-validators from 0.15.0 to 0.16.0 (#372)
- [532c3f1](https://github.com/oxidecomputer/terraform-provider-oxide/commit/532c3f1) release: bump version (#375)
- [dac5730](https://github.com/oxidecomputer/terraform-provider-oxide/commit/dac5730) release: update changelog (#374)

# 0.6.0 (2025/Jan/06)

### Enhancements

- **Instance resource:** Updates to `Memory` and `NCPUs` no longer require resource replace [#370](https://github.com/oxidecomputer/terraform-provider-oxide/pull/370).

### List of commits

- [b19273d](https://github.com/oxidecomputer/terraform-provider-oxide/commit/b19273d) release: bump version (#373)
- [69c8d9a](https://github.com/oxidecomputer/terraform-provider-oxide/commit/69c8d9a) Update to Go SDK eb153ea (#370)
- [9c66bad](https://github.com/oxidecomputer/terraform-provider-oxide/commit/9c66bad) Bump github.com/hashicorp/terraform-plugin-testing from 1.10.0 to 1.11.0 (#369)
- [a8c9da0](https://github.com/oxidecomputer/terraform-provider-oxide/commit/a8c9da0) Bump github.com/stretchr/testify from 1.9.0 to 1.10.0 (#368)
- [e92be74](https://github.com/oxidecomputer/terraform-provider-oxide/commit/e92be74) codeowners: update owners (#367)
- [731ad55](https://github.com/oxidecomputer/terraform-provider-oxide/commit/731ad55) Bump goreleaser/goreleaser-action from 6.0.0 to 6.1.0 (#366)
- [432eb1b](https://github.com/oxidecomputer/terraform-provider-oxide/commit/432eb1b) Bump github.com/hashicorp/terraform-plugin-framework-validators from 0.14.0 to 0.15.0 (#364)
- [7b3ab82](https://github.com/oxidecomputer/terraform-provider-oxide/commit/7b3ab82) Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.34.0 to 2.35.0 (#365)
- [922ad1c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/922ad1c) Bump github.com/hashicorp/terraform-plugin-framework from 1.12.0 to 1.13.0 (#363)
- [7fb9c74](https://github.com/oxidecomputer/terraform-provider-oxide/commit/7fb9c74) Bump github.com/hashicorp/terraform-plugin-framework-validators from 0.13.0 to 0.14.0 (#360)
- [bd0df07](https://github.com/oxidecomputer/terraform-provider-oxide/commit/bd0df07) Upgrade to Go 1.23 and tf plugins (#351)
- [98a5e76](https://github.com/oxidecomputer/terraform-provider-oxide/commit/98a5e76) Fix example on returning instance external ip (#359)
- [26dcc35](https://github.com/oxidecomputer/terraform-provider-oxide/commit/26dcc35) Update to 0.6 for development (#358)

# 0.5.0 (2024/Oct/3)

### New features

- **New resource:** `oxide_ip_pool_silo_link` [#345](https://github.com/oxidecomputer/terraform-provider-oxide/pull/345).
- **New datasource:** `oxide_silo` [#347](https://github.com/oxidecomputer/terraform-provider-oxide/pull/347).
- **New instance attribute:** It is now possible to specify a boot disk by setting `boot_disk_id` [#352](https://github.com/oxidecomputer/terraform-provider-oxide/pull/352).

### Enhancements

- **Instance resource:** Disk attachments no longer require resource replacement [#352](https://github.com/oxidecomputer/terraform-provider-oxide/pull/352).

### List of commits

- [5f90026](https://github.com/oxidecomputer/terraform-provider-oxide/commit/5f90026) Docs version bump (#355)
- [d85888a](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d85888a) Refine instance update (#353)
- [10ab870](https://github.com/oxidecomputer/terraform-provider-oxide/commit/10ab870) `oxide_instance` boot disk implementation (#352)
- [b323c37](https://github.com/oxidecomputer/terraform-provider-oxide/commit/b323c37) Silo datasource (#347)
- [3e4303a](https://github.com/oxidecomputer/terraform-provider-oxide/commit/3e4303a) Update to Go SDK 7b8deef (#348)
- [8e5e448](https://github.com/oxidecomputer/terraform-provider-oxide/commit/8e5e448) `oxide_ip_pool_silo_link` resource (#345)
- [800bdb8](https://github.com/oxidecomputer/terraform-provider-oxide/commit/800bdb8) Bump goreleaser/goreleaser-action from 5.1.0 to 6.0.0 (#324)
- [7067777](https://github.com/oxidecomputer/terraform-provider-oxide/commit/7067777) Update goreleaser version to 2 (#343)
- [f5932a7](https://github.com/oxidecomputer/terraform-provider-oxide/commit/f5932a7) Bump version to 0.5.0 (#342)

# 0.4.0 (2024/Sep/3)

### Breaking changes

- **`oxide_vpc_firewall_rules` resource:** Setting an empty array for `filters.hosts`, `filters.ports` and `filters.protocols` is no longer supported. To omit they must be unset. [#322](https://github.com/oxidecomputer/terraform-provider-oxide/pull/322)

### New features

- **New resource:** `oxide_ip_pool` [#337](https://github.com/oxidecomputer/terraform-provider-oxide/pull/337)

### List of commits

- [b7004cb](https://github.com/oxidecomputer/terraform-provider-oxide/commit/b7004cb) Bump version and update go SDK to 0.1.0-beta8 (#340)
- [f926330](https://github.com/oxidecomputer/terraform-provider-oxide/commit/f926330) IP pools resource (#337)
- [7479a12](https://github.com/oxidecomputer/terraform-provider-oxide/commit/7479a12) Update oxide SDK to 3ece27 (#336)
- [84308fa](https://github.com/oxidecomputer/terraform-provider-oxide/commit/84308fa) Update README to refer to credentials.toml (#331)
- [8222c50](https://github.com/oxidecomputer/terraform-provider-oxide/commit/8222c50) Bump github.com/hashicorp/terraform-plugin-framework from 1.10.0 to 1.11.0 (#333)
- [151a08c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/151a08c) Bump github.com/hashicorp/terraform-plugin-testing from 1.9.0 to 1.10.0 (#332)
- [2aa7d20](https://github.com/oxidecomputer/terraform-provider-oxide/commit/2aa7d20) Bump go SDK to v0.1.0-beta7 (#330)
- [48b1a91](https://github.com/oxidecomputer/terraform-provider-oxide/commit/48b1a91) Bump github.com/hashicorp/terraform-plugin-framework-validators from 0.12.0 to 0.13.0 (#329)
- [9eefa18](https://github.com/oxidecomputer/terraform-provider-oxide/commit/9eefa18) Bump github.com/hashicorp/terraform-plugin-testing from 1.8.0 to 1.9.0 (#328)
- [0f47840](https://github.com/oxidecomputer/terraform-provider-oxide/commit/0f47840) Bump github.com/hashicorp/terraform-plugin-framework from 1.9.0 to 1.10.0 (#327)
- [76b76f7](https://github.com/oxidecomputer/terraform-provider-oxide/commit/76b76f7) Update Go SDK to oxide.go@06dd780 (#326)
- [5509a22](https://github.com/oxidecomputer/terraform-provider-oxide/commit/5509a22) [examples] Improve the instance example (#325)
- [65667bc](https://github.com/oxidecomputer/terraform-provider-oxide/commit/65667bc) Bump github.com/hashicorp/terraform-plugin-framework from 1.8.0 to 1.9.0 (#323)
- [d3448b4](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d3448b4) Do not support empty arrays for filters (#322)
- [334451a](https://github.com/oxidecomputer/terraform-provider-oxide/commit/334451a) Bump github.com/hashicorp/terraform-plugin-testing from 1.7.0 to 1.8.0 (#320)
- [051a1b2](https://github.com/oxidecomputer/terraform-provider-oxide/commit/051a1b2) Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.33.0 to 2.34.0 (#319)
- [9648298](https://github.com/oxidecomputer/terraform-provider-oxide/commit/9648298) Bump goreleaser/goreleaser-action from 5.0.0 to 5.1.0 (#318)
- [a159c1c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/a159c1c) Bump github.com/oxidecomputer/oxide.go from 0.1.0-beta5 to 0.1.0-beta6 (#317)
- [346115f](https://github.com/oxidecomputer/terraform-provider-oxide/commit/346115f) Bump github.com/hashicorp/terraform-plugin-go from 0.22.2 to 0.23.0 (#316)
- [4b75916](https://github.com/oxidecomputer/terraform-provider-oxide/commit/4b75916) Update SDK to v0.1.0-beta5 (#315)
- [a4c3b2d](https://github.com/oxidecomputer/terraform-provider-oxide/commit/a4c3b2d) Bump github.com/hashicorp/terraform-plugin-framework from 1.7.0 to 1.8.0 (#314)
- [2c87da0](https://github.com/oxidecomputer/terraform-provider-oxide/commit/2c87da0) Bump github.com/hashicorp/terraform-plugin-go from 0.22.1 to 0.22.2 (#313)
- [7cf197c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/7cf197c) Bump golang.org/x/net from 0.21.0 to 0.23.0 (#312)
- [6d46708](https://github.com/oxidecomputer/terraform-provider-oxide/commit/6d46708) Bump github.com/hashicorp/terraform-plugin-framework from 1.6.1 to 1.7.0 (#298)
- [5ac843f](https://github.com/oxidecomputer/terraform-provider-oxide/commit/5ac843f) Fix acceptance testing action (#303)
- [d4eb035](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d4eb035) Bump version to 0.4.0 (#302)

# 0.3.0 (2024/Apr/3)

### Enhancements

- **Documentation:** Clarification about retrieving silo images [#288](https://github.com/oxidecomputer/terraform-provider-oxide/pull/288)

### List of commits

- [be49e35](https://github.com/oxidecomputer/terraform-provider-oxide/commit/be49e35) Update Go SDK to v0.1.0-beta4 (#300)
- [afcbd16](https://github.com/oxidecomputer/terraform-provider-oxide/commit/afcbd16) Update to Go SDK f488d8e875 (#299)
- [d7a016c](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d7a016c) Update versions in docs to reflect upcoming release (#297)
- [765a156](https://github.com/oxidecomputer/terraform-provider-oxide/commit/765a156) Bump github.com/hashicorp/terraform-plugin-go from 0.22.0 to 0.22.1 (#294)
- [a849db1](https://github.com/oxidecomputer/terraform-provider-oxide/commit/a849db1) Bump github.com/hashicorp/terraform-plugin-testing from 1.6.0 to 1.7.0 (#293)
- [5830ffd](https://github.com/oxidecomputer/terraform-provider-oxide/commit/5830ffd) Bump github.com/hashicorp/terraform-plugin-framework from 1.6.0 to 1.6.1 (#292)
- [5f77675](https://github.com/oxidecomputer/terraform-provider-oxide/commit/5f77675) Update terraform framework and plugin dependencies (#291)
- [c52063a](https://github.com/oxidecomputer/terraform-provider-oxide/commit/c52063a) Bump github.com/stretchr/testify from 1.8.4 to 1.9.0 (#289)
- [786ff95](https://github.com/oxidecomputer/terraform-provider-oxide/commit/786ff95) Include documentation about retrieving silo images (#288)
- [2e31c95](https://github.com/oxidecomputer/terraform-provider-oxide/commit/2e31c95) Small internal linter fix to instance resource (#287)
- [ceaa457](https://github.com/oxidecomputer/terraform-provider-oxide/commit/ceaa457) Fix ssh_public_keys docs (#286)
- [643ac53](https://github.com/oxidecomputer/terraform-provider-oxide/commit/643ac53) Run acceptance tests as part of CI in releases (#285)
- [ef5502f](https://github.com/oxidecomputer/terraform-provider-oxide/commit/ef5502f) Update to Go SDK 043c873 (#283)
- [fa365ab](https://github.com/oxidecomputer/terraform-provider-oxide/commit/fa365ab) Bump version and add tag make target (#278)
- [cde8700](https://github.com/oxidecomputer/terraform-provider-oxide/commit/cde8700) Fix changelog (#277)

# 0.2.0 (2024/Feb/13)

### Breaking changes

- **`oxide_instance` resource:** The `name` field in `external_ips` for the `oxide_instance` resource has been replaced with `id`. This ensures correctness, and helps avoid unintenional drift if the IP pool's name were to change outside the scope of terraform. [#263](https://github.com/oxidecomputer/terraform-provider-oxide/pull/263)
- **`oxide_instance` resource:** A new optional `ssh_public_keys` field has been added to the `oxide_instance` resource. It is an allowlist of IDs of the saved SSH public keys to be transferred to the instance via cloud-init during instance creation. Saved SSH keys will no longer be automatically added to the instances [#269](https://github.com/oxidecomputer/terraform-provider-oxide/pull/269)

### New features

- **New data source:** `oxide_ip_pool` [#263](https://github.com/oxidecomputer/terraform-provider-oxide/pull/263)

### List of commits

- [79d8a04](https://github.com/oxidecomputer/terraform-provider-oxide/commit/79d8a04) Update to Go SDK v0.1.0-beta3 (#274)
- [b28a6b7](https://github.com/oxidecomputer/terraform-provider-oxide/commit/b28a6b7) Update to oxide.go 428a544 (#273)
- [35a5757](https://github.com/oxidecomputer/terraform-provider-oxide/commit/35a5757) Select SSH keys to be added to instance on create (#269)
- [8382a9e](https://github.com/oxidecomputer/terraform-provider-oxide/commit/8382a9e) Smalll tweaks to release template (#272)
- [7c2edfa](https://github.com/oxidecomputer/terraform-provider-oxide/commit/7c2edfa) Update examples (#271)
- [532ac0a](https://github.com/oxidecomputer/terraform-provider-oxide/commit/532ac0a) Refactor acceptance tests (#270)
- [2fc402d](https://github.com/oxidecomputer/terraform-provider-oxide/commit/2fc402d) Bump github.com/google/uuid from 1.5.0 to 1.6.0 (#268)
- [70d0b60](https://github.com/oxidecomputer/terraform-provider-oxide/commit/70d0b60) Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.31.0 to 2.32.0 (#266)
- [8f26ee5](https://github.com/oxidecomputer/terraform-provider-oxide/commit/8f26ee5) Small template fixes (#265)
- [8c11e5d](https://github.com/oxidecomputer/terraform-provider-oxide/commit/8c11e5d) New IP pool datasrouce and update IP pools to match API (#263)
- [c2d5108](https://github.com/oxidecomputer/terraform-provider-oxide/commit/c2d5108) Bump actions/cache from 3 to 4 (#262)
- [86c03ed](https://github.com/oxidecomputer/terraform-provider-oxide/commit/86c03ed) Update Go SDK to 172bbb155e83 (#261)
- [ebb36e6](https://github.com/oxidecomputer/terraform-provider-oxide/commit/ebb36e6) Bump github.com/hashicorp/terraform-plugin-framework from 1.4.2 to 1.5.0 (#260)
- [d085704](https://github.com/oxidecomputer/terraform-provider-oxide/commit/d085704) Rename changelog file to the correct version (#259)
- [bd05f22](https://github.com/oxidecomputer/terraform-provider-oxide/commit/bd05f22) Bump github.com/cloudflare/circl from 1.3.3 to 1.3.7 (#258)
- [85109d8](https://github.com/oxidecomputer/terraform-provider-oxide/commit/85109d8) [docs] Fix to match the new image resource schema (#253)
- [befd412](https://github.com/oxidecomputer/terraform-provider-oxide/commit/befd412) Update to Go 1.21 (#251)
- [e086706](https://github.com/oxidecomputer/terraform-provider-oxide/commit/e086706) Bump golang.org/x/crypto from 0.16.0 to 0.17.0 (#249)
- [61452b5](https://github.com/oxidecomputer/terraform-provider-oxide/commit/61452b5) Bump github.com/google/uuid from 1.4.0 to 1.5.0 (#248)
- [bceaa69](https://github.com/oxidecomputer/terraform-provider-oxide/commit/bceaa69) Bump github.com/hashicorp/terraform-plugin-sdk/v2 from 2.30.0 to 2.31.0 (#246)
- [1480d63](https://github.com/oxidecomputer/terraform-provider-oxide/commit/1480d63) [github] Feature request issue template (#245)
- [798e679](https://github.com/oxidecomputer/terraform-provider-oxide/commit/798e679) [github] Fix issue template yaml (#244)
- [56f131d](https://github.com/oxidecomputer/terraform-provider-oxide/commit/56f131d) [github] Rename bug template (#243)
- [8e4b88b](https://github.com/oxidecomputer/terraform-provider-oxide/commit/8e4b88b) [github] Use the correct file extension for template (#242)
- [8279f52](https://github.com/oxidecomputer/terraform-provider-oxide/commit/8279f52) [github] Issue templates (#241)
- [ef79a39](https://github.com/oxidecomputer/terraform-provider-oxide/commit/ef79a39) Update files to upcoming version (#239)
- [fdedb88](https://github.com/oxidecomputer/terraform-provider-oxide/commit/fdedb88) Add changelog (#238)

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

