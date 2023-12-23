---
title: "Global Config"
weight: 10
---

# Global Config

Global config is some CLI flags that can be set in every command.

## `-n/--ns`

Each namespace represents a Telegram account. Default: `default`.

If you want to add another account, just add `-n YOUR_ACCOUNT_NAME` option to every command:

{{< command >}}
tdl -n iyear
{{< /command >}}

## `--proxy`

Set the proxy. Default: `""`.

Format: `protocol://username:password@host:port`

{{< command >}}
tdl --proxy socks5://localhost:1080
tdl --proxy http://localhost:8080
tdl --proxy https://localhost:8081
{{< /command >}}

## `--storage`

Set the storage. Default: `type=bolt,path=~/.tdl/data`

Format: `type=DRIVER,opt1=val1,opt2=val2,...`

Available drivers:

|      Driver      |            Options             | Description                                                                                                   |
|:----------------:|:------------------------------:|---------------------------------------------------------------------------------------------------------------|
| `bolt` (Default) | `path=/path/to/data-directory` | Store data in separate database files. So you can use it in multiple processes(must be different namespaces). |
|      `file`      |   `path=/path/to/data.json`    | Store data in a single JSON file, which is useful for debugging.                                              |
|     `legacy`     |    `path=/path/to/data.kv`     | **Deprecated.** Store data in a single database file. So you **can't** use it in multiple processes.          |
|        -         |               -                | Wait for more drivers...                                                                                      |

{{< command >}}
tdl --storage type=bolt,path=/path/to/data-dir
{{< /command >}}

## `--ntp`

Set ntp server host. If it's empty, system time will be used. Default: `""`.

{{< command >}}
tdl --ntp pool.ntp.org
{{< /command >}}

## `--reconnect-timeout`

Set Telegram client reconnect timeout. Default: `2m`.

{{< hint info >}}
Set higher timeout or 0(INF) if your network is poor.
{{< /hint >}}

{{< command >}}
tdl --reconnect-timeout 1m30s
{{< /command >}}

## `--debug`

Enable debug level log. Default: `false`.

{{< command >}}
tdl --debug
{{< /command >}}

## `--pool`

Set the DC pool size of Telegram client. Default: `8`.

{{< hint info >}}
Set higher timeout or 0(INF) if you want faster speed.
{{< /hint >}}

{{< command >}}
tdl --pool 2
{{< /command >}}
