---
title: "列出聊天"
weight: 10
---

# 列出聊天

## 列出所有聊天

{{< command >}}
tdl chat ls
{{< /command >}}

## JSON 格式

{{< command >}}
tdl chat ls -o json
{{< /command >}}

## 过滤器

请参考 [过滤器指南](/zh/guide/tools/filter) 以获取有关过滤器的基本知识。

列出所有可用的过滤字段：

{{< command >}}
tdl chat ls -f -
{{< /command >}}

列出名字包含 "Telegram" 的频道：

{{< command >}}
tdl chat ls -f "Type contains 'channel' && VisibleName contains 'Telegram'"
{{< /command >}}

列出具有主题的群组：

{{< command >}}
tdl chat ls -f "len(Topics)>0"
{{< /command >}}
