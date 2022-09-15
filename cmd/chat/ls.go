package chat

import (
	"github.com/iyear/tdl/app/chat"
	"github.com/spf13/cobra"
)

var cmdList = &cobra.Command{
	Use:     "ls",
	Short:   "List your chats",
	Example: "tdl chat ls -n iyear --proxy socks5://localhost:1080",
	RunE: func(cmd *cobra.Command, args []string) error {
		return chat.List(cmd.Context(), cmd.Flag("ns").Value.String(), cmd.Flag("proxy").Value.String())
	},
}
