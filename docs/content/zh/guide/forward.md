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
表达式必须返回一个**字符串**或者**结构体**作为目标 CHAT，空字符串表示转发到 `收藏夹`。
{{< /hint >}}

{{< command >}}
tdl forward --from tdl-export.json \
--to 'Message.Message contains "foo" ? "CHAT1" : ""'
{{< /command >}}

转发含有 `foo` 的消息到 `CHAT1`，否则转发到 `CHAT2` 中 ID 为 4 的消息/主题：

{{< command >}}
tdl forward --from tdl-export.json \
--to 'Message.Message contains "foo" ? "CHAT1" : { Peer: "CHAT2", Thread: 4 }'
{{< /command >}}

如果表达式较复杂，你可以传递文件名：

{{< details "router.txt" >}}
你可以像写 `switch` 一样编写表达式：

```javascript
Message.Message contains "foo" ? "CHAT1" :
From.ID == 123456 ? "CHAT2" :
Message.Views > 30 ? { Peer: "CHAT3", Thread: 101 } :
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

## 编辑

使用[表达式引擎](/reference/expr)编辑转发前的消息。

{{< hint info >}}
- 你必须传递合并照片的第一条消息才能编辑标题。
- 你可以传递任何合并文档的消息以编辑相应的评论。
{{< /hint >}}

你可以在表达式中引用原始消息的相关字段。

列出所有可用字段：
{{< command >}}
tdl forward --from tdl-export.json --edit -
{{< /command >}}

在原始消息后附加 `测试转发消息`：
{{< command >}}
tdl forward --from tdl-export.json --edit 'Message.Message + " 测试转发消息"'
{{< /command >}}

以[HTML](https://core.telegram.org/bots/api#html-style)格式编写带有样式的消息：
{{< command >}}
tdl forward --from tdl-export.json --edit \
'Message.Message + `<b>粗体</b> <a href="https://example.com">链接</a>`'
{{< /command >}}

如果表达式较复杂，可以传递文件名：

{{< details "edit.txt" >}}
```javascript
repeat(Message.Message, 2) + `
<a href="https://www.google.com">谷歌</a>
<a href="https://www.bing.com">必应</a>
<b>粗体</b>
<i>斜体</i>
<code>代码</code>
<tg-spoiler>剧透</tg-spoiler>
<pre><code class="language-go">
package main

import "fmt"

func main() {
    fmt.Println("hello world")
}
</code></pre>
` + From.VisibleName
```
{{< /details >}}

{{< command >}}
tdl forward --from tdl-export.json --edit edit.txt
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

## 取消分组检测

默认情况下，tdl 将自动探测到分组消息并将它们转发为合并的消息。

你可以通过 `--single` 禁用此行为，将其作为单个消息转发。

{{< command >}}
tdl forward --from tdl-export.json --single
{{< /command >}}

## 反序

对每个来源的消息进行反序转发。

{{< command >}}
tdl forward --from tdl-export.json --desc
{{< /command >}}
