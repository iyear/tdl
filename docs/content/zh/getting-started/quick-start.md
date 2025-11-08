---
title: "快速开始"
weight: 20
---

# 快速开始

## 登录

我们不在这里指定命名空间，它将使用 `default` 命名空间。如果你想使用其他命名空间，可以使用 `-n` 标志指定命名空间。

### **使用桌面客户端登录**

{{< hint warning >}}
请确保从[官方网站](https://desktop.telegram.org/)下载客户端（不要从 Microsoft Store 或 App Store 下载）
{{< /hint >}}

使用默认路径：

{{< command >}}
tdl login
{{< /command >}}

如果您设置了本地密码：

{{< command >}}
tdl login -p YOUR_PASSCODE
{{< /command >}}

或者指定自定义客户端路径：

{{< command >}}
tdl login -d /path/to/TelegramDesktop
{{< /command >}}

### **使用二维码登录**

{{< command >}}
tdl login -T qr
{{< /command >}}

### **使用手机号码和验证码登录**

{{< command >}}
tdl login -T code
{{< /command >}}

## 验证设置

登录后，您可以验证设置是否正常工作：

{{< command >}}
tdl doctor
{{< /command >}}

这将检查您的时间同步、连通性、数据库和登录状态。更多详细信息请参阅 [诊断](/zh/guide/doctor)。

## 下载

我们从 Telegram 官方频道下载文件：

{{< command >}}
tdl dl -u https://t.me/telegram/193
{{< /command >}}
