---
title: "Global Config"
weight: 10
---

# Global Config

Global config is some CLI flags that can be set in every command.

## `-n/--ns`

Each namespace represents a Telegram account.

You should set the namespace **when each command is executed**:

{{< command >}}
tdl -n iyear
{{< /command >}}

## `--proxy`

Set the proxy. Only support `socks5` now. Default: `""`.

Format: `protocol://username:password@host:port`

{{< command >}}
tdl --proxy socks5://localhost:1080
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

Set the DC pool size of Telegram client. Default: `3`.

{{< hint warning >}}
DO NOT set it too large, or tdl may be forced to quit by Telegram.
{{< /hint >}}

{{< command >}}
tdl --pool 2
{{< /command >}}
