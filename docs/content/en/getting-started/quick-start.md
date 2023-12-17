---
title: "Quick Start"
weight: 20
---

# Quick Start

## Login

### **Login with official clients(Recommended)**

{{< hint warning >}}
Please ensure that clients are downloaded from [official website](https://desktop.telegram.org/) (NOT from Microsoft
Store or App Store)
{{< /hint >}}

Automatically find the client path:

{{< command >}}
tdl login
{{< /command >}}

Or if you set a local passcode:

{{< command >}}
tdl login -p YOUR_PASSCODE
{{< /command >}}

Or specify custom client path:

{{< command >}}
tdl login -d /path/to/TelegramDesktop
{{< /command >}}

### **Login with phone & code**

{{< command >}}
tdl login --code
{{< /command >}}

## Download

We download media from Telegram official channel:

{{< command >}}
tdl dl -u https://t.me/telegram/193
{{< /command >}}
