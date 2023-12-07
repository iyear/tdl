---
title: "转发"
weight: 35
---

# 转发

具有自动回退和消息路由的转发功能

一行命令将消息从 `https://t.me/telegram/193` 转发到 `收藏夹`：

{{< command >}}
tdl forward --from https://t.me/telegram/193
{{< /command >}}

## 自定义来源

{{< include "snippets/link.md" >}}

您可以从链接和[导出的JSON文件](/zh/guide/download/#从-json-下载)转发消息：

{{< command >}}
tdl forward \
--from https://t.me/telegram/193 \
--from https://t.me/telegram/195 \
--from tdl-export.json \
--from tdl-export2.json
{{< /command >}}

## 自定义目标

{{< include "snippets/chat.md" >}}

### 特定聊天

转发到特定的聊天：

{{< command >}}
tdl forward --from tdl-export.json --to CHAT
{{< /command >}}

### 消息路由

通过基于 [expr](/zh/reference/expr) 的路由将消息转发至不同的聊天

列出所有可用的字段：

{{< command >}}
tdl forward --from tdl-export.json --to -
{{< /command >}}

如果消息包含 `foo`，则转发到 `CHAT1`，否则转发到 `收藏夹`：

{{< hint info >}}
表达式必须返回一个字符串作为目标 CHAT，空字符串表示转发到 `收藏夹`。
{{< /hint >}}

{{< command >}}
tdl forward --from tdl-export.json \
--to 'Message.Message contains "foo" ? "CHAT1" : ""'
{{< /command >}}

如果表达式较复杂，你可以传递文件名：

{{< details "router.txt" >}}
你可以像写 `switch` 一样编写表达式：

```
Message.Message contains "foo" ? "CHAT1" :
From.ID == 123456 ? "CHAT2" :
Message.Views > 30 ? "CHAT3" :
""
```

{{< /details >}}

{{< command >}}
tdl forward --from tdl-export.json --to router.txt
{{< /command >}}

## 模式

消息转发采取自动降级策略

可用模式：
- `direct`（默认）
- `clone`

### Direct

优先使用官方的转发API。

如果聊天或消息不允许使用官方转发API，将自动降级为 `clone` 模式。

{{< command >}}
tdl forward --from tdl-export.json --mode direct
{{< /command >}}

### Clone

通过复制方式转发消息，将不包含转发来源的标头。

将自动忽略一些无法复制的消息内容，例如投票、发票等

{{< command >}}
tdl forward --from tdl-export.json --mode clone
{{< /command >}}

## 试运行

只打印进度而不实际发送消息，可以用于调试消息路由的效果。

{{< command >}}
tdl forward --from tdl-export.json --dry-run
{{< /command >}}

## 静默发送

发送消息而不通知其他成员。

{{< command >}}
tdl forward --from tdl-export.json --silent
{{< /command >}}

