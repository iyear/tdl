## tdl login

Login to Telegram

```
tdl login [flags]
```

### Examples

```
tdl login -n iyear --proxy socks5://localhost:1080
```

### Options

```
  -d, --desktop string   Official desktop client path, import session from it
  -h, --help             help for login
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

