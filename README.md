# Zenlayer Cloud CLI

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/zenlayer/zenlayercloud-cli)](https://goreportcard.com/report/github.com/zenlayer/zenlayercloud-cli)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-00ADD8?logo=go)](https://golang.org/dl/)

The official command line interface for [Zenlayer Cloud](https://console.zenlayer.com/).

**[English](README.md)** | **[简体中文](README_zh-CN.md)**

## Overview

Zenlayer Cloud CLI (`zeno`) is a powerful command-line tool that enables you to manage your Zenlayer Cloud resources efficiently. It provides a unified interface to interact with various Zenlayer Cloud services.

## Features

- 🚀 **Easy to Use** - Intuitive command structure with helpful prompts
- 🔐 **Secure** - Credentials stored with restricted file permissions
- 📝 **Multiple Output Formats** - JSON and table output support
- 🔄 **Multi-Profile Support** - Manage multiple environments easily
- 🌐 **Cross-Platform** - Support for Linux, macOS, and Windows
- 💻 **Shell Completion** - Auto-completion for Bash, Zsh, Fish, and PowerShell
- 🐛 **Debug Mode** - Detailed logging for troubleshooting

## Installation

### Quick Install (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | bash
```

Install a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | bash -s -- --version v0.1.0
```

Custom install directory:

```bash
curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | INSTALL_DIR=~/.local/bin bash
```

### Pre-built Binaries

Download pre-built binaries from the [Releases](https://github.com/zenlayer/zenlayercloud-cli/releases) page.

Choose the appropriate binary for your platform:

| Platform | Architecture | Binary |
|----------|-------------|--------|
| Linux | amd64 | `zeno_linux_amd64` |
| Linux | arm64 | `zeno_linux_arm64` |
| macOS | Universal | `zeno_darwin_all` |
| Windows | amd64 | `zeno_windows_amd64.exe` |

### From Source

Requires Go 1.21 or later.

```bash
git clone https://github.com/zenlayer/zenlayercloud-cli.git
cd zenlayercloud-cli
make build
```

The binary will be available at `bin/zeno`.

### Verify Installation

```bash
zeno version
```

## Quick Start

### 1. Configure Credentials

Run the interactive configuration:

```bash
zeno configure
```

This will prompt you for:
- Profile name (default: `default`)
- Access Key ID
- Access Key Secret
- Language preference (en/zh)
- Output format (json/table)

### 2. Start Using the CLI

```bash
# List load balancers
zeno zlb describe-load-balancers

# Create a load balancer
zeno zlb create-load-balancer --help

# Get bandwidth cluster information
zeno traffic describe-bandwidth-clusters
```

## Configuration

Configuration files are stored in `~/.zenlayer/`:

- `config.json` - General settings (language, output format)
- `credentials.json` - Access credentials (with restricted permissions)

### Multiple Profiles

You can configure multiple profiles for different environments:

```bash
# Configure a new profile
zeno configure
# Enter "prod" when prompted for profile name

# Use a specific profile
zeno --profile prod describe-load-balancers

# Or set via environment variable
export ZENLAYER_PROFILE=prod
```

### Configuration Priority

Settings are applied in the following priority (highest to lowest):

1. Command line flags (`--profile`, `--output`, etc.)
2. Environment variables (`ZENLAYER_PROFILE`, `ZENLAYER_ACCESS_KEY_ID`, etc.)
3. Configuration files

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ZENLAYER_PROFILE` | Profile name to use |
| `ZENLAYER_ACCESS_KEY_ID` | Access Key ID |
| `ZENLAYER_ACCESS_KEY_SECRET` | Access Key Secret |
| `ZENLAYER_OUTPUT` | Output format (`json`/`table`) |
| `ZENLAYER_DEBUG` | Enable debug mode (`true`/`false`) |

**Example:**

```bash
# Set output format via environment variable
export ZENLAYER_OUTPUT=json
zeno zlb describe-load-balancers

# Or use inline
ZENLAYER_OUTPUT=table zeno zlb describe-load-balancers
```

## Commands

### Configure

```bash
# Interactive configuration
zeno configure

# List current configuration
zeno configure list

# Get a configuration value
zeno configure get <key>

# Set a configuration value
zeno configure set <key> <value>
```

### Version

```bash
zeno version
```

### Load Balancer Commands

```bash
# List load balancers
zeno zlb describe-load-balancers

# Create a load balancer
zeno zlb create-load-balancer [flags]

# Modify load balancer attributes
zeno zlb modify-load-balancers-attribute [flags]

# Delete load balancers
zeno zlb terminate-load-balancer [flags]
```

### Bandwidth Cluster Commands

```bash
# List bandwidth clusters
zeno traffic describe-bandwidth-clusters

# Create a bandwidth cluster
zeno traffic create-bandwidth-cluster [flags]

# Get cluster usage
zeno traffic describe-bandwidth-cluster-usage [flags]
```

Use `zeno [command] --help` for more information about a command.

## Shell Completion

Enable tab completion for commands, subcommands, and flags.

### Bash

```bash
# Linux
zeno completion bash > /etc/bash_completion.d/zeno

# macOS (requires bash-completion)
zeno completion bash > $(brew --prefix)/etc/bash_completion.d/zeno
```

### Zsh

```bash
# Option 1: use default fpath (may require sudo)
zeno completion zsh > "${fpath[1]}/_zeno"

# Option 2: use custom completion directory
mkdir -p ~/.zsh/completions
zeno completion zsh > ~/.zsh/completions/_zeno
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -U compinit; compinit' >> ~/.zshrc
```

### Fish

```bash
zeno completion fish > ~/.config/fish/completions/zeno.fish
```

### PowerShell

```powershell
zeno completion powershell | Out-String | Invoke-Expression
```

Restart your shell after setup for the changes to take effect.

### Uninstall Completion

```bash
zeno completion --uninstall                    # uninstall all (bash, zsh, fish, powershell)
zeno completion --uninstall [bash|zsh|fish|powershell]  # uninstall specific shell
```

Removes completion from standard installation paths. Restart your shell after uninstalling.

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--profile` | `-p` | Profile name to use |
| `--output` | `-o` | Output format: json, table |
| `--query` | `-q` | JMESPath query to filter response (see below) |
| `--access-key-id` | | Access Key ID (overrides config) |
| `--access-key-secret` | | Access Key Secret (overrides config) |
| `--debug` | | Enable debug mode |
| `--help` | `-h` | Help for command |

### Response Filtering with --query

Use the `--query` (or `-q`) flag to filter and transform API response output using [JMESPath](https://jmespath.org/) syntax. This is useful for extracting specific fields, filtering arrays, or formatting output for scripts and pipelines.

**Examples:**

```bash
# Extract all instance IDs from the dataSet array
zeno zec describe-instances --query "dataSet[*].instanceId"

# Filter instances with state RUNNING
zeno zec describe-instances --query "dataSet[?state=='RUNNING']"

# Get only load balancer names
zeno zlb describe-load-balancers -o json -q "loadBalancerSet[*].loadBalancerName"

# Extract requestId from response
zeno zec describe-instances --query "requestId"

# Filter bandwidth clusters by status
zeno traffic describe-bandwidth-clusters -q "dataSet[?status=='active']"
```

**JMESPath Quick Reference:**
- `foo` - Access field `foo`
- `foo.bar` - Nested path
- `items[*].id` - Project: get `id` from each element in `items`
- `items[?status=='active']` - Filter: elements where `status` equals `active`
- `items[0]` - First element of array

See [JMESPath Tutorial](https://jmespath.org/tutorial.html) for full syntax.

## Development

### Prerequisites

- Go 1.21+
- golangci-lint (for linting)

### Build

```bash
make build
```

### Test

```bash
make test
```

### Lint

```bash
make lint
```

### Build for All Platforms

```bash
make build-all
```

### Build Mac Universal Binary

On macOS, build a fat binary for both Intel and Apple Silicon:

```bash
make build-mac-universal
```

### Version

Version is set via `-ldflags` during build. When building from a git tag (e.g. `v1.0.0`), the binary version will be `1.0.0` (the `v` prefix is stripped). For untagged commits, version defaults to `dev`.

```bash
# Build with explicit version
make build VERSION=1.0.0

# Or pass TAG (v prefix stripped automatically)
make build TAG=v1.0.0
```

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Security

If you discover a security vulnerability, please report it by emailing security@zenlayer.com. Do not open a public issue.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a list of changes.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- 📖 [Documentation](https://docs.zenlayer.com)
- 💬 [GitHub Issues](https://github.com/zenlayer/zenlayercloud-cli/issues)
- 📧 [Email Support](mailto:support@zenlayer.com)

## Acknowledgments

This project uses the following open source packages:

- [Cobra](https://github.com/spf13/cobra) - A Commander for modern Go CLI interactions
- [Zenlayer Cloud SDK for Go](https://github.com/zenlayer/zenlayercloud-sdk-go) - Official Zenlayer Cloud SDK
