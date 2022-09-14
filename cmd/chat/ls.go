package chat

import (
	"github.com/iyear/tdl/app/chat"
	"github.com/spf13/cobra"
)

var cmdList = &cobra.Command{
	Use:     "ls",
	Short:   "List your all chats with info",
	Example: "",
	RunE: func(cmd *cobra.Command, args []string) error {
		return chat.List(cmd.Context(), cmd.Flag("ns").Value.String(), cmd.Flag("proxy").Value.String())
	},
}
