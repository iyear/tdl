package login

import (
	"context"
	"fmt"
	"github.com/iyear/tdl/app/login"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
)

var Cmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to Telegram",
	Example: "tdl login",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
		defer cancel()

		proxy, err := cmd.Flags().GetString("proxy")
		if err != nil {
			return err
		}

		ns, err := cmd.Flags().GetString("ns")
		if err != nil {
			return err
		}
		if err := login.Run(ctx, ns, proxy); err != nil {
			return fmt.Errorf("login failed: %v", err)
		}
		return nil
	},
}
