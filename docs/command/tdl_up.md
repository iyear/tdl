## tdl up

Upload anything to Telegram

```
tdl up [flags]
```

### Examples

```
tdl up -n iyear --proxy socks5://localhost:1080 -p /path/to/file -p /path -e .so -e .tmp -s 262144 -t 16 -l 3
```

### Options

```
  -e, --excludes strings   exclude the specified file extensions
  -h, --help               help for up
  -p, --path strings       dirs or files
```

### Options inherited from parent commands

```
      --debug          enable debug mode
  -l, --limit int      max number of concurrent tasks (default 2)
  -n, --ns string      namespace for Telegram session
      --proxy string   proxy address, only socks5 is supported, format: protocol://username:password@host:port
  -s, --size int       part size for transfer, max is 512*1024 (default 524288)
  -t, --threads int    threads for transfer one item (default 8)
```

### SEE ALSO

* [tdl](tdl.md)	 - Telegram Downloader, but more than a downloader

