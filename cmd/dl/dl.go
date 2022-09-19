package dl

import (
	"github.com/spf13/cobra"
)

var CmdDL = &cobra.Command{
	Use:     "dl",
	Aliases: []string{"download"},
	Short:   "Download what you want",
	Example: "tdl dl -h",
}

func init() {
	CmdDL.AddCommand(cmdURL)
}
