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

Custom caption is based on [expression](/reference/expr).

List all available fields:

{{< command >}}
tdl up -p ./downloads --caption -
{{< /command >}}

Custom simple caption:
{{< command >}}
tdl up -p ./downloads --caption 'File.Name + " - uploaded by tdl"'
{{< /command >}}

Write styled message with [HTML](https://core.telegram.org/bots/api#html-style):
{{< command >}}
tdl up -p ./downloads --caption  \
'FileName + `<b>Bold</b> <a href="https://example.com">Link</a>`'
{{< /command >}}

Pass a file name if the expression is complex:

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
tdl up -p ./downloads --caption caption.txt
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
