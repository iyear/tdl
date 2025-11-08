package cmd

import (
	"github.com/spf13/cobra"

	"github.com/iyear/tdl/app/doctor"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/tclient"
)

func NewDoctor() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "doctor",
		Short:   "Diagnose Telegram connection and configuration issues",
		GroupID: groupTools.ID,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := logctx.Named(cmd.Context(), "doctor")

			// Get storage
			storage := kv.From(ctx)

			// Create client options
			o, err := tOptions(ctx)
			if err != nil {
				return err
			}

			// Create client
			client, err := tclient.New(ctx, o, false)
			if err != nil {
				return err
			}

			// Run doctor
			return doctor.Run(ctx, doctor.Options{
				KV:     storage,
				Client: client,
			})
		},
	}

	return cmd
}
