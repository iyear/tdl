package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/iyear/tdl/pkg/kv"
)

// CredentialExport represents exported credentials
type CredentialExport struct {
	Namespace string            `json:"namespace"`
	Data      map[string][]byte `json:"data"`
}

func main() {
	var (
		storagePath string
		namespace   string
		outputFile  string
	)

	homeDir, _ := os.UserHomeDir()
	defaultPath := filepath.Join(homeDir, ".tdl", "data")

	flag.StringVar(&storagePath, "path", defaultPath, "Path to tdl storage")
	flag.StringVar(&namespace, "namespace", "default", "Namespace to export")
	flag.StringVar(&outputFile, "output", "", "Output file (default: stdout)")
	flag.Parse()

	if err := exportCredentials(storagePath, namespace, outputFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func exportCredentials(storagePath, namespace, outputFile string) error {
	// Detect storage type
	driver, storagePath, err := detectStorageType(storagePath)
	if err != nil {
		return err
	}

	// Open storage
	var kvStorage kv.Storage
	switch driver {
	case kv.DriverFile:
		kvStorage, err = kv.New(kv.DriverFile, map[string]any{"path": storagePath})
	case kv.DriverBolt:
		kvStorage, err = kv.New(kv.DriverBolt, map[string]any{"path": storagePath})
	default:
		return fmt.Errorf("unsupported storage type: %s", driver)
	}
	if err != nil {
		return fmt.Errorf("open storage: %w", err)
	}
	defer kvStorage.Close()

	// Export all data
	meta, err := kvStorage.MigrateTo()
	if err != nil {
		return fmt.Errorf("export data: %w", err)
	}

	nsData, ok := meta[namespace]
	if !ok {
		return fmt.Errorf("namespace %s not found", namespace)
	}

	export := CredentialExport{
		Namespace: namespace,
		Data:      nsData,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	// Write to file or stdout
	if outputFile != "" {
		if err := os.WriteFile(outputFile, jsonData, 0o600); err != nil {
			return fmt.Errorf("write file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Credentials exported to: %s\n", outputFile)
		fmt.Fprintf(os.Stderr, "⚠️  Keep this file secure! It contains your account credentials.\n")
	} else {
		fmt.Println(string(jsonData))
	}

	return nil
}

func detectStorageType(path string) (kv.Driver, string, error) {
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return kv.DriverFile, path, nil
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		return kv.DriverBolt, path, nil
	}
	return "", "", fmt.Errorf("cannot determine storage type for: %s", path)
}
