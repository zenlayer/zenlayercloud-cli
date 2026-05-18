# Upgrading zeno

The `zeno upgrade` command checks GitHub Releases for a newer version of zeno and installs it in place.

## Synopsis

```
zeno upgrade [flags]
```

## Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--check` | | Check for updates without installing |
| `--list` | | List all available versions |
| `--version <tag>` | | Install a specific version (e.g. `v1.0.1`) |
| `--yes` | `-y` | Skip confirmation prompt |
| `--rollback` | | Roll back to the previous backup version |

## Examples

**Upgrade to the latest version:**

```bash
zeno upgrade
```

```
Current version: v1.0.2
Target version:  v1.0.3
Proceed with upgrade? [y/N]: y
Downloading zeno_1.0.3_darwin_all.tar.gz...
Verifying checksum...
Extracting...
Installing...
Successfully upgraded zeno to v1.0.3
```

**Check whether a newer version is available (no install):**

```bash
zeno upgrade --check
```

```
Update available: v1.0.2 → v1.0.3
```

**List all available versions:**

```bash
zeno upgrade --list
```

```
  v1.0.3
  v1.0.2  *
  v1.0.1
  v1.0.0
```

The `*` marks the currently installed version.

**Install a specific version:**

```bash
zeno upgrade --version v1.0.1
```

**Skip confirmation prompt:**

```bash
zeno upgrade --yes
```

**Roll back to the previous version:**

```bash
zeno upgrade --rollback
```

```
Rolling back to previous version...
Rollback successful.
```

> Note: rollback requires a backup file (`.bak`) left by a previous upgrade. If no backup exists, an error is returned.

## Notes

- Supported platforms: Linux (amd64/arm64) and macOS (universal binary).
- If zeno is installed in a system directory (e.g. `/usr/local/bin`), you may need to run the command with elevated privileges:

  ```bash
  sudo zeno upgrade
  ```
