## tdl upload

Upload anything to Telegram

```
tdl upload [flags]
```

### Options

```
  -c, --chat string        chat id or domain, and empty means 'Saved Messages'
  -e, --excludes strings   exclude the specified file extensions
  -h, --help               help for upload
  -p, --path strings       dirs or files
```

### Options inherited from parent commands

```
      --debug          enable debug mode
  -l, --limit int      max number of concurrent tasks (default 2)
  -n, --ns string      namespace for Telegram session
      --ntp string     ntp server host, if not set, use system time
      --proxy string   proxy address, only socks5 is supported, format: protocol://username:password@host:port
  -s, --size int       part size for transfer, max is 512*1024 (default 524288)
  -t, --threads int    max threads for transfer one item (default 4)
```

### SEE ALSO

* [tdl](tdl.md)	 - Telegram Downloader, but more than a downloader

