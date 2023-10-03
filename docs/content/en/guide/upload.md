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

{{< details title="CHAT Examples" open=false >}}

- `@iyear`
- `iyear`
- `123456789` (ID)
- `https://t.me/iyear`
- `+1 123456789` (Phone)
  {{< /details >}}

{{< command >}}
tdl up -p /path/to/file -c CHAT
{{< /command >}}

## Custom Parameters

Upload with 8 threads per task, 512KiB(MAX) part size, 4 concurrent tasks:

{{< command >}}
tdl up -p /path/to/file -t 8 -s 524288 -l 4
{{< /command >}}

## Filter

Upload files except specified extensions:

{{< command >}}
tdl up -p /path/to/file -p /path/to/dir -e .so -e .tmp
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


