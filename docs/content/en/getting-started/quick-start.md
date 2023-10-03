---
title: "Quick Start"
weight: 20
---

# Quick Start

## Login

Each namespace represents a Telegram account

You should set the namespace **when each command is executed**:

### **Login with official clients(Recommended)**

{{< hint warning >}}
Please ensure that clients are downloaded from [official website](https://desktop.telegram.org/) (NOT from Microsoft
Store or App Store)
{{< /hint >}}

Automatically find the client path:

{{< command >}}
tdl login -n quickstart
{{< /command >}}

Or if you set a local passcode:

{{< command >}}
tdl login -n quickstart -p YOUR_PASSCODE
{{< /command >}}

Or specify custom client path:

{{< command >}}
tdl login -n quickstart -d /path/to/TelegramDesktop
{{< /command >}}

### **Login with phone & code**

{{< command >}}
tdl login -n quickstart --code
{{< /command >}}

## Download

We use account `quickstart` to download media from Telegram official channel:

{{< command >}}
tdl dl -n quickstart -u https://t.me/telegram/193
{{< /command >}}
