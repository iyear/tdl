---
title: "Migration"
weight: 50
---

# Migration

Backup or recover your data

## Backup

Backup your data to a zip file. Default: `tdl-backup-<time>.zip`.

```
tdl backup
```

Or specify the output file:

```
tdl backup -d /path/to/backup.zip
```

## Recover

Recover your data from a zip file.

```
tdl recover -f /path/to/backup.zip
```
