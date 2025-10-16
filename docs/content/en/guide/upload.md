---
title: "Upload"
weight: 40
---

# Upload

## Upload Files

Upload specified files and directories to `Saved Messages`:

{{< command >}}
tdl up -p /path/to/file -p /path/to/dir
{{< /command >}}

## Custom Destination

Upload to custom chat.

{{< include "snippets/chat.md" >}}

{{< command >}}
tdl up -p /path/to/file -c CHAT
{{< /command >}}

## Custom Parameters

Upload with 8 threads per task, 4 concurrent tasks:

{{< command >}}
tdl up -p /path/to/file -t 8 -l 4
{{< /command >}}

## Custom Caption

Custom Caption is based on [expression](/reference/expr).

List all available fields:

{{< command >}}
tdl up -p ./foo --caption -
{{< /command >}}

Supported Style:

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

Example:

{{< command >}}
tdl  up -p ./downloads --caption '[{style: "code", text: File}, "-", {style: "bold", text: Filename}, "-", {style: "strike", text: Extension}, "-", {style: "italic", text: Mime}]'
{{< /command >}}

## Filter

Upload files except specified extensions:

{{< command >}}
tdl up -p /path/to/file -p /path/to/dir -e .so -e .tmp
{{< /command >}}

## Filters

Upload files with extension filters:

{{< hint warning >}}
The extension is only matched with the file name, not the MIME type. So it may not work as expected.

Whitelist and blacklist can not be used at the same time.
{{< /hint >}}

Whitelist: Only upload files with `.jpg` `.png` extension

{{< command >}}
tdl up -p /path/to/file -p /path/to/dir -i jpg,png
{{< /command >}}

Blacklist: Upload all files except `.mp4` `.flv` extension

{{< command >}}
tdl up -p /path/to/file -p /path/to/dir -e mp4 -e flv
{{< /command >}}

## Delete Local

Delete the uploaded file after uploading successfully:

{{< command >}}
tdl up -p /path/to/file --rm
{{< /command >}}

## Photo

Upload images as photos instead of documents:

{{< command >}}
tdl up -p /path/to/file --photo
{{< /command >}}
