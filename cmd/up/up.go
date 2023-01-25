package up

import (
	"github.com/iyear/tdl/app/up"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
)

var (
	chat     string
	paths    []string
	excludes []string
)

var Cmd = &cobra.Command{
	Use:     "up",
	Aliases: []string{"upload"},
	Short:   "Upload anything to Telegram",
	RunE: func(cmd *cobra.Command, args []string) error {
		return up.Run(cmd.Context(), chat, paths, excludes)
	},
}

func init() {
	Cmd.Flags().StringVarP(&chat, "chat", "c", "", "chat id or domain, and empty means 'Saved Messages'")
	Cmd.Flags().StringSliceVarP(&paths, consts.FlagUpPath, "p", []string{}, "dirs or files")
	Cmd.Flags().StringSliceVarP(&excludes, consts.FlagUpExcludes, "e", []string{}, "exclude the specified file extensions")
}
