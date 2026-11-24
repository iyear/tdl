---
title: "上传"
weight: 40
---

# 上传

## 上传文件

上传指定的文件和目录到 `保存的消息`：

{{< command >}}
tdl up -p /path/to/file -p /path/to/dir
{{< /command >}}

## 分组上传

使用 `--group` 将文件发成相册。每组包含 2 到 10 个连续文件，组内文件的发送类型需要兼容，并且发往同一聊天和话题。只有一个文件时会单独发送。

照片和视频可以混在同一个相册里。文档和音频各自只与同类文件分组。要把支持的图片作为照片与视频一起发到同一个相册，请同时传入 `--photo`。

{{< command >}}
tdl up --group --photo -p cover.jpg -p clip.mp4
{{< /command >}}

每个相册使用组内第一个文件的 `--caption` 说明文字。单独发送的文件使用自己的说明文字。

同时使用 `--group` 和 `--rm` 时，Telegram 确认消息或相册发送成功后，程序才会删除对应的本地文件。

## 自定义目标

上传到自定义聊天。

{{< include "snippets/chat.md" >}}

## 指定聊天

上传到指定的聊天：

{{< command >}}
tdl up -p /path/to/file -c CHAT
{{< /command >}}

上传到论坛型聊天的指定主题：

{{< command >}}
tdl up -p /path/to/file -c CHAT --topic TOPIC_ID
{{< /command >}}

## 消息路由

通过基于[表达式](/reference/expr)的消息路由，将文件上传到不同的聊天：

{{< hint warning >}}
`--to` 标志与 `-c/--chat` 和 `--topic` 标志冲突，只能使用其中一个。
{{< /hint >}}

列出所有可用字段：

{{< command >}}
tdl up -p /path/to/file --to -
{{< /command >}}

如果 MIME 包含 `video` 则上传到 `CHAT1`，否则上传到 `收藏夹`：

{{< hint info >}}
必须返回一个字符串或结构体作为目标聊天，空字符串表示上传到 `收藏夹`。
{{< /hint >}}

{{< command >}}
tdl up -p /path/to/file \
--to 'MIME contains "video" ? "CHAT1" : ""'
{{< /command >}}

如果 MIME 包含 `video` 则上传到 `CHAT1`，否则回复 `CHAT2` 的消息/主题 `4`：

{{< command >}}
tdl up -p /path/to/file \
--to 'MIME contains "video" ? "CHAT1" : { Peer: "CHAT2", Thread: 4 }'
{{< /command >}}

如果表达式较复杂，可以传递文件名：

{{< details "router.txt" >}}
像使用 `switch` 一样编写表达式：

```javascript
MIME contains "video" ? "CHAT1" :
FileExt contains ".mp3" ? "CHAT2" :
FileName contains "chat3" > 30 ? {Peer: "CHAT3", Thread: 101} :
""
```

{{< /details >}}

{{< command >}}
tdl up -p /path/to/file --to router.txt
{{< /command >}}

## 自定义参数

使用每个任务8个线程、4个并发任务上传：

{{< command >}}
tdl up -p /path/to/file -t 8 -l 4
{{< /command >}}

## 自定义标题

使用[表达式引擎](/reference/expr)编写自定义标题。

列出所有可用字段：

{{< command >}}
tdl up -p /path/to/file --caption -
{{< /command >}}

自定义简单的标题：
{{< command >}}
tdl up -p ./path/to/file --caption 'FileName + " - uploaded by tdl"'
{{< /command >}}

以[HTML](https://core.telegram.org/bots/api#html-style)格式编写带有样式的消息：
{{< command >}}
tdl up -p /path/to/file --caption  \
'FileName + `<b>Bold</b> <a href="https://example.com">Link</a>`'
{{< /command >}}

如果表达式较复杂，可以传递文件名：

{{< details "caption.txt" >}}

```javascript
repeat(FileName, 2) + `
<a href="https://www.google.com">Google</a>
<a href="https://www.bing.com">Bing</a>
<b>bold</b>
<i>italic</i>
<code>code</code>
<tg-spoiler>spoiler</tg-spoiler>
<pre><code class="language-go">
package main

import "fmt"

func main() {
    fmt.Println("hello world")
}
</code></pre>
` + MIME
```

{{< /details >}}

{{< command >}}
tdl up -p /path/to/file --caption caption.txt
{{< /command >}}

## 过滤器

使用扩展名过滤器上传文件：

{{< hint warning >}}
扩展名仅与文件名匹配，而不是 MIME 类型。因此，这可能不会按预期工作。

白名单和黑名单不能同时使用。
{{< /hint >}}

白名单：只上传扩展名为 `.jpg` `.png` 的文件

{{< command >}}
tdl up -p /path/to/file -p /path/to/dir -i jpg,png
{{< /command >}}

黑名单：上传除了扩展名为 `.mp4` `.flv` 的所有文件

{{< command >}}
tdl up -p /path/to/file -p /path/to/dir -e mp4 -e flv
{{< /command >}}

## 自动删除

删除已上传成功的文件：

{{< command >}}
tdl up -p /path/to/file --rm
{{< /command >}}

## 照片

将图像作为照片而不是文件上传：

{{< command >}}
tdl up -p /path/to/file --photo
{{< /command >}}
