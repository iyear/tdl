---
title: "Export Members"
weight: 20
---

# Export Members

Export chat members/subscribers, admins, bots, etc.

{{< hint info >}}
Chat administrator permission is required.
{{< /hint >}}

## Export all members

Export all users to `tdl-users.json`

{{< details title="CHAT Examples" open=false >}}

- `@iyear`
- `iyear`
- `123456789` (ID)
- `https://t.me/iyear`
- `+1 123456789` (Phone)
  {{< /details >}}

```
tdl chat users -c CHAT
```

## Custom Destination

Export with specified file path

```
tdl chat users -c CHAT -o /path/to/export.json
```

## Raw

Export Telegram MTProto raw user structure, which is useful for debugging.

```
tdl chat users -c CHAT --raw
```
