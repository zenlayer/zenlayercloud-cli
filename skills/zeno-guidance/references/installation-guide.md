# zeno Installation & Configuration Guide

This reference is loaded when `zeno version` fails (not installed) or when the user needs to set up / add a new profile. All values shown are placeholders — never commit real credentials.

Official docs: https://docs.console.zenlayer.com/zenlayer-cli

## Install

### Option A — Quick install (Linux / macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | bash
```

Pin a specific version:

```bash
curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | bash -s -- --version v1.0.2
```

Custom install directory (default is a system-writable path, this lets you stay in user space):

```bash
curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | INSTALL_DIR=~/.local/bin bash
```

Make sure the install directory is on `PATH`. For `~/.local/bin`:

```bash
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc   # or ~/.bashrc
```

### Option B — Pre-built binaries

Download from https://github.com/zenlayer/zenlayercloud-cli/releases and pick the right asset:

| Platform | Architecture | Binary                     |
|----------|--------------|----------------------------|
| Linux    | amd64        | `zeno_linux_amd64`         |
| Linux    | arm64        | `zeno_linux_arm64`         |
| macOS    | Universal    | `zeno_darwin_all`          |
| Windows  | amd64        | `zeno_windows_amd64.exe`   |

`chmod +x` the binary and move it onto `PATH`.

### Option C — From source

Requires Go 1.21+.

```bash
git clone https://github.com/zenlayer/zenlayercloud-cli.git
cd zenlayercloud-cli
make build
# binary lands at bin/zeno
```

### Verify

```bash
zeno version
```

Should print a version string (e.g. `1.0.2`). If not, `PATH` is wrong or the binary didn't install.

## Configure access credentials

Zenlayer uses an **Access Key ID + Access Key Secret** pair. Create one in the console at the user/API key section, then:

```bash
zeno configure
```

Interactive prompts will ask for:

- **Profile name** (default: `default`) — logical name for this key pair
- **Access Key ID** — shown in the console
- **Access Key Secret** — shown once at creation; store securely
- **Language preference** — `en` or `zh`
- **Output format** — `json` (recommended for scripting) or `table`

Files written:

- `~/.zenlayer/credentials.json` — access keys, mode `0600`
- `~/.zenlayer/config.json` — non-secret settings

**Never commit either file to git.** Add `~/.zenlayer/` to your global gitignore if you work in repos under your home directory:

```bash
git config --global core.excludesfile ~/.gitignore_global
echo '.zenlayer/' >> ~/.gitignore_global
```

### View current (non-secret) config

```bash
zeno configure list
```

This prints only `language`, `output`, `profile` — it does **not** print the access key or secret. Safe to share.

### Change a single setting

```bash
zeno configure set output json
zeno configure set language en
zeno configure get output
```

## Multiple profiles

Each profile holds its own key pair and settings. Use them to separate environments (dev / staging / prod), teams, or accounts.

### Create additional profiles

```bash
zeno configure
# When prompted, enter a name like "prod" or "dev"
```

Each new profile is appended to `~/.zenlayer/credentials.json` and `~/.zenlayer/config.json` keyed by name.

### Select a profile per command

```bash
zeno --profile prod zec describe-zones -o json
```

### Select a profile for the whole session

```bash
export ZENLAYER_PROFILE=prod
zeno zec describe-zones -o json
```

### Priority order

Settings resolve highest-to-lowest:

1. Command-line flags (`--profile`, `--access-key-id`, `--output`, ...)
2. Environment variables (see below)
3. `~/.zenlayer/config.json` + `credentials.json`

## Environment variables

| Variable                      | Purpose                                        |
|-------------------------------|------------------------------------------------|
| `ZENLAYER_PROFILE`            | Profile name to use                            |
| `ZENLAYER_ACCESS_KEY_ID`      | Access Key ID (overrides profile)              |
| `ZENLAYER_ACCESS_KEY_SECRET`  | Access Key Secret (overrides profile)          |
| `ZENLAYER_OUTPUT`             | `json` or `table`                              |
| `ZENLAYER_DEBUG`              | `true` / `false` — enables request logging     |

Example — CI or ephemeral shells without a credentials file:

```bash
export ZENLAYER_ACCESS_KEY_ID="AKID_EXAMPLE_PLACEHOLDER"
export ZENLAYER_ACCESS_KEY_SECRET="SECRET_EXAMPLE_PLACEHOLDER"
export ZENLAYER_OUTPUT=json
zeno zec describe-zones
```

Unset after use (`unset ZENLAYER_ACCESS_KEY_SECRET`) so keys don't linger in the environment.

## Credential hygiene checklist

- [ ] `~/.zenlayer/credentials.json` is mode `0600` (owner read/write only)
- [ ] `.zenlayer/` appears in your global gitignore
- [ ] No access key or secret appears in shell history, commit messages, code comments, or logs
- [ ] Rotate keys in the console if one has ever been shared, pasted into chat, or committed
- [ ] CI uses environment variables injected from a secret manager — never a checked-in credentials file

## Smoke test after configuring

```bash
zeno version                         # CLI is on PATH
zeno configure list                  # profile is active
zeno zec describe-zones -o json      # signing works end-to-end
```

If step 3 returns a `zoneSet`, you're done.
