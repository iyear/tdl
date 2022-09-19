package up

import (
	"github.com/iyear/tdl/app/up"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
)

var (
	paths    []string
	excludes []string
)

var Cmd = &cobra.Command{
	Use:     "up",
	Aliases: []string{"upload"},
	Short:   "Upload anything to Telegram",
	Example: "tdl up -n iyear --proxy socks5://localhost:1080 -p /path/to/file -p /path -e .so -e .tmp -s 262144 -t 16 -l 3",
	RunE: func(cmd *cobra.Command, args []string) error {
		return up.Run(cmd.Context(), paths, excludes)
	},
}

func init() {
	Cmd.Flags().StringSliceVarP(&paths, consts.FlagUpPath, "p", []string{}, "dirs or files")
	Cmd.Flags().StringSliceVarP(&excludes, consts.FlagUpExcludes, "e", []string{}, "exclude the specified file extensions")
}
