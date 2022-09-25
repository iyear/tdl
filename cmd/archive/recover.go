package archive

import (
	"github.com/iyear/tdl/app/archive"
	"github.com/spf13/cobra"
)

var file string

var CmdRecover = &cobra.Command{
	Use:   "recover",
	Short: "Recover your data",
	RunE: func(cmd *cobra.Command, args []string) error {
		return archive.Recover(cmd.Context(), file)
	},
}

func init() {
	CmdRecover.Flags().StringVarP(&file, "file", "f", "", "backup file path")
}
