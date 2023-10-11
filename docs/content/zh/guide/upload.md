---
title: "上传"
weight: 40
---

# 上传

## 上传文件

上传指定的文件和目录到 `保存的消息`：

{{< command >}}
tdl up -p /path/to/file -p /path/to/dir
{{< /command >}}

## 自定义目标

上传到自定义聊天。

{{< details title="CHAT 示例" open=false >}}

- `@iyear`
- `iyear`
- `123456789`（ID）
- `https://t.me/iyear`
- `+1 123456789`（电话）
  {{< /details >}}

{{< command >}}
tdl up -p /path/to/file -c CHAT
{{< /command >}}

## 自定义参数

使用每个任务8个线程、512KiB（最大）分块大小、4个并发任务上传：

{{< command >}}
tdl up -p /path/to/file -t 8 -s 524288 -l 4
{{< /command >}}

## 过滤器

上传除指定扩展名之外的文件：

{{< command >}}
tdl up -p /path/to/file -p /path/to/dir -e .so -e .tmp
{{< /command >}}

## 自动删除

删除已上传成功的文件：

{{< command >}}
tdl up -p /path/to/file --rm
{{< /command >}}

## 照片

将图像作为照片而不是文件上传：

{{< command >}}
tdl up -p /path/to/file --photo
{{< /command >}}


