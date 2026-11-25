---
title: "Data"
weight: 30
---

# Data

Your account information will be stored in the `~/.tdl` directory.

Log files will be stored in the `~/.tdl/log` directory.

Each invocation writes to its own `tdl-<pid>-<start-time>-<unique-id>.log` file,
so simultaneous commands do not share a log file. The namespace is recorded in
the log. Sort files by modification time to find a recent run.

Logs rotate at 10 MiB, with up to three compressed backups per invocation.
At startup, tdl also attempts to remove logs and backups last modified more than
seven days ago if their owning process no longer exists. Files are kept if the
process is still running or its status cannot be checked. Reused process IDs can
delay cleanup. Legacy `latest.log` files and their backups are left untouched.
