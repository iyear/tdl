---
title: "Upload"
weight: 40
---

# Upload

## Upload Files

Upload specified files and directories to `Saved Messages`:

```
tdl up -p /path/to/file -p /path/to/dir
```

## Custom Destination

Upload to custom chat.

{{< details title="CHAT Examples" open=false >}}

- `@iyear`
- `iyear`
- `123456789` (ID)
- `https://t.me/iyear`
- `+1 123456789` (Phone)
  {{< /details >}}

```
tdl up -p /path/to/file -c CHAT
```

## Custom Parameters

Upload with 8 threads per task, 512KiB(MAX) part size, 4 concurrent tasks:

```
tdl up -p /path/to/file -t 8 -s 524288 -l 4
```

## Filter

Upload files except specified extensions:

```
tdl up -p /path/to/file -p /path/to/dir -e .so -e .tmp
```

## Delete Local

Delete the uploaded file after uploading successfully:

```
tdl up -p /path/to/file --rm
```

## Photo

Upload images as photos instead of documents:

```
tdl up -p /path/to/file --photo
```


