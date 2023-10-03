---
title: "Export Members"
weight: 20
---

# Export Members

Export chat members/subscribers, admins, bots, etc.

{{< hint info >}}
Chat administrator permission is required.
{{< /hint >}}

{{< details title="CHAT Examples" open=false >}}

- `@iyear`
- `iyear`
- `123456789` (ID)
- `https://t.me/iyear`
- `+1 123456789` (Phone)
  {{< /details >}}

## All

Export all users to `tdl-users.json`

{{< command >}}
tdl chat users -c CHAT
{{< /command >}}

## Custom Destination

Export with specified file path

{{< command >}}
tdl chat users -c CHAT -o /path/to/export.json
{{< /command >}}

## Raw

Export Telegram MTProto raw user structure, which is useful for debugging.

{{< command >}}
tdl chat users -c CHAT --raw
{{< /command >}}
