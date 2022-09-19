package dl

import (
	"github.com/iyear/tdl/pkg/consts"
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

	CmdDL.PersistentFlags().IntVarP(&partSize, consts.FlagPartSize, "s", 512*1024, "part size for downloading, max is 512*1024")
	CmdDL.PersistentFlags().IntVarP(&threads, consts.FlagThreads, "t", 8, "threads for downloading one item")
	CmdDL.PersistentFlags().IntVarP(&limit, consts.FlagLimit, "l", 2, "max number of concurrent tasks")
}
