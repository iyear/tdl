## tdl dl

Download what you want

```
tdl dl [flags]
```

### Examples

```
tdl dl
```

### Options

```
  -h, --help            help for dl
  -l, --limit int       max number of concurrent tasks (default 2)
  -m, --mode string     mode for download
  -s, --part-size int   part size for download, max is 512*1024 (default 524288)
  -t, --threads int     threads for downloading one item (default 8)
  -u, --url strings     array of message links to be downloaded
```

### Options inherited from parent commands

```
  -n, --ns string      namespace for Telegram session
      --proxy string   proxy address, only socks5 is supported, format: protocol://username:password@host:port
```

### SEE ALSO

* [tdl](tdl.md)	 - Telegram downloader, but not only a downloader

