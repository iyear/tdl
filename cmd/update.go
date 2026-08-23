package cmd

import (
	"github.com/spf13/cobra"

	"github.com/iyear/tdl/app/update"
)

func NewUpdate() *cobra.Command {
	var opts update.Options

	cmd := &cobra.Command{
		Use:     "update",
		Short:   "Update tdl to the latest version",
		GroupID: groupTools.ID,
		RunE: func(cmd *cobra.Command, args []string) error {
			return update.Run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "update without confirmation prompt")

	return cmd
}
