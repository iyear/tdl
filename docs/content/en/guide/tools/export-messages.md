---
title: "Export Messages"
weight: 30
---

# Export Messages

Export media messages from chats, channels, groups, etc. in JSON format.

{{< details title="CHAT Examples" open=false >}}

- `@iyear`
- `iyear`
- `123456789` (ID)
- `https://t.me/iyear`
- `+1 123456789` (Phone)
  {{< /details >}}

## All

Export all messages containing media to `tdl-export.json`

{{< command >}}
tdl chat export -c CHAT
{{< /command >}}

## From Topic/Replies

Export media messages from specific topic:
{{< hint info >}}
Get Topic ID:

1. Message Link: `https://t.me/c/1492447836/251011/269724` (`251011` is topic id)
2. `tdl chat ls` command
   {{< /hint >}}

{{< command >}}
tdl chat export -c CHAT --topic TOPIC_ID
{{< /command >}}

Export media messages from specific channel post replies:

{{< command >}}
tdl chat export -c CHAT --reply POST_ID
{{< /command >}}

## Custom Destination

Export with specific output file path. Default: `tdl-export.json`.

{{< command >}}
tdl chat export -c CHAT -o /path/to/output.json
{{< /command >}}

## Custom Type

### Time Range

Export with specific timestamp range. Default: `1970-01-01` - `NOW`

{{< command >}}
tdl chat export -c CHAT -T time -i 1665700000,1665761624
{{< /command >}}

`time` is also the default value of `-T` option, so you can omit it

{{< command >}}
tdl chat export -c CHAT -i 1665700000,1665761624
{{< /command >}}

### ID Range

Export with specific message id range. Default: `0` - `latest`

{{< command >}}
tdl chat export -c CHAT -T id -i 100,500
{{< /command >}}

### Last

Export last 100 media messages:

{{< command >}}
tdl chat export -c CHAT -T last -i 100
{{< /command >}}

## Filter

Please refer to [Filter Guide](/guide/tools/filter) for basic knowledge about filter.

List all available filter fields:

{{< command >}}
tdl chat export -c CHAT -f -
{{< /command >}}

Export last 10 zip files that `size > 5MiB` and `views > 200`:

{{< command >}}
tdl chat export -c CHAT -T last -i 10 -f "Views>200 && Media.Name endsWith '.zip' && Media.Size > 5*1024*1024"
{{< /command >}}

## With Content

Export with message content:

{{< command >}}
tdl chat -c CHAT --with-content
{{< /command >}}

## Raw

Export Telegram MTProto raw message structure, which is useful for debugging.

{{< command >}}
tdl chat export -c CHAT --raw
{{< /command >}}

## Non-Media

Export all messages including non-media messages, which is useful for debugging/backup.

{{< command >}}
tdl chat export -c CHAT --all
{{< /command >}}
