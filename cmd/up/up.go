package up

import (
	"github.com/iyear/tdl/app/up"
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
	Example: "tdl up -h",
	RunE: func(cmd *cobra.Command, args []string) error {
		return up.Run(cmd.Context(), cmd.Flag("ns").Value.String(), cmd.Flag("proxy").Value.String(), partSize, threads, limit, paths, excludes)
	},
}

func init() {
	Cmd.PersistentFlags().IntVarP(&partSize, "part-size", "s", 512*1024, "part size for uploading, max is 512*1024")
	Cmd.PersistentFlags().IntVarP(&threads, "threads", "t", 8, "threads for uploading one item")
	Cmd.PersistentFlags().IntVarP(&limit, "limit", "l", 2, "max number of concurrent tasks")

	Cmd.Flags().StringSliceVarP(&paths, "path", "p", []string{}, "it can be dirs or files")
	Cmd.Flags().StringSliceVarP(&excludes, "excludes", "e", []string{}, "exclude the specified file extensions")
}
