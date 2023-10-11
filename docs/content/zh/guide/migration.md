---
title: "迁移"
weight: 50
---

# 迁移

备份或恢复您的数据

## 备份

将您的数据备份到 zip 文件中。默认值：`tdl-backup-<time>.zip`。

{{< command >}}
tdl backup
{{< /command >}}

或者指定输出文件：

{{< command >}}
tdl backup -d /path/to/backup.zip
{{< /command >}}

## 恢复

从 zip 文件中恢复您的数据。

{{< command >}}
tdl recover -f /path/to/backup.zip
{{< /command >}}
