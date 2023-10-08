---
title: "全局设置"
weight: 10
---
# 全局配置

全局配置是可以在每个命令中设置的选项。

## `-n/--ns`

每个命名空间代表一个 Telegram 帐号。

在执行每个命令时，您应该设置命名空间：

{{< command >}}
tdl -n iyear
{{< /command >}}

## `--proxy`

设置代理。目前仅支持 `socks5`。默认值：`""`。

格式：`protocol://username:password@host:port`

{{< command >}}
tdl --proxy socks5://localhost:1080
{{< /command >}}

## `--ntp`

设置 NTP 服务器。如果为空，将使用系统时间。默认值：`""`。

{{< command >}}
tdl --ntp pool.ntp.org
{{< /command >}}

## `--reconnect-timeout`

设置 Telegram 连接的重连超时。默认值：`2m`。

{{< hint info >}}
如果您的网络不稳定，请将超时设置为较长时间或0（无限）。
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

设置 Telegram 客户端的连接池大小。默认值：`3`。

{{< hint warning >}}
不要将其设置得过大，否则 Telegram 可能会强制断开连接。
{{< /hint >}}

{{< command >}}
tdl --pool 2
{{< /command >}}
