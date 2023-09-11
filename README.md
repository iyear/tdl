# tdl

English | [简体中文](README_zh.md)

![](https://img.shields.io/github/go-mod/go-version/iyear/tdl?style=flat-square)
![](https://img.shields.io/github/license/iyear/tdl?style=flat-square)
![](https://img.shields.io/github/actions/workflow/status/iyear/tdl/master.yml?branch=master&style=flat-square)
![](https://img.shields.io/github/v/release/iyear/tdl?color=red&style=flat-square)
![](https://img.shields.io/github/downloads/iyear/tdl/total?style=flat-square)

📥 Telegram Downloader, but more than a downloader

## TOC

* [Features](#features)
* [Preview](#preview)
* [Install](#install)
* [Quick Start](#quick-start)
* [Workflows](#workflows)
* [Usage](#usage)
   * [Basic Configs](#basic-configs)
   * [Login](#login)
   * [Download](#download)
   * [Upload](#upload)
   * [Migration](#migration)
   * [Chat Utilities](#chat-utilities)
* [Env](#env)
* [Data](#data)
* [Commands](#commands)
* [Best Practice](#best-practice)
* [Troubleshooting](#troubleshooting)
* [FAQ](#faq)

## Features

- Single file start-up
- Low resource usage
- Take up all your bandwidth
- Faster than official clients
- Download files from (protected) chats
- Upload files to Telegram
- Export messages/members/subscribers to JSON

## Preview

It reaches my proxy's speed limit, and the **speed depends on whether you are a premium**

![](img/preview.gif)

## Install

You can download prebuilt binaries from [releases](https://github.com/iyear/tdl/releases/latest) or install with below methods:

### Linux & macOS

<details>

- Install with Shell:

`tdl` will be installed to `/usr/local/bin/tdl`, and script also can be used to upgrade `tdl`.

```shell
# Install latest version
curl -sSL https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.sh | sudo bash
```

```shell
# Use ghproxy.com to speed up download
curl -sSL https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.sh | sudo bash -s -- --proxy
# Install specific version
curl -sSL https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.sh | sudo bash -s -- --version VERSION
```

- Install with package managers:

Make your contribution to package managers: [File an issue](https://github.com/iyear/tdl/issues/new/choose)

</details>

### Windows

<details>

- Install with PowerShell(Administrator):

`tdl` will be installed to `$Env:SystemDrive\tdl`(will be added to `PATH`), and script also can be used to upgrade `tdl`.

```powershell
# Install latest version
iwr -useb https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.ps1 | iex
```

```powershell
# Use `ghproxy.com` to speed up download
$Script=iwr -useb https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.ps1; $Block=[ScriptBlock]::Create($Script); Invoke-Command -ScriptBlock $Block -ArgumentList "", "$True"
# Install specific version
$Env:TDLVersion = "VERSION"
$Script=iwr -useb https://ghproxy.com/https://raw.githubusercontent.com/iyear/tdl/master/scripts/install.ps1; $Block=[ScriptBlock]::Create($Script); Invoke-Command -ScriptBlock $Block -ArgumentList "$Env:TDLVersion"
```

- Install with package managers:

```powershell
# Scoop (Windows) https://scoop.sh/#/apps?s=2&d=1&o=true&p=1&q=telegram+downloader
scoop bucket add extras
scoop install telegram-downloader
```

</details>

## Quick Start

```shell
# login with existing official desktop clients (recommended)
tdl login -n quickstart
# if you set a local passcode
tdl login -n quickstart -p YOUR_PASSCODE
# specify custom path
tdl login -n quickstart -d /path/to/TelegramDesktop
# or login with phone & code
tdl login -n quickstart --code

tdl dl -n quickstart -u https://t.me/telegram/193
```

## Workflows

<details>

**Only show workflows, not all configs. So read [Usage](#usage) and set configs you need.**

### Download files from message urls

```shell
export TDL_NS=iyear # set our namespace
tdl login
tdl dl -u https://t.me/tdl/1 -u https://t.me/tdl/2
```

### Download files from protected chats

```shell
export TDL_NS=iyear # set our namespace
tdl login
tdl chat export -o result.json
tdl dl -f result.json
```

### Migrate data to remote server

```shell
export TDL_NS=iyear # set our namespace
tdl login
tdl backup -d backup.zip
# upload backup.zip to remote server
tdl recover -f backup.zip # run on remote server
```

### Continuously downloading regardless of errors

It is recommended to use daemon program + `tdl` download, some errors may require reboot `tdl` to work properly.

`tdl` is not responsible for daemon, you can choose a daemon program for different platforms, for example systemd for Linux.

Command: `tdl dl <OTHER_FLAGS> --continue`

This way `tdl` will be restarted in case of errors and will continue downloading task sources.

</details>

## Usage

- Get help

```shell
tdl -h
```

- Check the version

```shell
tdl version
```

- Shell completion

Run corresponding command to enable shell completion in all sessions:

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

### Basic Configs

> The following command documents will not write basic configs. Please add the basic configs you need.

Each namespace represents a Telegram account

You should set the namespace **when each command is executed**:

```shell
tdl -n iyear
# or
export TDL_NS=iyear # recommended
```

- (optional) Set the proxy. Only support socks now:

```shell
tdl --proxy socks5://localhost:1080
# or
export TDL_PROXY=socks5://localhost:1080 # recommended
```

- (optional) Set ntp server host. If is empty, use system time:

```shell
tdl --ntp pool.ntp.org
# or
export TDL_NTP=pool.ntp.org # recommended
```

- (optional) Set Telegram client reconnect timeout. Default is 2m:

> **Note**
> Set higher timeout or 0(INF) if your network is poor.

```shell
tdl --reconnect-timeout 1m30s
# or
export TDL_RECONNECT_TIMEOUT=1m30s
```

### Login

> When you first use tdl, you need to login to get a Telegram session

- If you have [official desktop clients](https://desktop.telegram.org/) locally, you can import existing sessions.

This may reduce the risk of blocking, but is unproven:

```shell
tdl login
# if you set a local passcode
tdl login -p YOUR_PASSCODE
#  specify custom path
tdl login -d /path/to/TelegramDesktop
```

- Login to Telegram with phone & code:

```shell
tdl login --code
```

### Download

> If you need higher speed, set higher threads. But do not arbitrarily set too large `threads`.

- Download (protected) chat files from message urls:

```shell
tdl dl -u https://t.me/tdl/1 -u https://t.me/tdl/2
```

- Download (protected) chat files from [official desktop client exported JSON](docs/desktop_export.md):

```shell
tdl dl -f result1.json -f result2.json
```

- You can combine sources:

```shell
tdl dl \
-u https://t.me/tdl/1 -u https://t.me/tdl/2 \
-f result1.json -f result2.json
```

- Download with 8 threads, 512KiB(MAX) part size, 4 concurrent tasks:

```shell
tdl dl -u https://t.me/tdl/1 -t 8 -s 524288 -l 4
```

- Download with real extension according to MIME type:

> **Note**
> If the file extension is not matched with the MIME type, tdl will rename the file with the correct extension.
> 
> Side effect: like `.apk` file, it will be renamed to `.zip`.

```shell
tdl dl -u https://t.me/tdl/1 --rewrite-ext
```

- Skip the same files when downloading:

> **Note**
> IF: file name(without extension) and size is the same

```shell
tdl dl -u https://t.me/tdl/1 --skip-same
```

- Download files to custom directory:

```shell
tdl dl -u https://t.me/tdl/1 -d /path/to/dir
```

- Download files with custom order:

> **Note**
> Different order will affect resuming download

```shell
# download files in descending order(from newest to oldest)
tdl dl -f result.json --desc
# Default is ascending order
tdl dl -f result.json
```

- Download files with [takeout session](https://arabic-telethon.readthedocs.io/en/stable/extra/examples/telegram-client.html#exporting-messages):

> **Note**
> If you plan to download a lot of media, you may prefer to do this within a takeout session. Takeout sessions let you export data from your account with lower flood wait limits.

```shell
tdl dl -u https://t.me/tdl/1 --takeout
```

- Download files with extension filters:

> **Note**
> The extension is only matched with the file name, not the MIME type. So it may not work as expected.
> 
> Whitelist and blacklist can not be used at the same time.

```shell
# whitelist filter, only download files with `.jpg` `.png` extension
tdl dl -u https://t.me/tdl/1 -i jpg,png

# blacklist filter, download all files except `.mp4` `.flv` extension
tdl dl -u https://t.me/tdl/1 -e mp4,flv
```

- Download with custom file name template:

Please refer to [template guide](docs/template.md) for more details.

```shell
tdl dl -u https://t.me/tdl/1 \
--template "{{ .DialogID }}_{{ .MessageID }}_{{ .DownloadDate }}_{{ .FileName }}"
```

- Resume or restart download without UI interaction:

```shell
# resume
tdl dl -u https://t.me/tdl/1 --continue
# restart
tdl dl -u https://t.me/tdl/1 --restart
```

- Full example:
```shell
tdl dl --debug --ntp pool.ntp.org \
-n iyear --proxy socks5://localhost:1080 \
-u https://t.me/tdl/1 -u https://t.me/tdl/2 \
-f result1.json -f result2.json \
--rewrite-ext --skip-same -i jpg,png \
-d /path/to/dir --desc \
-t 8 -s 262144 -l 4
```

### Upload

> Same instructions and advanced options as **Download**

- Upload files to `Saved Messages`, exclude the specified file extensions:

```shell
tdl up -p /path/to/file -p /path/to/dir -e .so -e .tmp
```

- Upload with 8 threads, 512KiB(MAX) part size, 4 concurrent tasks:

```shell
tdl up -p /path/to/file -t 8 -s 524288 -l 4
```

- Delete the uploaded file after successful upload:

```shell
tdl up -p /path/to/file --rm
```

- Upload images as photos:

```shell
tdl up -p /path/to/file --photo
```

- Upload to custom chat:

```shell
# chat input examples: `@iyear`, `iyear`, `123456789`(chat id), `https://t.me/iyear`, `+1 123456789`

# empty chat means `Saved Messages`
tdl up -p /path/to/file -c CHAT_INPUT
```

- Full example:
```shell
tdl up --debug --ntp pool.ntp.org \
-n iyear --proxy socks5://localhost:1080 \
-p /path/to/file -p /path/to/dir \
-e .so -e .tmp \
-t 8 -s 262144 -l 4
-c @iyear
```

### Migration

> Backup or recover your data

- Backup (Default: `tdl-backup-<time>.zip`):

```shell
tdl backup
# or specify the backup file path
tdl backup -d /path/to/backup.zip
```

- Recover:

```shell
tdl recover -f /path/to/backup.zip
```

### Chat Utilities

- List all your chats:

```shell
tdl chat ls

# output with JSON format
tdl chat ls -o json

# specify filter that powered by expression engine, default is `true`(match all)
# feel free to file an issue if you have any questions about the expression engine.
# expression engine docs: https://expr.medv.io/docs/Language-Definition

# list all available filter fields
tdl chat ls -f -
# list channels that VisibleName contains "Telegram"
tdl chat ls -f "Type contains 'channel' && VisibleName contains 'Telegram'"
# list groups that have topics
tdl chat ls -f "len(Topics)>0"
```

- Export chat members/subscribers, admins, bots, etc:

> **Note**
> Chat admin required

```shell
# chat input examples: `@iyear`, `iyear`, `123456789`(chat id), `https://t.me/iyear`, `+1 123456789`

# export all users to tdl-users.json
tdl chat users -c CHAT_INPUT
# export with specified path
tdl chat users -c CHAT_INPUT -o /path/to/export.json
# export Telegram MTProto raw user structure, useful for debugging
tdl chat users -c CHAT_INPUT --raw
```

- Export JSON for `tdl` download:

```shell
# will export all media files in the chat.
# chat input examples: `@iyear`, `iyear`, `123456789`(chat id), `https://t.me/iyear`, `+1 123456789`

# export all media messages
tdl chat export -c CHAT_INPUT

# export all messages including non-media messages
tdl chat export -c CHAT_INPUT --all

# export Telegram MTProto raw message structure, useful for debugging
tdl chat export -c CHAT_INPUT --raw

# export from specific topic
# You can get topic id from:
# 1. message link: https://t.me/c/1492447836/251011/269724(251011 is topic id)
# 2. `tdl chat ls` command
tdl chat export -c CHAT_INPUT --topic TOPIC_ID
# export from specific channel post replies
tdl chat export -c CHAT_INPUT --reply MSG_ID

# export with specific timestamp range, default is start from 1970-01-01, end to now
tdl chat export -c CHAT_INPUT -i 1665700000,1665761624
# or (time is default type)
tdl chat export -c CHAT_INPUT -T time -i 1665700000,1665761624
# export with specific message id range, default to start from 0, end to latest message
tdl chat export -c CHAT_INPUT -T id -i 100,500
# export last N media files
tdl chat export -c CHAT_INPUT -T last -i 100 

# specify filter that powered by expression engine, default is `true`(match all)
# feel free to file an issue if you have any questions about the expression engine.
# expression engine docs: https://expr.medv.io/docs/Language-Definition

# list all available filter fields
tdl chat export -c CHAT_INPUT -f -
# match last 10 zip files that size > 5MiB and views > 200
tdl chat export -c CHAT_INPUT -T last -i 10 -f "Views>200 && Media.Name endsWith '.zip' && Media.Size > 5*1024*1024"

# specify the output file path, default is `tdl-export.json`
tdl chat export -c CHAT_INPUT -o /path/to/output.json

# export with message content
tdl chat export -c CHAT_INPUT --with-content
```

## Env

Avoid typing the same flag values repeatedly every time by setting environment variables.

**Note: The values of all environment variables have a lower priority than flags.**

What flags mean: [flags](docs/command/tdl.md#options)

|         NAME          |         FLAG          |
|:---------------------:|:---------------------:|
|        TDL_NS         |       `-n/--ns`       |
|       TDL_PROXY       |       `--proxy`       |
|       TDL_DEBUG       |       `--debug`       |
|       TDL_SIZE        |      `-s/--size`      |
|      TDL_THREADS      |    `-t/--threads`     |
|       TDL_LIMIT       |     `-l/--limit`      |
|       TDL_POOL        |       `--pool`        |
|        TDL_NTP        |        `--ntp`        |
| TDL_RECONNECT_TIMEOUT | `--reconnect-timeout` |
|     TDL_TEMPLATE      |    dl `--template`    |

## Data

Your account information will be stored in the `~/.tdl` directory.

Log files will be stored in the `~/.tdl/log` directory.

## Commands

Go to [docs](docs/command/tdl.md) for full command docs.

## Best Practice
How to minimize the risk of blocking?

- Login with the official client session.
- Use the default download and upload options as possible. Do not set too large `threads` and `size`.
- Do not use the same account to login on multiple devices at the same time.
- Don't download or upload too many files at once.
- Become a Telegram premium user. 😅

## Troubleshooting

**Q: Why no response after entering the command? And why there is 'msg_id too high' in the log?**

A: Check if you need to use a proxy (use `proxy` flag); Check if your system's local time is correct (use `ntp` flag or calibrate system time)

If that doesn't work, run again with `--debug` flag. Then file a new issue and paste your log in the issue.

**Q: Desktop client stop working after using tdl?**

A: If your desktop client can't receive messages, load chats, or send messages, you may encounter session conflicts.

You can try re-login with desktop client and **select YES for logout**, which will delete the session files to separate sessions.

**Q: How to migrate session to another device?**

A: You can use the `tdl backup` and `tdl recover` commands to export and import sessions. See [Migration](#migration) for more details.

## FAQ

**Q: Is this a form of abuse?**

A: No. The download and upload speed is limited by the server side. Since the speed of official clients usually does not
reach the account limit, this tool was developed to download files at the highest possible speed.

**Q: Will this result in a ban?**

A: I am not sure. All operations do not involve dangerous actions such as actively sending messages to other people. But
it's safer to use a long-term account.

## LICENSE

AGPL-3.0 License
