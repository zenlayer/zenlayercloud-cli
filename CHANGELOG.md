# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project setup with Cobra CLI framework
- Configuration management with multi-profile support
- Credentials stored separately with restricted permissions (0600)
- Interactive `configure` command for setting up credentials
- `configure list` command to display current profile settings
- `configure get` and `configure set` commands for individual settings
- `version` command with build information
- JSON and table output formats
- Debug mode via `--debug` flag or `ZENLAYER_DEBUG` environment variable
- Configuration priority: CLI flags > environment variables > config files
- Cross-platform build support (Linux, macOS, Windows)

### Security
- Credentials file stored with 0600 permissions
- Access key secret hidden during interactive input
