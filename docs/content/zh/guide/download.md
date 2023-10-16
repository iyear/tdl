---
title: "下载"
weight: 30
---

# 下载

## 从链接下载：

{{< hint info >}}
点击官方客户端的 "复制链接" 按钮获取消息链接。
{{< /hint >}}

{{< details title="消息链接示例" open=false >}}

- `https://t.me/telegram/193`
- `https://t.me/c/1697797156/151`
- `https://t.me/iFreeKnow/45662/55005`
- `https://t.me/c/1492447836/251015/251021`
- `https://t.me/opencfdchannel/4434?comment=360409`
- `https://t.me/myhostloc/1485524?thread=1485523`
- `...`（如果发现新的链接格式，请提交新的 Issue）
  {{< /details >}}

{{< command >}}
tdl dl -u https://t.me/tdl/1 -u https://t.me/tdl/2
{{< /command >}}

## 从 JSON 下载：

有两种导出 JSON 文件的方式：

{{< tabs "json" >}}
{{< tab "tdl" >}}
请参考 [导出消息](/zh/guide/tools/export-messages)
{{< /tab >}}

{{< tab "桌面客户端" >}}

1. 选择要导出的对话，点击右上角的三个点，然后点击 `导出聊天历史`。
2. 取消选中所有选项（您现在不需要下载它们），将 `大小限制` 设置为最小值。
3. 设置格式为 `JSON` 并选择您需要的时间段。
4. 导出它！`result.json` 就是您需要的文件。
   {{< /tab >}}
   {{< /tabs >}}

{{< command >}}
tdl dl -f result1.json -f result2.json
{{< /command >}}

## 合并下载：

{{< command >}}
tdl dl \
-u https://t.me/tdl/1 -u https://t.me/tdl/2 \
-f result1.json -f result2.json
{{< /command >}}

## 自定义目录：

将文件下载到自定义目录

{{< command >}}
tdl dl -u https://t.me/tdl/1 -d /path/to/dir
{{< /command >}}

## 自定义参数：

使用每个任务8个线程，512KiB（最大）的分块大小，4个并发任务下载：

{{< command >}}
tdl dl -u https://t.me/tdl/1 -t 8 -s 524288 -l 4
{{< /command >}}

## 反序下载：

按反序下载文件（从最新到最旧）

{{< hint warning >}}
不同的顺序将影响“恢复下载”功能
{{< /hint >}}

{{< command >}}
tdl dl -f result.json --desc
{{< /command >}}

## MIME 探测：

如果文件扩展名与 MIME 类型不匹配，tdl将使用正确的扩展名重命名文件。

{{< hint warning >}}
副作用：例如 `.apk` 文件将被重命名为 `.zip`。
{{< /hint >}}

{{< command >}}
tdl dl -u https://t.me/tdl/1 --rewrite-ext
{{< /command >}}

## 自动跳过

在下载时跳过相同的文件（即名称和大小相同）。

{{< command >}}
tdl dl -u https://t.me/tdl/1 --skip-same
{{< /command >}}

## "Takeout" 会话

通过 ["Takeout" 会话](https://arabic-telethon.readthedocs.io/en/stable/extra/examples/telegram-client.html#exporting-messages) 下载文件：

> 如果下载大量媒体，更倾向于使用 "Takeout" 会话，它允许您以较低的速率限制从您的帐户中导出数据。

{{< command >}}
tdl dl -u https://t.me/tdl/1 --takeout
{{< /command >}}

## 过滤器

使用扩展名过滤器下载文件：

{{< hint warning >}}
扩展名仅与文件名匹配，而不是 MIME 类型。因此，这可能不会按预期工作。

白名单和黑名单不能同时使用。
{{< /hint >}}

白名单：只下载扩展名为 `.jpg` `.png` 的文件

{{< command >}}
tdl dl -u https://t.me/tdl/1 -i jpg,png
{{< /command >}}

黑名单：下载除了扩展名为 `.mp4` `.flv` 的所有文件

{{< command >}}
tdl dl -u https://t.me/tdl/1 -e mp4,flv
{{< /command >}}

## 文件名模板

使用自定义文件名模板下载：

请参考 [模板指南](/zh/guide/template) 了解更多。

{{< command >}}
tdl dl -u https://t.me/tdl/1 \
--template "{{ .DialogID }}_{{ .MessageID }}_{{ .DownloadDate }}_{{ .FileName }}"
{{< /command >}}

## 恢复/重新开始下载

在不需要交互的情况下恢复下载：

{{< command >}}
tdl dl -u https://t.me/tdl/1 --continue
{{< /command >}}

在不需要交互的情况下重新开始下载：

{{< command >}}
tdl dl -u https://t.me/tdl/1 --restart
{{< /command >}}

## HTTP 文件服务器

将文件暴露为 HTTP 服务器，而不使用内置下载它们

{{< hint info >}}
当您想要使用下载管理器（如 `aria2`/`wget`/`axel`/`IDM`）下载文件时，适合使用此选项。
{{< /hint >}}

{{< command >}}
tdl dl -u https://t.me/tdl/1 --serve
{{< /command >}}

使用自定义端口：

{{< command >}}
tdl dl -u https://t.me/tdl/1 --serve --port 8081
{{< /command >}}
