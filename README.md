# Zenlayer Cloud CLI

The official command line interface for [Zenlayer Cloud](https://www.zenlayer.com/).

## Installation

### From Source

Requires Go 1.21 or later.

```bash
git clone https://github.com/zenlayer/zenlayercloud-cli.git
cd zenlayercloud-cli
make build
```

The binary will be available at `bin/zencli`.

### Pre-built Binaries

Download pre-built binaries from the [Releases](https://github.com/zenlayer/zenlayercloud-cli/releases) page.

## Quick Start

### Configure Credentials

Run the interactive configuration:

```bash
zencli configure
```

This will prompt you for:
- Profile name (default: `default`)
- Access Key ID
- Access Key Secret
- Language preference (en/zh)
- Output format (json/table)

### Verify Installation

```bash
zencli version
```

## Configuration

Configuration files are stored in `~/.zenlayer/`:

- `config.json` - General settings (language, output format)
- `credentials.json` - Access credentials (with restricted permissions)

### Multiple Profiles

You can configure multiple profiles for different environments:

```bash
# Configure a new profile
zencli configure
# Enter "prod" when prompted for profile name

# Use a specific profile
zencli --profile prod <command>

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
| `ZENLAYER_DEBUG` | Enable debug mode (`true`/`false`) |

## Commands

### Configure

```bash
# Interactive configuration
zencli configure

# List current configuration
zencli configure list

# Get a configuration value
zencli configure get <key>

# Set a configuration value
zencli configure set <key> <value>
```

### Version

```bash
zencli version
```

## Shell Completion

Enable tab completion for commands, subcommands, and flags.

### Bash

```bash
# Linux
zencli completion bash > /etc/bash_completion.d/zencli

# macOS (requires bash-completion)
zencli completion bash > $(brew --prefix)/etc/bash_completion.d/zencli
```

### Zsh

```bash
# Option 1: use default fpath (may require sudo)
zencli completion zsh > "${fpath[1]}/_zencli"

# Option 2: use custom completion directory
mkdir -p ~/.zsh/completions
zencli completion zsh > ~/.zsh/completions/_zencli
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -U compinit; compinit' >> ~/.zshrc
```

### Fish

```bash
zencli completion fish > ~/.config/fish/completions/zencli.fish
```

### PowerShell

```powershell
zencli completion powershell | Out-String | Invoke-Expression
```

Restart your shell after setup for the changes to take effect.

### Uninstall

```bash
zencli completion --uninstall                    # uninstall all (bash, zsh, fish, powershell)
zencli completion --uninstall [bash|zsh|fish|powershell]  # uninstall specific shell
```

Removes completion from standard installation paths. Restart your shell after uninstalling.

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--profile` | `-p` | Profile name to use |
| `--output` | `-o` | Output format: json, table |
| `--access-key-id` | | Access Key ID (overrides config) |
| `--access-key-secret` | | Access Key Secret (overrides config) |
| `--debug` | | Enable debug mode |

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

## License

Apache License 2.0. See [LICENSE](LICENSE) for details.
