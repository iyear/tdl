---
title: "Download"
weight: 30
---

# Download

## From Links:

{{< hint info >}}
Get message links from "Copy Link" button in official clients.
{{< /hint >}}

{{< details title="Message Link Examples" open=false >}}

- `https://t.me/telegram/193`
- `https://t.me/c/1697797156/151`
- `https://t.me/iFreeKnow/45662/55005`
- `https://t.me/c/1492447836/251015/251021`
- `...` (File a new issue if you find a new link format)
  {{< /details >}}

```
tdl dl -u https://t.me/tdl/1 -u https://t.me/tdl/2
```

## From JSON:

There are two ways to export the JSON you need:

{{< tabs "json" >}}
{{< tab "tdl" >}}
This is especially for protected chats and more powerful than the desktop client.

Please refer to [Chat Export](/docs/)
{{< /tab >}}

{{< tab "Desktop Client" >}}

1. Choose the dialog you want to export, and click the three dots in the upper right corner, then
   click `Export Chat History`.
2. Uncheck all boxes(you don't need to download them now) and set `Size Limit` to minimum
3. Set Format to `JSON` and select the time period you need.
4. Export it! And `result.json` is what you need.
   {{< /tab >}}
   {{< /tabs >}}

```
tdl dl -f result1.json -f result2.json
```

## Combine Sources:

```
tdl dl \
-u https://t.me/tdl/1 -u https://t.me/tdl/2 \
-f result1.json -f result2.json
```

## Custom Destination:

Download files to custom directory

```
tdl dl -u https://t.me/tdl/1 -d /path/to/dir
```

## Custom Parameters:

Download with 8 threads per task, 512KiB(MAX) part size, 4 concurrent tasks:

```
tdl dl -u https://t.me/tdl/1 -t 8 -s 524288 -l 4
```

## Descending Order:

Download files in descending order(from newest to oldest)

{{< hint warning >}}
Different order will affect resuming download
{{< /hint >}}

```
tdl dl -f result.json --desc
```

## MIME Detection:

If the file extension is not matched with the MIME type, tdl will rename the file with the correct extension.

{{< hint warning >}}
Side effect: like `.apk` file, it will be renamed to `.zip`.
{{< /hint >}}

```
tdl dl -u https://t.me/tdl/1 --rewrite-ext
```

## Auto Skip

Skip the same files(name and size) when downloading.

```
tdl dl -u https://t.me/tdl/1 --skip-same
```

## Takeout Session

Download files
with [takeout session](https://arabic-telethon.readthedocs.io/en/stable/extra/examples/telegram-client.html#exporting-messages):

> If you plan to download a lot of media, you may prefer to do this within a takeout session. Takeout sessions let you
> export data from your account with lower flood wait limits.

```
tdl dl -u https://t.me/tdl/1 --takeout
```

## Filters

Download files with extension filters:

{{< hint warning >}}
The extension is only matched with the file name, not the MIME type. So it may not work as expected.

Whitelist and blacklist can not be used at the same time.
{{< /hint >}}

Whitelist: Only download files with `.jpg` `.png` extension

```
tdl dl -u https://t.me/tdl/1 -i jpg,png
```

Blacklist: Download all files except `.mp4` `.flv` extension

```
tdl dl -u https://t.me/tdl/1 -e mp4,flv
```

## Name Template

Download with custom file name template:

Please refer to [Template Guide](/docs/guide/reference/template) for more details.

```
tdl dl -u https://t.me/tdl/1 \
--template "{{ .DialogID }}_{{ .MessageID }}_{{ .DownloadDate }}_{{ .FileName }}"
```

## Resume/Restart

Resume without UI interaction:

```
tdl dl -u https://t.me/tdl/1 --continue
```

Restart without UI interaction:

```
tdl dl -u https://t.me/tdl/1 --restart
```

## Serve

Expose the files as an HTTP server instead of downloading them with built-in downloader

{{< hint info >}}
This is useful when you want to download files with a download manager like `aria2`/`wget`/`axel`/`IDM`...
{{< /hint >}}

```
tdl dl -u https://t.me/tdl/1 --serve
```

With custom port:

```
tdl dl -u https://t.me/tdl/1 --serve --port 8081
```
