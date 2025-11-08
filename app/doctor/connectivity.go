package doctor

import (
	"context"

	"github.com/fatih/color"
)

func checkConnectivity(ctx context.Context, opts Options) {
	client := opts.Client
	if client == nil {
		color.Yellow("  [WARN] Client not provided, skipping connectivity check")
		return
	}

	// Simple ping to check connectivity
	_, err := client.API().HelpGetConfig(ctx)
	if err != nil {
		color.Red("  [FAIL] Failed to connect to Telegram: %v", err)
		return
	}

	color.Green("  [OK] Successfully connected to Telegram server")
}
