# Zenlayer Cloud CLI

[![Go Version](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[Zenlayer Cloud](https://console.zenlayer.com/) 官方命令行工具。

[中文版](README_zh-CN.md) · [English](README.md) · [Reference](https://docs.console.zenlayer.com/zenlayer-cli/cn) · [Changelog](CHANGELOG.md)

[安装](#安装) · [Agent Skills](#agent-skills) · [快速开始](#快速开始) · [配置](#配置) · [命令](#命令)

## 概述

Zenlayer Cloud CLI (`zeno`) 是一个功能强大的命令行工具，可帮助您高效管理 Zenlayer Cloud 资源。它提供统一的界面来与各种 Zenlayer Cloud 服务交互——同时面向人类用户和 AI Agent，并内置结构化 [Agent Skills](./skills/)。

## 特性

- 🚀 **易于使用** - 直观的命令结构和友好的提示
- 🔐 **安全可靠** - 凭证以受限文件权限存储
- 📝 **多种输出格式** - 支持 JSON 和表格输出
- 🔄 **多配置文件支持** - 轻松管理多个环境
- 🌐 **跨平台** - 支持 Linux、macOS 和 Windows
- 💻 **Shell 自动补全** - 支持 Bash、Zsh、Fish 和 PowerShell
- 📄 **自动分页** - 使用 `--page-all` 一次获取所有分页数据
- 🐛 **调试模式** - 详细的日志记录便于故障排查
- 🤖 **Agent 原生** - 开箱即用的结构化 [Agent Skills](./skills/)

## 安装

### 快速安装（Linux / macOS）

```bash
curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | bash
```

安装指定版本：

```bash
curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | bash -s -- --version v0.1.0
```

自定义安装目录：

```bash
curl -fsSL https://raw.githubusercontent.com/zenlayer/zenlayercloud-cli/main/install.sh | INSTALL_DIR=~/.local/bin bash
```

### 预编译二进制文件

从 [Releases](https://github.com/zenlayer/zenlayercloud-cli/releases) 页面下载预编译的二进制文件。

根据您的平台选择相应的二进制文件：

| 平台 | 架构 | 二进制文件 |
|----------|-------------|--------|
| Linux | amd64 | `zeno_linux_amd64` |
| Linux | arm64 | `zeno_linux_arm64` |
| macOS | Universal | `zeno_darwin_all` |
| Windows | amd64 | `zeno_windows_amd64.exe` |

### 从源码编译

需要 Go 1.21 或更高版本。

```bash
git clone https://github.com/zenlayer/zenlayercloud-cli.git
cd zenlayercloud-cli
make build
```

编译后的二进制文件位于 `bin/zeno`。

### 验证安装

```bash
zeno version
```

## Agent Skills

| Skill | 描述 |
|-------|------|
| `zeno-guidance` | 安全操作 `zeno`：写/删操作的 dry-run 预检与确认机制、仅 JSON 输出、凭证安全规则，以及通过 `--help` 发现命令 |

### 安装 Skill

```bash
npx skills add zenlayer/zenlayercloud-cli -y -g
```

安装后重启 Agent 会话。


## 快速开始

### 1. 配置凭证

运行交互式配置：

```bash
zeno configure
```

系统将提示您输入：
- 配置文件名称（默认：`default`）
- Access Key ID
- Access Key Secret
- 语言偏好（en/zh）
- 输出格式（json/table）

### 2. 开始使用 CLI

```bash
# 列出负载均衡器
zeno zlb describe-load-balancers

# 创建负载均衡器
zeno zlb create-load-balancer --help

# 获取带宽集群信息
zeno traffic describe-bandwidth-clusters
```

## 配置

配置文件存储在 `~/.zenlayer/` 目录中：

- `config.json` - 通用设置（语言、输出格式）
- `credentials.json` - 访问凭证（具有受限权限）

### 多配置文件

您可以配置多个配置文件以管理不同的环境：

```bash
# 配置新的配置文件
zeno configure
# 在提示输入配置文件名称时输入 "prod"

# 使用指定的配置文件
zeno --profile prod describe-load-balancers

# 或通过环境变量设置
export ZENLAYER_PROFILE=prod
```

### 配置优先级

设置按以下优先级应用（从高到低）：

1. 命令行标志（`--profile`、`--output` 等）
2. 环境变量（`ZENLAYER_PROFILE`、`ZENLAYER_ACCESS_KEY_ID` 等）
3. 配置文件

### 环境变量

| 变量 | 描述 |
|----------|-------------|
| `ZENLAYER_PROFILE` | 使用的配置文件名称 |
| `ZENLAYER_ACCESS_KEY_ID` | Access Key ID |
| `ZENLAYER_ACCESS_KEY_SECRET` | Access Key Secret |
| `ZENLAYER_OUTPUT` | 输出格式（`json`/`table`） |
| `ZENLAYER_DEBUG` | 启用调试模式（`true`/`false`） |

**示例：**

```bash
# 通过环境变量设置输出格式
export ZENLAYER_OUTPUT=json
zeno zlb describe-load-balancers

# 或使用内联方式
ZENLAYER_OUTPUT=table zeno zlb describe-load-balancers
```

## 命令

### 配置命令

```bash
# 交互式配置
zeno configure

# 列出当前配置
zeno configure list

# 获取配置值
zeno configure get <key>

# 设置配置值
zeno configure set <key> <value>
```

### 版本命令

```bash
zeno version
```

### 负载均衡器命令

```bash
# 列出负载均衡器
zeno zlb describe-load-balancers

# 创建负载均衡器
zeno zlb create-load-balancer [flags]

# 修改负载均衡器属性
zeno zlb modify-load-balancers-attribute [flags]

# 删除负载均衡器
zeno zlb terminate-load-balancer [flags]
```

### 带宽集群命令

```bash
# 列出带宽集群
zeno traffic describe-bandwidth-clusters

# 创建带宽集群
zeno traffic create-bandwidth-cluster [flags]

# 获取集群使用情况
zeno traffic describe-bandwidth-cluster-usage [flags]
```

使用 `zeno [command] --help` 查看命令的更多信息。

## Shell 自动补全

启用命令、子命令和标志的 Tab 补全功能。

### Bash

```bash
# Linux
zeno completion bash > /etc/bash_completion.d/zeno

# macOS（需要 bash-completion）
zeno completion bash > $(brew --prefix)/etc/bash_completion.d/zeno
```

### Zsh

```bash
# 选项 1：使用默认 fpath（可能需要 sudo）
zeno completion zsh > "${fpath[1]}/_zeno"

# 选项 2：使用自定义补全目录
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

设置完成后重启您的 Shell 以使更改生效。

### 卸载补全

```bash
zeno completion --uninstall                    # 卸载所有（bash、zsh、fish、powershell）
zeno completion --uninstall [bash|zsh|fish|powershell]  # 卸载特定 shell
```

从标准安装路径中移除补全。卸载后请重启您的 Shell。

## 全局标志

| 标志 | 简写 | 描述 |
|------|-------|-------------|
| `--profile` | `-p` | 使用的配置文件名称 |
| `--output` | `-o` | 输出格式：json、table |
| `--query` | `-q` | 使用 JMESPath 过滤响应（见下文） |
| `--page-all` | | 自动获取所有分页数据并合并（仅列表命令） |
| `--access-key-id` | | Access Key ID（覆盖配置） |
| `--access-key-secret` | | Access Key Secret（覆盖配置） |
| `--cli-dry-run` | | 预览 API 请求，不实际发送 |
| `--debug` | | 启用调试模式 |
| `--help` | `-h` | 命令帮助 |

### 使用 --query 过滤响应

使用 `--query`（或 `-q`）标志可通过 [JMESPath](https://jmespath.org/) 语法过滤和转换 API 响应输出。适用于提取特定字段、过滤数组或为脚本与管道格式化输出。

**示例：**

```bash
# 提取 dataSet 数组中所有实例 ID
zeno zec describe-instances --query "dataSet[*].instanceId"

# 筛选状态为 RUNNING 的实例
zeno zec describe-instances --query "dataSet[?state=='RUNNING']"

# 仅获取负载均衡器名称
zeno zlb describe-load-balancers -o json -q "loadBalancerSet[*].loadBalancerName"

# 从响应中提取 requestId
zeno zec describe-instances --query "requestId"

# 按状态筛选带宽集群
zeno traffic describe-bandwidth-clusters -q "dataSet[?status=='active']"
```

**JMESPath 快速参考：**
- `foo` - 访问字段 `foo`
- `foo.bar` - 嵌套路径
- `items[*].id` - 投影：获取 `items` 中每个元素的 `id`
- `items[?status=='active']` - 过滤：`status` 等于 `active` 的元素
- `items[0]` - 数组的第一个元素

完整语法请参阅 [JMESPath 教程](https://jmespath.org/tutorial.html)。

### 使用 --page-all 自动分页

支持 `--page-num` / `--page-size` 的列表命令默认只返回单页数据。使用 `--page-all` 可自动遍历所有页，并将结果合并为一条完整响应。

```bash
# 一次获取所有实例
zeno zec describe-instances --page-all

# 控制每次请求的条数（默认：100）
zeno zec describe-instances --page-all --page-size 50

# 与其他参数组合使用
zeno zec describe-instances --page-all --output table
zeno zec describe-instances --page-all --query "instanceSet[*].instanceId"
```

`--page-all` 仅在支持分页的命令中出现，不适用于不支持分页的命令。

## 开发

### 前置要求

- Go 1.21+
- golangci-lint（用于代码检查）

### 编译

```bash
make build
```

### 测试

```bash
make test
```

### 代码检查

```bash
make lint
```

### 编译所有平台

```bash
make build-all
```

### 编译 macOS 通用二进制文件

在 macOS 上，为 Intel 和 Apple Silicon 编译通用二进制文件：

```bash
make build-mac-universal
```

### 版本

版本通过编译时的 `-ldflags` 设置。当从 git 标签（如 `v1.0.0`）编译时，二进制文件版本为 `1.0.0`（`v` 前缀会被去除）。对于未打标签的提交，版本默认为 `dev`。

```bash
# 使用显式版本编译
make build VERSION=1.0.0

# 或传递 TAG（v 前缀会自动去除）
make build TAG=v1.0.0
```

## 贡献

欢迎社区贡献！如果您发现 Bug 或有功能建议，请提交 [Issue](https://github.com/zenlayer/zenlayercloud-cli/issues) 或 Pull Request。

对于较大的改动，建议先通过 Issue 与我们讨论。

## 许可证

本项目采用 Apache License 2.0 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 支持

- 📖 [文档](https://docs.zenlayer.com)
- 💬 [GitHub Issues](https://github.com/zenlayer/zenlayercloud-cli/issues)
- 📧 [邮件支持](mailto:support@zenlayer.com)

## 致谢

本项目使用了以下开源软件包：

- [Cobra](https://github.com/spf13/cobra) - 现代 Go CLI 交互的命令行框架
- [Zenlayer Cloud SDK for Go](https://github.com/zenlayer/zenlayercloud-sdk-go) - Zenlayer Cloud 官方 SDK
