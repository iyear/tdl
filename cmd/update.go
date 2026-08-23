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
		Long: `Update tdl in place by downloading a release from GitHub,
verifying its checksum and replacing the current binary.

Examples:
  tdl update                     # install latest verified release
  tdl update --check             # report availability only
  tdl update --version v0.20.4   # install an exact release
  tdl update --force             # reinstall / allow downgrade`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return update.Run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "update without confirmation prompt")
	cmd.Flags().BoolVar(&opts.Check, "check", false, "report whether an update is available without downloading")
	cmd.Flags().StringVar(&opts.Target, "version", "", "install a specific release tag instead of the latest (e.g. v0.20.4)")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "reinstall even when up to date; also allows downgrades")

	return cmd
}
