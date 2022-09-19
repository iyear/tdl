package up

import (
	"github.com/iyear/tdl/app/up"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
)

var (
	partSize int
	threads  int
	limit    int

	paths    []string
	excludes []string
)

var Cmd = &cobra.Command{
	Use:     "up",
	Aliases: []string{"upload"},
	Short:   "Upload anything to Telegram",
	Example: "tdl up -n iyear --proxy socks5://localhost:1080 -p /path/to/file -p /path -e .so -e .tmp -s 262144 -t 16 -l 3",
	RunE: func(cmd *cobra.Command, args []string) error {
		return up.Run(cmd.Context(), cmd.Flag("ns").Value.String(), cmd.Flag("proxy").Value.String(), partSize, threads, limit, paths, excludes)
	},
}

func init() {
	Cmd.PersistentFlags().IntVarP(&partSize, consts.FlagPartSize, "s", 512*1024, "part size for uploading, max is 512*1024")
	Cmd.PersistentFlags().IntVarP(&threads, consts.FlagThreads, "t", 8, "threads for uploading one item")
	Cmd.PersistentFlags().IntVarP(&limit, consts.FlagLimit, "l", 2, "max number of concurrent tasks")

	Cmd.Flags().StringSliceVarP(&paths, consts.FlagUpPath, "p", []string{}, "dirs or files")
	Cmd.Flags().StringSliceVarP(&excludes, consts.FlagUpExcludes, "e", []string{}, "exclude the specified file extensions")
}
