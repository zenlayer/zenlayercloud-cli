# Changelog

All notable changes to this project will be documented in this file.

## [1.0.21](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.21) - 2026-06-24

### Documentation

- update CHANGELOG.md for v1.0.20

### Features

- update APIs
## [1.0.20](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.20) - 2026-06-23

### Documentation

- update CHANGELOG.md for v1.0.19

### Features

- update APIs and upgrade zenlayercloud-sdk-go to v0.2.44
## [1.0.19](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.19) - 2026-06-08

### Documentation

- update CHANGELOG.md for v1.0.18

### Features

- update Apis
## [1.0.18](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.18) - 2026-06-03

### Documentation

- update zec API & add mcpgw API
- update CHANGELOG.md for v1.0.17
## [1.0.17](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.17) - 2026-05-30

### Documentation

- update CHANGELOG.md for v1.0.16

### Features

- hide requestId from table output
## [1.0.16](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.16) - 2026-05-25

### Documentation

- add describe-image-copy-progress api & update api yaml
- update CHANGELOG.md for v1.0.15
## [1.0.15](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.15) - 2026-05-25

### Documentation

- update CHANGELOG.md for v1.0.14

### Features

- add token authentication
## [1.0.14](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.14) - 2026-05-22

### Documentation

- update CHANGELOG.md for v1.0.13

### Features

- table output field sort by schema
## [1.0.13](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.13) - 2026-05-21

### Bug Fixes

- update table output format

### Documentation

- update CHANGELOG.md for v1.0.12
## [1.0.12](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.12) - 2026-05-21

### Documentation

- update CHANGELOG.md for v1.0.11

### Features

- add HA VIP api
## [1.0.11](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.11) - 2026-05-21

### Documentation

- update CHANGELOG.md for v1.0.10

### Features

- add vm & aigw api
## [1.0.10](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.10) - 2026-05-19

### Documentation

- update CHANGELOG.md for v1.0.9

### Features

- bmc LB support zvm instance backend api
## [1.0.9](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.9) - 2026-05-18

### Bug Fixes

- add cross-device fallback to Rollback and add security/checksum tests
- correctly fall back to 'dev' version when no git tag is matched
- add flag mutual exclusion, handle stdin EOF, add cobra.NoArgs to upgrade
- improve upgrade --check label and --list alignment
- add size limit to ExtractBinary to prevent large file extraction
- detect and reject downloads exceeding 100 MB in Download func
- harden updater package (timeout, size limit, tar safety, cross-device install)
- update product command description
- skills yaml format error

### Documentation

- add upgrade section to installation guide
- add upgrade command to README and add --rollback to upgrade docs
- update CHANGELOG.md for v1.0.8

### Features

- add upgrade command with --check, --list, --version, --rollback, --yes
- add internal/updater package with core upgrade logic
## [1.0.8](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.8) - 2026-05-12

### Documentation

- update CHANGELOG.md for v1.0.7

### Features

- add zbc service api
## [1.0.7](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.7) - 2026-05-12

### Documentation

- update CHANGELOG.md for v1.0.6

### Features

- update api yaml
## [1.0.6](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.6) - 2026-05-06

### Bug Fixes

- --help output format & update zeno skills content, add "handle async operations"
- api schema yaml missing item-schema for array type field

### Documentation

- update CHANGELOG.md for v1.0.5
## [1.0.5](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.5) - 2026-04-30

### Documentation

- update CHANGELOG.md for v1.0.4

### Features

- add qos policy group apis
## [1.0.4](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.4) - 2026-04-28

### Bug Fixes

- update apis (user permission & resource group management)

### Features

- add Agent skills description
- add zeno-guidance skills
## [1.0.3](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.3) - 2026-04-19

### Documentation

- update CHANGELOG.md for v1.0.2

### Features

- add output infomation
## [1.0.2](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.2) - 2026-04-18

### Bug Fixes

- description of api yaml & such as vip/ddos command typos
- install.sh

### Documentation

- Add CLI documentation link
- update CHANGELOG.md for v1.0.1
## [1.0.1](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.1) - 2026-04-16

### Bug Fixes

- Add --cli-dry-run for command
- Update markdown and changelog
- install.sh macOS sed 的兼容性问题

### CI/CD

- update release workflow

### Documentation

- update CHANGELOG.md for v1.0.1

### Features

- Add --page-all to fetch all data
## [1.0.0](https://github.com/zenlayer/zenlayercloud-cli/releases/tag/v1.0.0) - 2026-04-11

### Features

- Zenlayer Cloud CLI (zeno)

