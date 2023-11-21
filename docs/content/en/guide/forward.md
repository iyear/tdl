---
title: "Forward"
weight: 35
---

# Forward

Forward messages with automatic fallback and message routing

One-liner to forward messages from `https://t.me/telegram/193` to `Saved Messages`:

{{< command >}}
tdl forward --from https://t.me/telegram/193
{{< /command >}}

## Custom Source

{{< include "snippets/link.md" >}}

You can forward messages from links and [exported JSON files](/guide/download#from-json):

{{< command >}}
tdl forward \ 
    --from https://t.me/telegram/193 \ 
    --from https://t.me/telegram/195 \
    --from tdl-export.json \
    --from tdl-export2.json
{{< /command >}}

## Custom Destination

{{< include "snippets/chat.md" >}}

### Specific Chat

Forward to specific one chat:

{{< command >}}
tdl forward --from tdl-export.json --to CHAT
{{< /command >}}

### Message Routing

Forward to different chats by message router which is based on [expression](/reference/expr).

List all available fields:

{{< command >}}
tdl forward --from tdl-export.json --to -
{{< /command >}}

Forward to `CHAT1` if message contains `foo`, otherwise forward to `Saved Messages`:

{{< hint info >}}
You must return a string as the target CHAT, and empty string means forward to `Saved Messages`.
{{< /hint >}}

{{< command >}}
tdl forward --from tdl-export.json \
    --to 'Message.Message contains "foo" ? "CHAT1" : ""'
{{< /command >}}

Pass a file name if the expression is complex:

{{< details "router.txt" >}}
Write your expression like `switch`:
```
Message.Message contains "foo" ? "CHAT1" :
From.ID == 123456 ? "CHAT2" :
Message.Views > 30 ? "CHAT3" :
""
```
{{< /details >}}

{{< command >}}
tdl forward --from tdl-export.json --to router.txt
{{< /command >}}

## Mode

Forward messages with automatic fallback strategy.

Available modes:
- `direct` (default)
- `clone`

### Direct

Prefer to use official forward API. 

If the chat or message is not allowed to use official forward API, it will be automatically downgraded to `clone` mode.

{{< command >}}
tdl forward --from tdl-export.json --mode direct
{{< /command >}}

### Clone

Forward messages by copying them, which doesn't have forwarded header.

Some message content can't be copied, such as poll, invoice, etc. They will be ignored.

{{< command >}}
tdl forward --from tdl-export.json --mode clone
{{< /command >}}

## Dry Run

Print the progress without actually sending messages, which is useful for message routing debugging.

{{< command >}}
tdl forward --from tdl-export.json --dry-run
{{< /command >}}

## Silent

Send messages without notification.

{{< command >}}
tdl forward --from tdl-export.json --silent
{{< /command >}}
