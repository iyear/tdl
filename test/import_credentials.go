package test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-faster/errors"

	"github.com/iyear/tdl/pkg/kv"
)

// CredentialImport represents imported credentials
type CredentialImport struct {
	Namespace string            `json:"namespace"`
	Data      map[string][]byte `json:"data"`
}

// SetupWithImportedCredentials sets up e2e tests with imported credentials from file
func SetupWithImportedCredentials(ctx context.Context) (account string, sessionFile string, _ error) {
	credFile := os.Getenv("TDL_TEST_CREDENTIALS_FILE")
	if credFile == "" {
		return "", "", errors.New("TDL_TEST_CREDENTIALS_FILE is required")
	}

	log.Printf("Loading credentials from file: %s", credFile)
	return setupFromFile(ctx, credFile)
}

func setupFromFile(ctx context.Context, credFile string) (string, string, error) {
	data, err := os.ReadFile(credFile)
	if err != nil {
		return "", "", errors.Wrap(err, "read credentials file")
	}

	return setupFromJSON(ctx, string(data))
}

func setupFromJSON(ctx context.Context, credJSON string) (string, string, error) {
	var cred CredentialImport
	if err := json.Unmarshal([]byte(credJSON), &cred); err != nil {
		return "", "", errors.Wrap(err, "unmarshal credentials")
	}

	if cred.Namespace == "" {
		return "", "", errors.New("namespace is empty in credentials")
	}

	if len(cred.Data) == 0 {
		return "", "", errors.New("no data in credentials")
	}

	// Verify session data exists
	if _, ok := cred.Data["session"]; !ok {
		return "", "", errors.New("session data not found in credentials")
	}

	// Create temporary session file
	account := fmt.Sprintf("e2e-imported-%s", cred.Namespace)
	sessionFile := filepath.Join(os.TempDir(), "tdl-e2e", account)

	// Create session directory
	if err := os.MkdirAll(filepath.Dir(sessionFile), 0o755); err != nil {
		return "", "", errors.Wrap(err, "create session directory")
	}

	log.Printf("Creating session file: %s", sessionFile)

	// Create KV storage
	kvd, err := kv.New(kv.DriverFile, map[string]any{
		"path": sessionFile,
	})
	if err != nil {
		return "", "", errors.Wrap(err, "create kv storage")
	}
	defer kvd.Close()

	// Open namespace
	stg, err := kvd.Open(account)
	if err != nil {
		return "", "", errors.Wrap(err, "open namespace")
	}

	// Import all data
	for key, value := range cred.Data {
		if err := stg.Set(ctx, key, value); err != nil {
			return "", "", errors.Wrapf(err, "set key: %s", key)
		}
		log.Printf("Imported key: %s (%d bytes)", key, len(value))
	}

	log.Printf("Successfully imported credentials for namespace: %s", cred.Namespace)
	log.Printf("Account: %s, Session file: %s", account, sessionFile)

	return account, sessionFile, nil
}

// ValidateCredentials validates that the credentials contain necessary data
func ValidateCredentials(credJSON string) error {
	var cred CredentialImport
	if err := json.Unmarshal([]byte(credJSON), &cred); err != nil {
		return errors.Wrap(err, "unmarshal credentials")
	}

	if cred.Namespace == "" {
		return errors.New("namespace is empty")
	}

	if _, ok := cred.Data["session"]; !ok {
		return errors.New("session data not found")
	}

	// Optional but recommended
	if _, ok := cred.Data["app"]; !ok {
		log.Println("Warning: app type not found in credentials")
	}

	return nil
}
