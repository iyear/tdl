# E2E Testing

End-to-end tests for tdl using exported credentials.

## Usage

```bash
# 1. Export credentials
cd tools
go run export_credentials.go -namespace default -output ../test/credentials.json

# 2. Run tests
cd ..
TDL_TEST_CREDENTIALS_FILE=$(pwd)/test/credentials.json go test ./test/... -v
```

## GitHub Actions

E2E tests run via `.github/workflows/e2e.yml`:

- **Trigger**: Comment `/e2e` on PR (requires write permission) or manual run
- **Credentials**: Stored in `secrets.TG_CREDENTIALS`

## Security

- Never commit credentials to version control
