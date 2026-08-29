package cmd

import (
	"github.com/spf13/cobra"

	"github.com/iyear/tdl/app/update"
)

func NewUpdate() *cobra.Command {
	var opts update.Options

	cmd := &cobra.Command{
		Use:     "update",
		Aliases: []string{"upgrade"},
		Short:   "Update tdl to the latest version",
		GroupID: groupTools.ID,
		Long: `Update tdl in place by downloading a release from GitHub,
verifying its checksum and replacing the current binary.

Examples:
  tdl update                     # install latest verified release
  tdl update --dry-run           # report availability only
  tdl update --version v0.20.4   # install an exact release
  tdl update --force             # reinstall / allow downgrade`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return update.Run(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Yes, "yes", "y", false, "update without confirmation prompt")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "report whether an update is available without downloading")
	cmd.Flags().StringVarP(&opts.Target, "version", "v", "", "install a specific release tag instead of the latest (e.g. v0.20.4)")
	cmd.Flags().BoolVarP(&opts.Force, "force", "f", false, "reinstall even when up to date; also allows downgrades")

	return cmd
}
