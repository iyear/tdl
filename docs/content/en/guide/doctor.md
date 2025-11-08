---
title: "Doctor"
weight: 15
---

# Doctor

The `doctor` command helps you diagnose common issues with TDL by running a series of automated checks.

## Usage

{{< command >}}
tdl doctor
{{< /command >}}

## What It Checks

The doctor command performs the following diagnostic checks:

### 1. Time Synchronization

Checks if your system time is synchronized with NTP servers. Telegram requires accurate time for authentication.

- Tests multiple NTP servers** with automatic fallback

**Status indicators:**
- **OK**: Time offset < 1 second
- **WARN**: Time offset between 1-10 seconds (may work, but consider syncing)
- **WARN**: Time offset > 10 seconds (may cause authentication issues)

### 2. Telegram Server Connectivity

Tests connection to Telegram servers using unauthenticated API endpoints:

- `help.getConfig` - Basic server configuration
- `help.getNearestDc` - Nearest datacenter location
- `langpack.getLanguages` - Language pack availability

This verifies network connectivity and access to Telegram's API.

### 3. Database Integrity

Checks your local TDL storage:

- Storage type and path
- Namespace availability
- Session data presence
- App configuration

### 4. Login Status

Verifies your authentication status and displays account information:

- Authentication status
- Account name and username
- User ID and phone number

## Custom NTP Server

You can specify a custom NTP server using the `--ntp` flag:

{{< command >}}
tdl doctor --ntp time.google.com
{{< /command >}}

## Common Issues

### Time Synchronization Failed

If all NTP servers fail, check your network connection and firewall settings. NTP uses UDP port 123.

### Connectivity Tests Failed

If connectivity tests fail:
1. Check your internet connection
2. Verify firewall settings allow connections to Telegram
3. Try using a proxy or VPN if Telegram is blocked in your region

### Database Issues

If database checks show warnings:
- Missing session data means you need to login: `tdl login`
- Database errors may require clean all local storages and re-initialization

### Not Authorized

If you're not logged in, run the login command first:

{{< command >}}
tdl login
{{< /command >}}
