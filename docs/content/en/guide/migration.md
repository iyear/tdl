---
title: "Migration"
weight: 50
---

# Migration

Backup or recover your data

## Backup

Backup your data to a file. Default: `<date>.backup.tdl`.

{{< command >}}
tdl backup
{{< /command >}}

Or specify the output file:

{{< command >}}
tdl backup -d /path/to/custom.tdl
{{< /command >}}

## Recover

Recover your data from a tdl backup file.

{{< command >}}
tdl recover -f /path/to/custom.backup.tdl
{{< /command >}}
