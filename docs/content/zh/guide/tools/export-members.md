---
title: "导出成员"
weight: 20
---

# 导出成员

导出聊天成员/订阅者、管理员、机器人等。

{{< hint info >}}
部分类型用户(被禁用户/被踢出用户/……)导出需要聊天管理员权限。
{{< /hint >}}

{{< include "snippets/chat.md" >}}

## 默认

将所有用户导出为 `tdl-users.json`

{{< command >}}
tdl chat users -c CHAT
{{< /command >}}

## 自定义路径

指定文件路径进行导出

{{< command >}}
tdl chat users -c CHAT -o /path/to/export.json
{{< /command >}}

## 原始数据

导出 Telegram MTProto 原始用户结构，用于调试。

{{< command >}}
tdl chat users -c CHAT --raw
{{< /command >}}
