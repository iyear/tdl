package chat

import (
	"github.com/iyear/tdl/app/chat"
	"github.com/spf13/cobra"
	"time"
)

var (
	_chat    string
	from, to int
	output   string
)

var cmdExport = &cobra.Command{
	Use:   "export",
	Short: "export messages from (protected) chat for download",
	RunE: func(cmd *cobra.Command, args []string) error {
		if to == 0 {
			to = int(time.Now().Unix())
		}

		return chat.Export(cmd.Context(), _chat, from, to, output)
	},
}

func init() {
	cmdExport.Flags().StringVarP(&_chat, "chat", "c", "", "chat id or domain")
	cmdExport.Flags().IntVar(&from, "from", 0, "timestamp of the starting message")
	cmdExport.Flags().IntVar(&to, "to", 0, "timestamp of the ending message, default value is NOW")
	cmdExport.Flags().StringVarP(&output, "output", "o", "tdl-export.json", "output JSON file path")
}
