package cmd

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/iyear/tdl/app/login"
	"github.com/iyear/tdl/pkg/logger"
)

func NewLogin() *cobra.Command {
	var (
		code bool
		opts login.Options
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to Telegram",
		RunE: func(cmd *cobra.Command, args []string) error {
			color.Yellow("WARN: If data exists in the namespace, data will be overwritten")

			if code {
				return login.Code(logger.Named(cmd.Context(), "login"))
			}

			return login.Desktop(cmd.Context(), &opts)
		},
	}

	const (
		desktop = "desktop"
	)

	cmd.Flags().StringVarP(&opts.Desktop, desktop, "d", "", "official desktop client path, and automatically find possible paths if empty")
	cmd.Flags().StringVarP(&opts.Passcode, "passcode", "p", "", "passcode for desktop client, keep empty if no passcode")
	cmd.Flags().BoolVar(&code, "code", false, "login with code, instead of importing session from desktop client")

	// completion and validation
	_ = cmd.MarkFlagDirname(desktop)

	return cmd
}
