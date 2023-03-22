# tdl

![](https://img.shields.io/github/go-mod/go-version/iyear/tdl?style=flat-square)
![](https://img.shields.io/github/license/iyear/tdl?style=flat-square)
![](https://img.shields.io/github/actions/workflow/status/iyear/tdl/master.yml?branch=master&style=flat-square)
![](https://img.shields.io/github/v/release/iyear/tdl?color=red&style=flat-square)
![](https://img.shields.io/github/downloads/iyear/tdl/total?style=flat-square)

📥 Telegram Downloader, but more than a downloader 🚀

## Contents

* [Features](#features)
* [Preview](#preview)
* [Install](#install)
* [Quick Start](#quick-start)
* [Usage](#usage)
   * [Basic Configs](#basic-configs)
   * [Login](#login)
   * [Download](#download)
   * [Upload](#upload)
   * [Backup](#backup)
   * [Chat Utilities](#chat-utilities)
* [Env](#env)
* [Data](#data)
* [Commands](#commands)
* [Best Practice](#best-practice)
* [FAQ](#faq)

## Features

- Single file start-up
- Low resource usage
- Take up all your bandwidth
- Faster than official clients
- Download files from (protected) chats
- Upload files to Telegram

## Preview

It reaches my proxy's speed limit, and the **speed depends on whether you are a premium**

![](img/preview.gif)

## Install

Go to [GitHub Releases](https://github.com/iyear/tdl/releases) to download the latest version

(optional) Use it everywhere:
```powershell
# Should run as root(Administrator)
# Linux & macOS
sudo mv tdl /usr/bin
# Windows (PowerShell)
Move-Item tdl.exe C:\Windows\System32
```

Install with a package manager:
```shell
# Scoop (Windows) https://scoop.sh/#/apps?s=2&d=1&o=true&p=1&q=telegram+downloader
scoop bucket add extras
scoop install telegram-downloader
```

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

## Usage

- Get help

```shell
tdl -h
```

- Check the version

```shell
tdl version
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

> Please do not arbitrarily set too large `threads` and `size`.
>
> **The default value of options is consistent with official clients to reduce the risk of blocking.**
>
> If you need higher speed, set higher threads and size
> 
> For details: https://github.com/iyear/tdl/issues/30

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

### Backup

> Backup or recover your data, often used for migrating session to remote server

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

> Some useful utils

- List all your chats:

```shell
tdl chat ls
```

- Export minimal JSON for tdl download (NOT for backup):

```shell
# will export all media files in the chat.
# chat input examples: `@iyear`, `iyear`, `123456789`(chat id), `https://t.me/iyear`, `+1 123456789`

# export all messages
tdl chat export -c CHAT_INPUT

# export with specific timestamp range, default is start from 1970-01-01, end to now
tdl chat export -c CHAT_INPUT -i 1665700000,1665761624
# or (time is default type)
tdl chat export -c CHAT_INPUT -i 1665700000,1665761624 -T time

# export with specific message id range, default is start from 0, end to latest message
tdl chat export -c CHAT_INPUT -i 100,500 -T id

# export last N media files
tdl chat export -c CHAT_INPUT -i 100 -T last

# specify the output file path, default is `tdl-export.json`
tdl chat export -c CHAT_INPUT -o /path/to/output.json
```

## Env

Avoid typing the same flag values repeatedly every time by setting environment variables.

**Note: The values of all environment variables have a lower priority than flags.**

What flags mean: [flags](docs/command/tdl.md#options)

|     NAME     |      FLAG       |
|:------------:|:---------------:|
|    TDL_NS    |    `-n/--ns`    |
|  TDL_PROXY   |    `--proxy`    |
|  TDL_DEBUG   |    `--debug`    |
|   TDL_SIZE   |   `-s/--size`   |
| TDL_THREADS  | `-t/--threads`  |
|  TDL_LIMIT   |  `-l/--limit`   |
|   TDL_NTP    |     `--ntp`     |
| TDL_TEMPLATE | dl `--template` |

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

## FAQ

**Q: Is this a form of abuse?**

A: No. The download and upload speed is limited by the server side. Since the speed of official clients usually does not
reach the account limit, this tool was developed to download files at the highest possible speed.

**Q: Will this result in a ban?**

A: I am not sure. All operations do not involve dangerous actions such as actively sending messages to other people. But
it's safer to use an unused account for download and upload operations.

**Q: No response after entering the command?**

A: Check if you need to use a proxy (use `proxy` flag); Check if your system's local time is correct (use `ntp` flag or calibrate system time)

If that doesn't work, run again with `debug` flag. Then file a new issue and paste your log in the issue.

## LICENSE

AGPL-3.0 License
