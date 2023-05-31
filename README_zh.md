# tdl

[English](README.md) | 简体中文

![](https://img.shields.io/github/go-mod/go-version/iyear/tdl?style=flat-square)
![](https://img.shields.io/github/license/iyear/tdl?style=flat-square)
![](https://img.shields.io/github/actions/workflow/status/iyear/tdl/master.yml?branch=master&style=flat-square)
![](https://img.shields.io/github/v/release/iyear/tdl?color=red&style=flat-square)
![](https://img.shields.io/github/downloads/iyear/tdl/total?style=flat-square)

📥 Telegram Downloader, but more than a downloader

> **Note**
> 中文文档可能落后于英文文档，如果有问题请先查看英文文档。

## 目录

* [特性](#特性)
* [预览](#预览)
* [安装](#安装)
* [快速开始](#快速开始)
* [工作流](#工作流)
* [使用方法](#使用方法)
    * [基础设置](#基础设置)
    * [登录](#登录)
    * [下载](#下载)
    * [上传](#上传)
    * [迁移](#迁移)
    * [实用工具](#实用工具)
* [环境变量](#环境变量)
* [数据](#数据)
* [命令](#命令)
* [最佳实践](#最佳实践)
* [疑难解答](#疑难解答)
* [FAQ](#faq)

## 特性

- 单文件启动
- 低资源占用
- 吃满你的带宽
- 比官方客户端更快
- 支持从受保护的会话中下载文件
- 支持上传文件至 Telegram

## 预览

预览中的速度已经达到了代理的限制，同时**速度取决于你是否是付费用户**

![](img/preview.gif)

## 安装

你可以从 [releases](https://github.com/iyear/tdl/releases/latest) 下载预编译的二进制文件，或者使用下面的方法安装：

### Linux & macOS

<details>

- 使用一键脚本安装：

`tdl` 将会被安装到 `/usr/local/bin/tdl`，同时脚本也可以用于升级 `tdl`。

```shell
# 安装最新版本
curl -sSL https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.sh | sudo bash
```

```shell
# 使用 `ghproxy.com` 加速下载
curl -sSL https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.sh | sudo bash -s -- --proxy
# 安装指定版本
curl -sSL https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.sh | sudo bash -s -- --version VERSION
```

- 使用包管理器安装

为包管理器部分做贡献：[提交 issue](https://github.com/iyear/tdl/issues/new/choose)

</details>

### Windows

<details>

- 使用一键脚本安装(管理员)：

`tdl` 将会被安装到 `$Env:SystemDrive\tdl`（该路径会被添加到 `PATH` 中），同时脚本也可以用于升级 `tdl`。

```powershell
# 安装最新版本
iwr -useb https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.ps1 | iex
```

```powershell
# 使用 `ghproxy.com` 加速下载
$Script = iwr -useb https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.ps1; $Block = [ScriptBlock]::Create($Script); Invoke-Command -ScriptBlock $Block -ArgumentList "", "$True"
# 安装指定版本
$Env:TDLVersion = "VERSION"
$Script = iwr -useb https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.ps1; $Block = [ScriptBlock]::Create($Script); Invoke-Command -ScriptBlock $Block -ArgumentList "$Env:TDLVersion"
```

- 使用包管理器安装:

```powershell
# Scoop (Windows) https://scoop.sh/#/apps?s=2&d=1&o=true&p=1&q=telegram+downloader
scoop bucket add extras
scoop install telegram-downloader
```

</details>

## 快速开始

```shell
# 借助电脑上已有的官方桌面客户端登录
tdl login -n quickstart
# 如果设置了 passcode, 需要指定 passcode
tdl login -n quickstart -p YOUR_PASSCODE
# 如果路径非默认路径，需要指定路径
tdl login -n quickstart -d /path/to/TelegramDesktop
# 如果希望使用电话验证码登录，使用以下命令
tdl login -n quickstart --code

tdl dl -n quickstart -u https://t.me/telegram/193
```

## 工作流

<details>

该部分只展示工作流，而非所有设置项。所以你还需要阅读 [使用方法](#使用方法) 并设置你需要的设置项。

### 从消息链接下载文件

```shell
export TDL_NS=iyear # 设置账号
tdl login
tdl dl -u https://t.me/tdl/1 -u https://t.me/tdl/2
```

### 从受保护的会话下载文件

```shell
export TDL_NS=iyear # 设置账号
tdl login
tdl chat export -o result.json
tdl dl -f result.json
```

### 迁移数据至远程服务器

```shell
export TDL_NS=iyear # 设置账号
tdl login
tdl backup -d backup.zip
# 上传 backup.zip 到远程服务器
tdl recover -f backup.zip # 在远程服务器上执行
```

### 持续下载并忽略错误

推荐的做法是使用守护进程 + `tdl` 下载，因为某些错误可能需要重启 `tdl` 才能正常工作。

`tdl` 不负责守护进程，你可以根据不同平台选择不同的守护进程，例如 Linux 可以使用 systemd。

命令: `tdl dl <其他参数> --continue`

这样 `tdl` 就会在出现错误时重启，并继续执行下载任务。

</details>

## 使用方法

- 获取帮助

```shell
tdl -h
```

- 检查版本

```shell
tdl version
```

- 自动补全

根据你的 shell 运行相应的命令，并在所有会话中启用 shell 补全：

```shell
# bash
echo "source <(tdl completion bash)" >> ~/.bashrc
# zsh
echo "source <(tdl completion zsh)" >> ~/.zshrc
# fish
echo "tdl completion fish | source" >> ~/.config/fish/config.fish
# powershell
Add-Content -Path $PROFILE -Value "tdl completion powershell | Out-String | Invoke-Expression"
```

### 基础设置

> 该文档其他部分不会写基础设置，因此请根据需要添加基础设置。

每个命名空间代表一个 Telegram 账号

你应该在每次执行命令时设置命名空间：

```shell
tdl -n iyear
# 或
export TDL_NS=iyear # 推荐做法
```

- (可选) 设置代理。目前仅支持 socks5 代理：

```shell
tdl --proxy socks5://localhost:1080
# 或
export TDL_PROXY=socks5://localhost:1080 # 推荐做法
```

- (可选) 设置 NTP 服务器。如果为空，则使用系统时间：

```shell
tdl --ntp pool.ntp.org
# 或
export TDL_NTP=pool.ntp.org # 推荐做法
```

- (可选) 设置 Telegram 连接重试超时时间。默认为 2m：

> **Note**
> 如果网络环境较差请设置更高的超时时间或 0(无限)

```shell
tdl --reconnect-timeout 1m30s
# or
export TDL_RECONNECT_TIMEOUT=1m30s
```


### 登录

> 当你第一次使用 tdl 时，你需要登录以获取一个 Telegram 会话

- 如果你有 [Telegram Desktop](https://desktop.telegram.org/) 存在于本机，你可以导入现有的会话。

这将降低被封禁的风险，但尚未经过验证：

```shell
tdl login
# 如果设置了 passcode, 需要指定 passcode
tdl login -p YOUR_PASSCODE
# 如果路径非默认路径，需要指定路径
tdl login -d /path/to/TelegramDesktop
```

- 使用短信验证码的方式登录：

```shell
tdl login --code
```

### 下载

> 如果你需要更高的下载速度，请设置更高的 `threads`，但是不要随意设置过大的 `threads`。

- 从消息链接下载（受保护的）文件：

```shell
tdl dl -u https://t.me/tdl/1 -u https://t.me/tdl/2
```

- 从 [官方客户端导出的 JSON](docs/desktop_export.md) 下载文件：

```shell
tdl dl -f result1.json -f result2.json
```

- 同时从消息链接和导出文件下载：

```shell
tdl dl \
-u https://t.me/tdl/1 -u https://t.me/tdl/2 \
-f result1.json -f result2.json
```

- 使用 8 个线程，每个线程 512KiB(MAX) 的分片大小，4 个并发任务下载：

```shell
tdl dl -u https://t.me/tdl/1 -t 8 -s 524288 -l 4
```

- 根据 MIME 类型下载真实的文件扩展名：

> **Note**
> 如果文件扩展名与 MIME 类型不匹配，tdl 将重命名文件以使用正确的扩展名。
>
> 副作用：例如 `.apk` 文件，它将被重命名为 `.zip`。

```shell
tdl dl -u https://t.me/tdl/1 --rewrite-ext
```

- 跳过已下载的文件：

> **Note**
> 判断依据：文件名（不包括扩展名）和大小相同

```shell
tdl dl -u https://t.me/tdl/1 --skip-same
```

- 下载文件到自定义目录：

```shell
tdl dl -u https://t.me/tdl/1 -d /path/to/dir
```

- 按照自定义顺序下载文件：

> **Note**
> 不同的顺序会影响“恢复下载”功能

```shell
# 按照时间倒序下载文件（从最新到最旧）
tdl dl -f result.json --desc
# 默认按照时间顺序下载文件（从最旧到最新）
tdl dl -f result.json
```

- 使用 [takeout session](https://arabic-telethon.readthedocs.io/en/stable/extra/examples/telegram-client.html#exporting-messages) 下载文件：

> **Note**
> If you plan to download a lot of media, you may prefer to do this within a takeout session. Takeout sessions let you export data from your account with lower flood wait limits.
> 如果你想下载大量的媒体文件，推荐在 takeout session 下进行。Takeout session 可以让你以更低的接口限制导出你的账户数据。

```shell
tdl dl -u https://t.me/tdl/1 --takeout
```

- 使用扩展名过滤器下载文件：

> **Note**
> 扩展名只与文件名匹配，而不与 MIME 类型匹配。因此，它可能无法按预期工作。
>
> 白名单和黑名单不能同时使用。

```shell
# 白名单过滤，只下载扩展名为 `.jpg` `.png` 的文件
tdl dl -u https://t.me/tdl/1 -i jpg,png

# 黑名单过滤，下载除了 `.mp4` `.flv` 扩展名的所有文件
tdl dl -u https://t.me/tdl/1 -e mp4,flv
```

- 使用自定义文件名模板下载文件：

请参考 [模板指南](docs/template.md) 以获取更多详细信息。

```shell
tdl dl -u https://t.me/tdl/1 \
--template "{{ .DialogID }}_{{ .MessageID }}_{{ .DownloadDate }}_{{ .FileName }}"
```

- 无需 UI 交互的恢复或重新开始下载：

```shell
# 恢复下载
tdl dl -u https://t.me/tdl/1 --continue
# 重新下载
tdl dl -u https://t.me/tdl/1 --restart
```

- 完整例子:

```shell
tdl dl --debug --ntp pool.ntp.org \
-n iyear --proxy socks5://localhost:1080 \
-u https://t.me/tdl/1 -u https://t.me/tdl/2 \
-f result1.json -f result2.json \
--rewrite-ext --skip-same -i jpg,png \
-d /path/to/dir --desc \
-t 8 -s 262144 -l 4
```

### 上传

> 部分指令和高级选项与 **下载** 相同

- 上传文件到 `收藏夹`，并排除指定的文件扩展名：

```shell
tdl up -p /path/to/file -p /path/to/dir -e .so -e .tmp
```

- 使用 8 个线程，512KiB(MAX) 分片大小，4 个并发任务上传文件：

```shell
tdl up -p /path/to/file -t 8 -s 524288 -l 4
```

- 删除本地已上传成功的文件：

```shell
tdl up -p /path/to/file --rm
```

- 上传文件到自定义会话：

```shell
# CHAT_INPUT 可接受例子: `@iyear`, `iyear`, `123456789`(会话 ID), `https://t.me/iyear`, `+1 123456789`

# 空会话意味着 `收藏夹`
tdl up -p /path/to/file -c CHAT_INPUT
```

- 完整例子:

```shell
tdl up --debug --ntp pool.ntp.org \
-n iyear --proxy socks5://localhost:1080 \
-p /path/to/file -p /path/to/dir \
-e .so -e .tmp \
-t 8 -s 262144 -l 4
-c @iyear
```

### 迁移

> 备份或恢复你的数据

- 备份（默认文件名：`tdl-backup-<time>.zip`）：

```shell
tdl backup
# 或者指定备份文件路径
tdl backup -d /path/to/backup.zip
```

- 恢复：

```shell
tdl recover -f /path/to/backup.zip
```

### 实用工具

- 列出所有会话：

```shell
tdl chat ls

# 输出为 JSON 格式
tdl chat ls -o json

# 指定使用表达式引擎的过滤器，默认值为 `true`(匹配所有)
# 如果你对表达式引擎有任何问题，请发起新的 ISSUE
# 表达式引擎文档: https://expr.medv.io/docs/Language-Definition

# 列出所有可用的过滤器字段
tdl chat ls -f -
# 列出所有名称包含 "Telegram" 的频道
tdl chat ls -f "Type contains 'channel' && VisibleName contains 'Telegram'"
# 列出所有设置了话题功能的群组
tdl chat ls -f "len(Topics)>0"
```

- 导出 JSON 文件，可用于 `tdl` 下载

```shell
# 将导出会话中的所有媒体文件
# CHAT_INPUT 可接受例子: `@iyear`, `iyear`, `123456789`(会话 ID), `https://t.me/iyear`, `+1 123456789`

# 导出所有含媒体文件的消息
tdl chat export -c CHAT_INPUT

# 导出包含非媒体文件的所有消息
tdl chat export -c CHAT_INPUT --all

# 导出 Telegram MTProto 原生消息结构，可用于调试
tdl chat export -c CHAT_INPUT --raw

# 从指定 Topic 导出
# 你可以从以下方式获取 topic id:
# 1. 消息链接: https://t.me/c/1492447836/251011/269724(251011 为 topic id)
# 2. `tdl chat ls` 命令
tdl chat export -c CHAT_INPUT --topic TOPIC_ID

# 从指定频道文章的讨论区导出
tdl chat export -c CHAT_INPUT --reply MSG_ID

# 导出指定时间范围内的消息
tdl chat export -c CHAT_INPUT -i 1665700000,1665761624
# 或
tdl chat export -c CHAT_INPUT -T time -i 1665700000,1665761624
# 导出指定消息 ID 范围内的消息
tdl chat export -c CHAT_INPUT -T id -i 100,500
# 导出最近 N 条消息(计数受过滤器影响)
tdl chat export -c CHAT_INPUT -T last -i 100 

# 使用由表达式引擎提供的过滤器，默认为 `true`（即匹配所有）
# 如果你对表达式引擎有任何问题，请发起新的 ISSUE
# 表达式引擎文档: https://expr.medv.io/docs/Language-Definition

# 列出所有可用的过滤器字段
tdl chat export -c CHAT_INPUT -f -
# 匹配所有 zip 文件，大小 > 5MiB，且消息浏览量 > 200 的最近 10 条消息
tdl chat export -c CHAT_INPUT -T last -i 10 -f "Views>200 && Media.Name endsWith '.zip' && Media.Size > 5*1024*1024"

# 指定输出文件路径，默认为 `tdl-export.json`
tdl chat export -c CHAT_INPUT -o /path/to/output.json

# 同时导出消息内容
tdl chat export -c CHAT_INPUT --with-content
```

## 环境变量

可以通过设置环境变量来避免每次都输入相同的参数值。

**注意：所有环境变量的值都比命令行参数的优先级低。**

命令行参数含义: [flags](docs/command/tdl.md#options)

|         环境变量          |         命令行参数         |
|:---------------------:|:---------------------:|
|        TDL_NS         |       `-n/--ns`       |
|       TDL_PROXY       |       `--proxy`       |
|       TDL_DEBUG       |       `--debug`       |
|       TDL_SIZE        |      `-s/--size`      |
|      TDL_THREADS      |    `-t/--threads`     |
|       TDL_LIMIT       |     `-l/--limit`      |
|        TDL_NTP        |        `--ntp`        |
| TDL_RECONNECT_TIMEOUT | `--reconnect-timeout` |
|     TDL_TEMPLATE      |    dl `--template`    |

## 数据

你的账号数据会被存储在 `~/.tdl` 目录下。

日志文件会被存储在 `~/.tdl/log` 目录下。

## 命令

前往 [docs](docs/command/tdl.md) 查看完整的命令文档。

## 最佳实践

如何将封禁的风险降至最低？

- 导入官方客户端会话登录。
- 使用默认的下载和上传参数。不要设置过大的 `threads` 和 `size`。
- 不要在多个设备同时登录同一个账号。
- 不要短时间内下载或上传大量文件。
- 成为 Telegram 会员。😅

## 疑难解答

**Q: 为什么输入命令后没有任何反应？为什么日志中有 'msg_id too high' 的错误？**

A: 检查是否需要使用代理（使用 `proxy` 参数）；检查系统的本地时间是否正确（使用 `ntp` 参数或校准系统时间）

如果都没有用，使用 `--debug` 参数再次运行，然后提交一个 issue 并将日志粘贴到 issue 中。

**Q: Telegram 桌面客户端在使用 tdl 后无法正常工作？**

A: If your desktop client can't receive messages, load chats, or send messages, you may encounter session conflicts.

A: 如果桌面客户端无法接收消息、加载聊天或发送消息，那么可能是会话冲突导致的。

你可以尝试使用 `tdl` 重新登录，并在 ”logout“ 部分选择 `YES`，这将分离 `tdl` 和桌面客户端的会话。

**Q: 如何将会话迁移到另一台设备？**

A: 你可以使用 `tdl backup` 和 `tdl recover` 命令来导出和导入会话。更多细节请参阅 [迁移](#迁移) 部分。

## FAQ

**Q: 这是一种滥用行为吗？**

A: 不是。下载和上传速度受服务器端限制。由于官方客户端的下载速度通常不会达到最高限制，所以开发了这个工具来实现最高速度的下载。

**Q: 这会导致封禁吗？**

A: 不确定。所有操作都不涉及敏感的行为，例如主动向其他人发送消息。但是，使用长期使用的帐户进行下载和上传操作更安全。

## LICENSE

AGPL-3.0 License
