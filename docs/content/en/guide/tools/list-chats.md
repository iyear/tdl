---
title: "List Chats"
weight: 10
---

# List Chats

## List all chats

{{< command >}}
tdl chat ls
{{< /command >}}

## JSON Output

{{< command >}}
tdl chat ls -o json
{{< /command >}}

## Filter

Please refer to [Filter Guide](/guide/tools/filter) for basic knowledge about filter.

List all available filter fields:

{{< command >}}
tdl chat ls -f -
{{< /command >}}

List channels that VisibleName contains "Telegram":

{{< command >}}
tdl chat ls -f "Type contains 'channel' && VisibleName contains 'Telegram'"
{{< /command >}}

List groups that have topics:

{{< command >}}
tdl chat ls -f "len(Topics)>0"
{{< /command >}}

