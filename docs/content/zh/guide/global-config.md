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
