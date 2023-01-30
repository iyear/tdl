package chat

import (
	"github.com/iyear/tdl/app/chat"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/spf13/cobra"
)

var cmdList = &cobra.Command{
	Use:   "ls",
	Short: "List your chats",
	RunE: func(cmd *cobra.Command, args []string) error {
		return chat.List(logger.Named(cmd.Context(), "ls"))
	},
}
