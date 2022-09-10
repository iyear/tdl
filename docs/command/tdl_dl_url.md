## tdl dl url

Download in url mode

```
tdl dl url [flags]
```

### Examples

```
tdl dl url -n iyear --proxy socks5://127.0.0.1:1080 -u https://t.me/tdl/1 -u https://t.me/tdl/2 -s 262144 -t 16 -l 3
```

### Options

```
  -h, --help           help for url
  -u, --urls strings   telegram message links to be downloaded
```

### Options inherited from parent commands

```
  -l, --limit int       max number of concurrent tasks (default 2)
  -n, --ns string       namespace for Telegram session
  -s, --part-size int   part size for download, max is 512*1024 (default 524288)
      --proxy string    proxy address, only socks5 is supported, format: protocol://username:password@host:port
  -t, --threads int     threads for downloading one item (default 8)
```

### SEE ALSO

* [tdl dl](tdl_dl.md)	 - Download what you want

