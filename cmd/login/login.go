package login

import (
	"github.com/iyear/tdl/app/login"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to Telegram",
	Example: "tdl login -n iyear --proxy socks5://localhost:1080",
	RunE: func(cmd *cobra.Command, args []string) error {
		return login.Run(cmd.Context(), cmd.Flag("ns").Value.String(), cmd.Flag("proxy").Value.String())
	},
}
