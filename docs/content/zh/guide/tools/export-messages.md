---
title: "导出消息"
weight: 30
---

# 导出消息

以 JSON 格式导出聊天、频道、群组等中的媒体消息。

{{< details title="CHAT 示例" open=false >}}

- `@iyear`
- `iyear`
- `123456789` (ID)
- `https://t.me/iyear`
- `+1 123456789` (电话号码)
  {{< /details >}}

## 所有消息

将包含媒体的所有消息导出到 `tdl-export.json`

{{< command >}}
tdl chat export -c CHAT
{{< /command >}}

## 从主题/回复中导出

从特定主题导出媒体消息：
{{< hint info >}}
获取主题 ID 的方式：

1. 消息链接：`https://t.me/c/1492447836/251011/269724`（`251011` 是主题 ID）
2. `tdl chat ls` 命令
   {{< /hint >}}

{{< command >}}
tdl chat export -c CHAT --topic TOPIC_ID
{{< /command >}}

从特定频道帖子的回复中导出媒体消息：

{{< command >}}
tdl chat export -c CHAT --reply POST_ID
{{< /command >}}

## 自定义路径

指定输出文件路径进行导出。默认：`tdl-export.json`。

{{< command >}}
tdl chat export -c CHAT -o /path/to/output.json
{{< /command >}}

## 自定义类型

### 时间范围

根据特定的时间戳范围进行导出。默认：`1970-01-01` - `当前`

{{< command >}}
tdl chat export -c CHAT -T time -i 1665700000,1665761624
{{< /command >}}

`time` 也是 `-T` 选项的默认值，因此您可以省略它

{{< command >}}
tdl chat export -c CHAT -i 1665700000,1665761624
{{< /command >}}

### ID 范围

根据特定的消息 ID 范围进行导出。默认：`0` - `最新`

{{< command >}}
tdl chat export -c CHAT -T id -i 100,500
{{< /command >}}

### 最新

导出最后 100 条媒体文件：

{{< command >}}
tdl chat export -c CHAT -T last -i 100
{{< /command >}}

## 过滤

请参考[过滤指南](/zh/guide/tools/filter)以获取有关过滤器的基本知识。

列出所有可用的过滤字段：

{{< command >}}
tdl chat export -c CHAT -f -
{{< /command >}}

导出最后的 10 个媒体文件，其中 `大小 > 5MiB` 且 `查看次数 > 200`：

{{< command >}}
tdl chat export -c CHAT -T last -i 10 -f "Views>200 && Media.Name endsWith '.zip' && Media.Size > 5*1024*1024"
{{< /command >}}

## 包含内容

附带消息内容：

{{< command >}}
tdl chat -c CHAT --with-content
{{< /command >}}

## 原始数据

导出 Telegram MTProto 原始消息结构，用于调试。

{{< command >}}
tdl chat export -c CHAT --raw
{{< /command >}}

## 非媒体消息

导出包括非媒体消息的所有消息，用于调试/备份。

{{< command >}}
tdl chat export -c CHAT --all
{{< /command >}}
