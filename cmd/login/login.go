package login

import (
	"fmt"
	"github.com/iyear/tdl/app/login"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to Telegram",
	Example: "tdl login",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := cmd.Flags().GetString("proxy")
		if err != nil {
			return err
		}

		ns, err := cmd.Flags().GetString("ns")
		if err != nil {
			return err
		}

		if err := login.Run(cmd.Context(), ns, proxy); err != nil {
			return fmt.Errorf("login failed: %v", err)
		}
		return nil
	},
}
