## tdl dl

Download anything from Telegram (protected) chat

```
tdl dl [flags]
```

### Examples

```
tdl dl -n iyear --proxy socks5://localhost:1080 -u https://t.me/tdl/1 -u https://t.me/tdl/2 -s 262144 -t 16 -l 3
```

### Options

```
  -h, --help          help for dl
  -u, --url strings   telegram message links to be downloaded
```

### Options inherited from parent commands

```
      --debug          enable debug mode
  -l, --limit int      max number of concurrent tasks (default 2)
  -n, --ns string      namespace for Telegram session
      --ntp string     ntp server host, if not set, use system time
      --proxy string   proxy address, only socks5 is supported, format: protocol://username:password@host:port
  -s, --size int       part size for transfer, max is 512*1024 (default 524288)
  -t, --threads int    threads for transfer one item (default 8)
```

### SEE ALSO

* [tdl](tdl.md)	 - Telegram Downloader, but more than a downloader

