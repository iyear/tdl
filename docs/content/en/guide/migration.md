---
title: "Migration"
weight: 50
---

# Migration

Backup or recover your data

## Backup

Backup your data to a zip file. Default: `tdl-backup-<time>.zip`.

{{< command >}}
tdl backup
{{< /command >}}

Or specify the output file:

{{< command >}}
tdl backup -d /path/to/backup.zip
{{< /command >}}

## Recover

Recover your data from a zip file.

{{< command >}}
tdl recover -f /path/to/backup.zip
{{< /command >}}
