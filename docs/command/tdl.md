## tdl

Telegram Downloader, but more than a downloader

### Examples

```
tdl -h
```

### Options

```
  -h, --help            help for tdl
  -l, --limit int       max number of concurrent tasks (default 2)
  -n, --ns string       namespace for Telegram session
  -s, --part-size int   part size for transfer, max is 512*1024 (default 524288)
      --proxy string    proxy address, only socks5 is supported, format: protocol://username:password@host:port
  -t, --threads int     threads for transfer one item (default 8)
```

### SEE ALSO

* [tdl chat](tdl_chat.md)	 - A set of chat tools
* [tdl dl](tdl_dl.md)	 - Download what you want
* [tdl login](tdl_login.md)	 - Login to Telegram
* [tdl up](tdl_up.md)	 - Upload anything to Telegram
* [tdl version](tdl_version.md)	 - Check the version info

