package login

import (
	"github.com/fatih/color"
	"github.com/iyear/tdl/app/login"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/spf13/cobra"
)

var (
	desktop, passcode string
)

var Cmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Telegram",
	RunE: func(cmd *cobra.Command, args []string) error {
		color.Yellow("WARN: If data exists in the namespace, data will be overwritten")

		if desktop != "" {
			return login.Desktop(cmd.Context(), desktop, passcode)
		}
		return login.Code(logger.Named(cmd.Context(), "login"))
	},
}

func init() {
	Cmd.Flags().StringVarP(&desktop, consts.FlagLoginDesktop, "d", "", "official desktop client path, import session from it")
	Cmd.Flags().StringVarP(&passcode, consts.FlagLoginPasscode, "p", "", "passcode for desktop client, keep empty if no passcode")
}
