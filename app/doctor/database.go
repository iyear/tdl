package doctor

import (
	"context"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/viper"

	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/key"
)

func checkDatabaseIntegrity(ctx context.Context, opts Options) {
	storage := opts.KV
	if storage == nil {
		color.Red("  [FAIL] Storage not initialized")
		return
	}

	hasIssues := false

	// Check storage type
	storageType := storage.Name()
	color.White("  Storage type: %s", storageType)

	// Get storage configuration
	storageConfig := viper.GetStringMapString(consts.FlagStorage)

	// Check if storage path exists and is accessible
	path := storageConfig["path"]
	if path == "" {
		color.Yellow("  [WARN] Storage path not configured")
		hasIssues = true
	} else {
		color.White("  Storage path: %s", path)

		// Check if path exists
		if info, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				color.White("  Path status: does not exist yet (will be created on first use)")
			} else {
				color.Red("  Path error: cannot access: %v", err)
				color.Red("[FAIL] Database integrity check failed")
				return
			}
		} else {
			// Show basic info about the storage
			if info.IsDir() {
				color.White("  Path type: directory")
			} else {
				color.White("  Path type: file (size: %d bytes)", info.Size())
			}
		}
	}

	// Check namespaces
	color.White("  Checking namespaces...")
	namespaces, err := storage.Namespaces()
	if err != nil {
		color.Red("  Namespace error: %v", err)
		color.Red("  [FAIL] Database integrity check failed")
		return
	}

	if len(namespaces) == 0 {
		color.Yellow("  [WARN] No namespaces found in storage")
		hasIssues = true
	} else {
		color.White("  Found %d namespace(s): %v", len(namespaces), namespaces)
	}

	// Check current namespace has required keys
	currentNS := viper.GetString(consts.FlagNamespace)
	if currentNS != "" {
		color.White("  Checking current namespace: %s", currentNS)

		nsStorage, err := storage.Open(currentNS)
		if err != nil {
			color.Yellow("  [WARN] Failed to open namespace: %v", err)
			hasIssues = true
		} else {
			nsHasIssues := checkNamespaceKeys(ctx, nsStorage)
			if nsHasIssues {
				hasIssues = true
			}
		}
	}

	// Final status
	if hasIssues {
		color.Yellow("  [WARN] Database check completed with warnings")
	} else {
		color.Green("  [OK] Database integrity check passed")
	}
}

func checkNamespaceKeys(ctx context.Context, storage interface{}) bool {
	type getter interface {
		Get(ctx context.Context, key string) ([]byte, error)
	}

	st, ok := storage.(getter)
	if !ok {
		return false
	}

	hasIssues := false

	// Check for session key
	if _, err := st.Get(ctx, "session"); err == nil {
		color.White("  - Session data: present")
	} else {
		color.White("  - Session data: missing (not logged in)")
		hasIssues = true
	}

	// Check for app key
	if data, err := st.Get(ctx, key.App()); err == nil {
		color.White("  - App config: %s", string(data))
	} else {
		color.White("  - App config: missing")
		hasIssues = true
	}

	return hasIssues
}
