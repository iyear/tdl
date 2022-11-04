## Intro

![](https://img.shields.io/github/go-mod/go-version/iyear/tdl?style=flat-square)
![](https://img.shields.io/github/license/iyear/tdl?style=flat-square)
![](https://img.shields.io/github/workflow/status/iyear/tdl/master%20builder?style=flat-square)
![](https://img.shields.io/github/v/release/iyear/tdl?color=red&style=flat-square)
![](https://img.shields.io/github/last-commit/iyear/tdl?style=flat-square)

📥 Telegram Downloader, but more than a downloader 🚀

> ⚠ Note: Command compatibility is not guaranteed in the early stages of development

> Improvements have been made to the risk of blocking, but it still can't be completely avoided. Go to [Discussion](https://github.com/iyear/tdl/discussions/29) for more information.

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
```shell
# Linux & macOS
sudo mv tdl /usr/local/bin
# Windows (PowerShell)
Move-Item tdl.exe C:\Windows\System32
```

## Quick Start

```shell
# login with existing official desktop clients (recommended)
tdl login -n quickstart -d /path/to/Desktop-Telegram-Client
# or login with phone & code
tdl login -n quickstart

tdl dl -n quickstart -u https://t.me/telegram/193
```

## Usage

Get help

```shell
tdl -h
```

Check the version

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

(optional) Set the proxy. Only support socks now:

```shell
tdl --proxy socks5://localhost:1080
# or
export TDL_PROXY=socks5://localhost:1080 # recommended
```

(optional) Set ntp server host. If is empty, use system time:

```shell
tdl --ntp pool.ntp.org
# or
export TDL_NTP=pool.ntp.org # recommended
```

### Login

> When you first use tdl, you need to login to get a Telegram session

If you have official desktop clients locally, you can import existing sessions.

This may reduce the risk of blocking, but is unproven:

```shell
tdl login -d /path/to/Telegram # recommended
```

Login to Telegram with phone & code:

```shell
tdl login
```

### Download

> Please do not arbitrarily set too large `threads` and `size`.
>
> **The default value of options is consistent with official clients to reduce the risk of blocking.**
>
> If you need higher speed, set higher threads and size
> 
> For details: https://github.com/iyear/tdl/issues/30

Advanced Options:

|      Flag      |                      Default                       |                 Desc                  |
|:--------------:|:--------------------------------------------------:|:-------------------------------------:|
| `-t/--threads` |                         4                          |     threads for transfer one item     |
|  `-s/--size`   |                   128*1024 Bytes                   | part size for transfer, max is 512KiB |
|  `-l/--limit`  |                         2                          |    max number of concurrent tasks     |
|  `--template`  | `{{ .DialogID }}_{{ .MessageID }}_{{ .FileName }}` |          file name template           |

Download (protected) chat files from message urls:

```shell
tdl dl -u https://t.me/tdl/1 -u https://t.me/tdl/2
```

Download (protected) chat files from [official desktop client exported JSON](docs/desktop_export.md):

```shell
tdl dl -f result1.json -f result2.json
```

You can combine sources:

```shell
tdl dl -u https://t.me/tdl/1 -u https://t.me/tdl/2 -f result1.json -f result2.json
```

Download with 8 threads, 512KiB(MAX) part size, 4 concurrent tasks:

```shell
tdl dl -u https://t.me/tdl/1 -t 8 -s 524288 -l 4
```

Download with custom file name template:

Following the [go template syntax](https://pkg.go.dev/text/template), you can use the variables:

|     Var      |                 Desc                 |
|:------------:|:------------------------------------:|
|   DialogID   |          Telegram dialog id          |
|  MessageID   |         Telegram message id          |
| MessageDate  |      Telegram message date(ts)       |
|   FileName   |          Telegram file name          |
|   FileSize   | Human-readable file size, like `1GB` |
| DownloadDate |          Download date(ts)           |

```shell
tdl dl -u https://t.me/tdl/1 --template "{{ .DialogID }}_{{ .MessageID }}_{{ .DownloadDate }}_{{ .FileName }}"
```

Full examples:
```shell
tdl dl --debug --ntp pool.ntp.org -n iyear --proxy socks5://localhost:1080 -u https://t.me/tdl/1 -u https://t.me/tdl/2 -f result1.json -f result2.json -t 8 -s 262144 -l 4
```

### Upload

> Same instructions and advanced options as **Download**

Upload files to `Saved Messages`, exclude the specified file extensions:

```shell
tdl up -p /path/to/file -p /path/to/dir -e .so -e .tmp
```

Upload with 8 threads, 512KiB(MAX) part size, 4 concurrent tasks:

```shell
tdl up -p /path/to/file -t 8 -s 524288 -l 4
```

Full examples:
```shell
tdl up --debug --ntp pool.ntp.org -n iyear --proxy socks5://localhost:1080 -p /path/to/file -p /path/to/dir -e .so -e .tmp -t 8 -s 262144 -l 4
```

### Backup

> Backup or recover your data

Backup (Default: `tdl-backup-<time>.zip`):

```shell
tdl backup
# or specify the backup file path
tdl backup -d /path/to/backup.zip
```

Recover:

```shell
tdl recover -f /path/to/backup.zip
```

### Chat Utilities

> Some useful utils

List all your chats:

```shell
tdl chat ls
```

Export minimal JSON for tdl download (NOT for backup):

```shell
# will export all media files in the chat.
# chat input examples: `@iyear`, `iyear`, `123456789`(chat id), `https://t.me/iyear`, `+1 123456789`

tdl chat export -c CHAT_INPUT

# specify the time period with timestamp format, default is start from 1970-01-01, end to now
tdl chat export -c CHAT_INPUT --from 1665700000 --to 1665761624
# or (timestamp is default format)
tdl chat export -c CHAT_INPUT --from 1665700000 --to 1665761624 --time

# specify with message id format, default is start from 0, end to latest message
tdl chat export -c CHAT_INPUT --from 100 --to 500 --msg

# specify the output file path, default is `tdl-export.json`
tdl chat export -c CHAT_INPUT -o /path/to/output.json
```

## Env

Avoid typing the same flag values repeatedly every time by setting environment variables.

**Note: The values of all environment variables have a lower priority than flags.**

What flags mean: [flags](docs/command/tdl.md#options)

|    NAME     |      FLAG      |
|:-----------:|:--------------:|
|   TDL_NS    |   `-n/--ns`    |
|  TDL_PROXY  |   `--proxy`    |
|  TDL_DEBUG  |   `--debug`    |
|  TDL_SIZE   |  `-s/--size`   |
| TDL_THREADS | `-t/--threads` |
|  TDL_LIMIT  |  `-l/--limit`  |
|   TDL_NTP   |    `--ntp`     |

## Data

Your account information will be stored in the `~/.tdl` directory.

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

## LICENSE

AGPL-3.0 License
