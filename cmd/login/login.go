package login

import (
	"github.com/fatih/color"
	"github.com/iyear/tdl/app/login"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
)

var (
	desktop string
)

var Cmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to Telegram",
	Example: "tdl login -n iyear --proxy socks5://localhost:1080",
	RunE: func(cmd *cobra.Command, args []string) error {
		color.Yellow("WARN: If data exists in the namespace, data will be overwritten")

		if desktop != "" {
			return login.Desktop(cmd.Context(), desktop)
		}
		return login.Code(cmd.Context())
	},
}

func init() {
	Cmd.Flags().StringVarP(&desktop, consts.FlagLoginDesktop, "d", "", "Official desktop client path, import session from it")
}
