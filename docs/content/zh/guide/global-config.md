---
title: "全局设置"
weight: 10
---
# 全局配置

全局配置是可以在每个命令中设置的选项。

## `-n/--ns`

每个命名空间代表一个 Telegram 帐号。默认值：`default`。

例如你想新增一个其他账户，为所有命令都添加 `-n YOUR_ACCOUNT_NAME` 选项即可：

{{< command >}}
tdl -n iyear
{{< /command >}}

## `--proxy`

设置代理。默认值：`""`。

格式：`protocol://username:password@host:port`

{{< command >}}
tdl --proxy socks5://localhost:1080
tdl --proxy http://localhost:8080
tdl --proxy https://localhost:8081
{{< /command >}}

## `--storage`

设置存储。默认值：`type=bolt,path=~/.tdl/data`

格式: `type=驱动,opt1=val1,opt2=val2,...`

可用的驱动：

|    驱动名     |               选项               | 描述                                          |
|:----------:|:------------------------------:|---------------------------------------------|
| `bolt`（默认） | `path=/path/to/data-directory` | 将数据存储在单独的数据库文件中，因此您可以在多个进程中使用（但必须是不同的命名空间）。 |
|   `file`   |   `path=/path/to/data.json`    | 将数据存储在单个 JSON 文件中，通常用于调试。                   |
|  `legacy`  |    `path=/path/to/data.kv`     | **已弃用。** 将数据存储在单个数据库文件中，因此你**不能**在多个进程中使用它。 |
|     -      |               -                | 等待更多驱动...                                   |

{{< command >}}
tdl --storage type=bolt,path=/path/to/data-dir
{{< /command >}}

## `--ntp`

设置 NTP 服务器。如果为空，将使用系统时间。默认值：`""`。

{{< command >}}
tdl --ntp pool.ntp.org
{{< /command >}}

## `--reconnect-timeout`

设置 Telegram 连接的重连超时。默认值：`2m`。

{{< hint info >}}
如果您的网络不稳定，请将超时设置为更长时间或0（无限）。
{{< /hint >}}

{{< command >}}
tdl --reconnect-timeout 1m30s
{{< /command >}}

## `--debug`

启用调试级别日志。默认值：`false`。

{{< command >}}
tdl --debug
{{< /command >}}

## `--pool`

设置 Telegram 客户端的连接池大小。默认值：`8`。

{{< hint info >}}
如果你想要更快的速度，请将连接池设置的更大或者0（无限）。
{{< /hint >}}

{{< command >}}
tdl --pool 2
{{< /command >}}
