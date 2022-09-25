## tdl dl

Download anything from Telegram (protected) chat

```
tdl dl [flags]
```

### Options

```
  -f, --file strings   official client export files
  -h, --help           help for dl
  -u, --url strings    telegram message links
```

### Options inherited from parent commands

```
      --debug          enable debug mode
  -l, --limit int      max number of concurrent tasks (default 2)
  -n, --ns string      namespace for Telegram session
      --ntp string     ntp server host, if not set, use system time
      --proxy string   proxy address, only socks5 is supported, format: protocol://username:password@host:port
  -s, --size int       part size for transfer, max is 512*1024 (default 131072)
  -t, --threads int    threads for transfer one item (default 4)
```

### SEE ALSO

* [tdl](tdl.md)	 - Telegram Downloader, but more than a downloader

