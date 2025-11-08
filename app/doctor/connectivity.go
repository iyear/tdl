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

	color.White("  Testing help.getConfig...")
	_, err := client.API().HelpGetConfig(ctx)
	if err != nil {
		color.Red("  [FAIL] Failed to access help.getConfig api: %v", err)
		return
	}
	color.White("  Server configuration retrieved")

	color.White("  Testing help.getNearestDc...")
	nearestDc, err := client.API().HelpGetNearestDC(ctx)
	if err != nil {
		color.Red("  [FAIL] Failed to access help.getNearestDc api: %v", err)
		return
	}
	color.White("  Nearest datacenter: DC%d (%s)", nearestDc.NearestDC, nearestDc.Country)

	color.White("  Testing langpack.getLanguages...")
	_, err = client.API().LangpackGetLanguages(ctx, "")
	if err != nil {
		color.Red("  [FAIL] Failed to access langpack.getLanguage api: %v", err)
		return
	}
	color.White("  Language pack accessible")

	color.Green("  [OK] Connectivity check completed successfully")
}
