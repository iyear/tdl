---
title: "Quick Start"
weight: 20
---

# Quick Start

## Login

We don't specify the namespace here, so it will use the `default` namespace. You can specify the namespace with
`-n` flag if you want to use another namespace.

### **Login with desktop clients**

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

### **Login with QR code**

{{< command >}}
tdl login -T qr
{{< /command >}}

### **Login with phone & code**

{{< command >}}
tdl login -T code
{{< /command >}}

## Download

We download media from Telegram official channel:

{{< command >}}
tdl dl -u https://t.me/telegram/193
{{< /command >}}
