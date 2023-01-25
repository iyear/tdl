package up

import (
	"github.com/iyear/tdl/app/up"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
)

var opts = &up.Options{}

var Cmd = &cobra.Command{
	Use:     "up",
	Aliases: []string{"upload"},
	Short:   "Upload anything to Telegram",
	RunE: func(cmd *cobra.Command, args []string) error {
		return up.Run(cmd.Context(), opts)
	},
}

func init() {
	Cmd.Flags().StringVarP(&opts.Chat, "chat", "c", "", "chat id or domain, and empty means 'Saved Messages'")
	Cmd.Flags().StringSliceVarP(&opts.Paths, consts.FlagUpPath, "p", []string{}, "dirs or files")
	Cmd.Flags().StringSliceVarP(&opts.Excludes, consts.FlagUpExcludes, "e", []string{}, "exclude the specified file extensions")
}
