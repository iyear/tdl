## tdl download

Download anything from Telegram (protected) chat

```
tdl download [flags]
```

### Options

```
      --continue          continue the last download directly
      --desc              download files from the newest to the oldest ones (may affect resume download)
  -d, --dir string        specify the download directory. If the directory does not exist, it will be created automatically (default "downloads")
  -e, --exclude strings   exclude the specified file extensions, and only judge by file name, not file MIME. Example: -e png,jpg
  -f, --file strings      official client exported files
  -h, --help              help for download
  -i, --include strings   include the specified file extensions, and only judge by file name, not file MIME. Example: -i mp4,mp3
      --pool int          specify the size of the DC pool (default 3)
      --restart           restart the last download directly
      --rewrite-ext       rewrite file extension according to file header MIME
      --skip-same         skip files with the same name(without extension) and size
      --takeout           takeout sessions let you export data from your account with lower flood wait limits.
      --template string   download file name template (default "{{ .DialogID }}_{{ .MessageID }}_{{ replace .FileName `/` `_` `\\` `_` `:` `_` `*` `_` `?` `_` `<` `_` `>` `_` `|` `_` ` ` `_`  }}")
  -u, --url strings       telegram message links
```

### Options inherited from parent commands

```
      --debug          enable debug mode
  -l, --limit int      max number of concurrent tasks (default 2)
  -n, --ns string      namespace for Telegram session
      --ntp string     ntp server host, if not set, use system time
      --proxy string   proxy address, only socks5 is supported, format: protocol://username:password@host:port
  -s, --size int       part size for transfer, max is 512*1024 (default 524288)
  -t, --threads int    max threads for transfer one item (default 4)
```

### SEE ALSO

* [tdl](tdl.md)	 - Telegram Downloader, but more than a downloader

