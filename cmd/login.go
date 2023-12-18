package cmd

import (
	"fmt"
	"strings"

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

			// Legacy flag
			if code {
				return login.Code(logger.Named(cmd.Context(), "login"))
			}

			return login.Run(logger.Named(cmd.Context(), "login"), opts)
		},
	}

	const desktop = "desktop"

	cmd.Flags().VarP(&opts.Type, "type", "T", fmt.Sprintf("login mode: [%s]", strings.Join(login.TypeNames(), ", ")))
	cmd.Flags().StringVarP(&opts.Desktop, desktop, "d", "", "official desktop client path, and automatically find possible paths if empty")
	cmd.Flags().StringVarP(&opts.Passcode, "passcode", "p", "", "passcode for desktop client, keep empty if no passcode")

	// Deprecated
	cmd.Flags().BoolVar(&code, "code", false, "login with code, instead of importing session from desktop client")

	// completion and validation
	_ = cmd.MarkFlagDirname(desktop)
	_ = cmd.Flags().MarkDeprecated("code", "use `-T code` instead")

	return cmd
}
