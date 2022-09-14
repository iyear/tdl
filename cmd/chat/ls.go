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
		proxy, err := cmd.Flags().GetString("proxy")
		if err != nil {
			return err
		}

		ns, err := cmd.Flags().GetString("ns")
		if err != nil {
			return err
		}

		return chat.List(cmd.Context(), ns, proxy)
	},
}
