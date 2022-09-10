package dl

import (
	"github.com/spf13/cobra"
)

var (
	partSize int
	threads  int
	limit    int
)

var CmdDL = &cobra.Command{
	Use:     "dl",
	Aliases: []string{"download"},
	Short:   "Download what you want",
	Example: "tdl dl -h",
}

func init() {
	CmdDL.AddCommand(cmdURL)

	CmdDL.PersistentFlags().IntVarP(&partSize, "part-size", "s", 512*1024, "part size for download, max is 512*1024")
	CmdDL.PersistentFlags().IntVarP(&threads, "threads", "t", 8, "threads for downloading one item")
	CmdDL.PersistentFlags().IntVarP(&limit, "limit", "l", 2, "max number of concurrent tasks")
}
