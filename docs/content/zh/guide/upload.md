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

## 自定义目标

上传到自定义聊天。

{{< include "snippets/chat.md" >}}

{{< command >}}
tdl up -p /path/to/file -c CHAT
{{< /command >}}

## 自定义参数

使用每个任务8个线程、4个并发任务上传：

{{< command >}}
tdl up -p /path/to/file -t 8 -l 4
{{< /command >}}

## 自定义说明文字

自定义说明文字基于 [表达式](/reference/expr)。

列出所有可用字段：

{{< command >}}
tdl up -p ./foo --caption -
{{< /command >}}

支持的样式：

```go
const (
	// StylePlain is a Style of type Plain.
	StylePlain Style = "Plain"
	// StyleUnknown is a Style of type Unknown.
	StyleUnknown Style = "Unknown"
	// StyleMention is a Style of type Mention.
	StyleMention Style = "Mention"
	// StyleHashtag is a Style of type Hashtag.
	StyleHashtag Style = "Hashtag"
	// StyleBotCommand is a Style of type BotCommand.
	StyleBotCommand Style = "BotCommand"
	// StyleURL is a Style of type URL.
	StyleURL Style = "URL"
	// StyleEmail is a Style of type Email.
	StyleEmail Style = "Email"
	// StyleBold is a Style of type Bold.
	StyleBold Style = "Bold"
	// StyleItalic is a Style of type Italic.
	StyleItalic Style = "Italic"
	// StyleCode is a Style of type Code.
	StyleCode Style = "Code"
	// StylePre is a Style of type Pre.
	StylePre Style = "Pre"
	// StyleTextURL is a Style of type TextURL.
	StyleTextURL Style = "TextURL"
	// StyleMentionName is a Style of type MentionName.
	StyleMentionName Style = "MentionName"
	// StylePhone is a Style of type Phone.
	StylePhone Style = "Phone"
	// StyleCashtag is a Style of type Cashtag.
	StyleCashtag Style = "Cashtag"
	// StyleUnderline is a Style of type Underline.
	StyleUnderline Style = "Underline"
	// StyleStrike is a Style of type Strike.
	StyleStrike Style = "Strike"
	// StyleBankCard is a Style of type BankCard.
	StyleBankCard Style = "BankCard"
	// StyleSpoiler is a Style of type Spoiler.
	StyleSpoiler Style = "Spoiler"
	// StyleCustomEmoji is a Style of type CustomEmoji.
	StyleCustomEmoji Style = "CustomEmoji"
	// StyleBlockquote is a Style of type Blockquote.
	StyleBlockquote Style = "Blockquote"
)
```

例子：

{{< command >}}
tdl  up -p ./downloads --caption '[{style: "code", text: File}, "-", {style: "bold", text: Filename}, "-", {style: "strike", text: Extension}, "-", {style: "italic", text: Mime}]'
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
