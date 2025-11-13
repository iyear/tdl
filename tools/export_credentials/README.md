# Export Credentials

Export credentials from tdl storage to JSON format.

### Usage

```bash
# Export to file
go run export_credentials.go -namespace default -output credentials.json

# Export to stdout
go run export_credentials.go -namespace default

# Export from custom storage path
go run export_credentials.go -path ~/.tdl/data -namespace default -output backup.json
```

### Options

- `-path` - Path to tdl storage (default: `~/.tdl/data`)
- `-namespace` - Namespace to export (default: `default`)
- `-output` - Output file path (default: stdout)

### Output Format

```json
{
  "namespace": "default",
  "data": {
    "session": "...",
    "app": "desktop"
  }
}
```

### Use Cases

**1. E2E Testing**

Export credentials and use them in tests:

```bash
# Export
go run export_credentials.go -namespace default -output ../test/test.json

# Run tests with exported credentials
cd ..
TDL_TEST_CREDENTIALS_FILE=test/test.json go test ./test/...
```

## Security Warning

⚠️ **Exported credentials grant full access to your Telegram account!**

- Never commit credentials to version control
- Store with restricted permissions (`chmod 600`)
- Never share publicly
- Delete after use if temporary
