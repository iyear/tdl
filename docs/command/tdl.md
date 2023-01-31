## tdl

Telegram Downloader, but more than a downloader

### Options

```
      --debug          enable debug mode
  -h, --help           help for tdl
  -l, --limit int      max number of concurrent tasks (default 2)
  -n, --ns string      namespace for Telegram session
      --ntp string     ntp server host, if not set, use system time
      --proxy string   proxy address, only socks5 is supported, format: protocol://username:password@host:port
  -s, --size int       part size for transfer, max is 512*1024 (default 131072)
  -t, --threads int    max threads for transfer one item (default 4)
```

### SEE ALSO

* [tdl backup](tdl_backup.md)	 - Backup your data
* [tdl chat](tdl_chat.md)	 - A set of chat tools
* [tdl dl](tdl_dl.md)	 - Download anything from Telegram (protected) chat
* [tdl login](tdl_login.md)	 - Login to Telegram
* [tdl recover](tdl_recover.md)	 - Recover your data
* [tdl up](tdl_up.md)	 - Upload anything to Telegram
* [tdl version](tdl_version.md)	 - Check the version info

