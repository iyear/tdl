package dl

import (
	"github.com/iyear/tdl/app/dl/dlchat"
	"github.com/spf13/cobra"
)

var (
	chat     string
	from, to int
)

var cmdChat = &cobra.Command{
	Use:     "chat",
	Short:   "Download in chat mode",
	Example: "tdl dl chat ",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := cmd.Flags().GetString("proxy")
		if err != nil {
			return err
		}

		ns, err := cmd.Flags().GetString("ns")
		if err != nil {
			return err
		}

		return dlchat.Run(cmd.Context(), ns, proxy, partSize, threads, limit)
	},
}

func init() {
	cmdChat.Flags().StringVarP(&chat, "chat", "c", "", "accp")
	cmdChat.Flags().IntVar(&from, "from", 0)
}
