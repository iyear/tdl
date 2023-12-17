---
title: "快速开始"
weight: 20
---

# 快速开始

## 登录

### **使用官方客户端登录（推荐）**

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

### **使用手机号码和验证码登录**

{{< command >}}
tdl login --code
{{< /command >}}

## 下载

我们从 Telegram 官方频道下载文件：

{{< command >}}
tdl dl -u https://t.me/telegram/193
{{< /command >}}
