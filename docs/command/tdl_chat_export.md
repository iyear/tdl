## tdl chat export

export messages from (protected) chat for download

```
tdl chat export [flags]
```

### Options

```
      --all             export all messages including non-media messages, but still affected by filter and type flag
  -c, --chat string     chat id or domain. If not specified, 'Saved Messages' will be used
  -f, --filter string   filter messages by expression, defaults to match all messages. Specify '-' to see available fields (default "true")
  -h, --help            help for export
  -i, --input ints      input data, depends on export type
  -o, --output string   output JSON file path (default "tdl-export.json")
      --raw             export raw message struct of Telegram MTProto API, useful for debugging
      --reply int       specify channel post id
      --topic int       specify topic id
  -T, --type string     export type. time: timestamp range, id: message id range, last: last N messages: {time|id|last} (default "time")
      --with-content    export with message content
```

### Options inherited from parent commands

```
      --debug                        enable debug mode
  -l, --limit int                    max number of concurrent tasks (default 2)
  -n, --ns string                    namespace for Telegram session
      --ntp string                   ntp server host, if not set, use system time
      --pool int                     specify the size of the DC pool (default 3)
      --proxy string                 proxy address, only socks5 is supported, format: protocol://username:password@host:port
      --reconnect-timeout duration   Telegram client reconnection backoff timeout, infinite if set to 0 (default 2m0s)
  -s, --size int                     part size for transfer, max is 512*1024 (default 524288)
      --test string                  use test Telegram client, only for developer
  -t, --threads int                  max threads for transfer one item (default 4)
```

### SEE ALSO

* [tdl chat](tdl_chat.md)	 - A set of chat tools

