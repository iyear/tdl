package doctor

import (
	"context"
	"fmt"

	"github.com/fatih/color"
)

func checkLoginStatus(ctx context.Context, opts Options) {
	client := opts.Client
	if client == nil {
		color.Yellow("  [WARN] Client not provided, skipping login check")
		return
	}

	status, err := client.Auth().Status(ctx)
	if err != nil {
		color.Red("  [FAIL] Failed to check login status: %v", err)
		return
	}

	if !status.Authorized {
		color.Yellow("  [WARN] Not logged in. Please run 'tdl login' first.")
		return
	}

	color.Green("  [OK] Login status: Authorized")

	// Get user info
	user, err := client.Self(ctx)
	if err != nil {
		color.Yellow("  [WARN] Failed to get user info: %v", err)
		return
	}

	name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
	if user.Username != "" {
		color.White("  Account: %s (@%s)", name, user.Username)
	} else {
		color.White("  Account: %s", name)
	}
	color.White("  User ID: %d", user.ID)
	if user.Phone != "" {
		color.White("  Phone: %s", user.Phone)
	}
}
