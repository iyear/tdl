---
title: "List Chats"
weight: 10
---

# List Chats

## List all chats

```
tdl chat ls
```

## JSON Output

```
tdl chat ls -o json
```

## Filter

Please refer to [Filter Guide](/docs/guide/tools/filter) for basic knowledge about filter.

List all available filter fields:

```
tdl chat ls -f -
```

List channels that VisibleName contains "Telegram":

```
tdl chat ls -f "Type contains 'channel' && VisibleName contains 'Telegram'"
```

List groups that have topics:

```
tdl chat ls -f "len(Topics)>0"
```

