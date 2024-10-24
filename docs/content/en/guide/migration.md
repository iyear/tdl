---
title: "Migration"
weight: 50
---

# Migration

Backup or recover your data

## Backup

Backup all namespace data to a file. Default: `<date>.backup.tdl`.

{{< command >}}
tdl backup
{{< /command >}}

Or specify the output file:

{{< command >}}
tdl backup -d /path/to/custom.tdl
{{< /command >}}

## Recover

Recover your data from a tdl backup file. Existing namespace data will be overwritten.

{{< command >}}
tdl recover -f /path/to/custom.backup.tdl
{{< /command >}}

## Migrate

Migrate your data to another storage

See [Storage Flag](/guide/global-config/#--storage) for storage option details.

Migrate current storage to file storage:
{{< command >}}
tdl migrate --to type=file,path=/path/to/data.json
{{< /command >}}

Migrate custom source storage to file storage:
{{< command >}}
tdl migrate --storage type=bolt,path=/path/to/data-directory  --to type=file,path=/path/to/data.json
{{< /command >}}
