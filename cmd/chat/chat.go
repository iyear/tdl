package chat

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "chat",
	Short: "A set of chat tools",
}

func init() {
	Cmd.AddCommand(cmdList, cmdExport)
}
