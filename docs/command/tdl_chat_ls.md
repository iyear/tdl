## tdl chat ls

List your chats

```
tdl chat ls [flags]
```

### Examples

```
tdl chat ls -n iyear --proxy socks5://localhost:1080
```

### Options

```
  -h, --help   help for ls
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

* [tdl chat](tdl_chat.md)	 - A set of chat tools

