package login

import (
	"github.com/fatih/color"
	"github.com/iyear/tdl/app/login"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	code bool
	opts login.Options
)

var Cmd = &cobra.Command{
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

func init() {
	Cmd.Flags().StringVarP(&opts.Desktop, consts.FlagLoginDesktop, "d", "", "official desktop client path, and automatically find possible paths if empty")
	Cmd.Flags().StringVarP(&opts.Passcode, consts.FlagLoginPasscode, "p", "", "passcode for desktop client, keep empty if no passcode")
	Cmd.Flags().BoolVar(&code, consts.FlagLoginCode, false, "login with code, instead of importing session from desktop client")
}
