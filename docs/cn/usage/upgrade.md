# 升级 zeno

`zeno upgrade` 命令会检查 GitHub Releases 是否有新版本，并就地替换当前安装的二进制文件。

## 语法

```
zeno upgrade [flags]
```

## 参数

| 参数 | 短参数 | 说明 |
|------|--------|------|
| `--check` | | 仅检查是否有新版本，不安装 |
| `--list` | | 列出所有可用版本 |
| `--version <tag>` | | 安装指定版本（例如 `v1.0.1`） |
| `--yes` | `-y` | 跳过确认提示直接安装 |
| `--rollback` | | 回滚到上一个备份版本 |

## 示例

**升级到最新版本：**

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

**仅检查是否有新版本（不安装）：**

```bash
zeno upgrade --check
```

```
Update available: v1.0.2 → v1.0.3
```

**列出所有可用版本：**

```bash
zeno upgrade --list
```

```
  v1.0.3
  v1.0.2  *
  v1.0.1
  v1.0.0
```

`*` 标记当前已安装的版本。

**安装指定版本：**

```bash
zeno upgrade --version v1.0.1
```

**跳过确认提示：**

```bash
zeno upgrade --yes
```

**回滚到上一个版本：**

```bash
zeno upgrade --rollback
```

```
Rolling back to previous version...
Rollback successful.
```

> 说明：回滚需要上一次升级留下的 `.bak` 备份文件。若备份不存在，命令会报错退出。

## 说明

- 支持平台：Linux（amd64/arm64）和 macOS（通用二进制）。
- 如果 zeno 安装在系统目录（如 `/usr/local/bin`），可能需要提权执行：

  ```bash
  sudo zeno upgrade
  ```
