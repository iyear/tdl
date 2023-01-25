package chat

import (
	"github.com/iyear/tdl/app/chat"
	"github.com/spf13/cobra"
	"time"
)

var expOpts = &chat.ExportOptions{}

var cmdExport = &cobra.Command{
	Use:   "export",
	Short: "export messages from (protected) chat for download",
	RunE: func(cmd *cobra.Command, args []string) error {
		// only support unique true value
		if expOpts.Time == expOpts.Msg {
			expOpts.Time, expOpts.Msg = true, false
		}

		if expOpts.To == 0 {
			expOpts.To = int(time.Now().Unix()) // it's also the latest message id(very big message id)
		}

		if expOpts.From > expOpts.To {
			expOpts.From, expOpts.To = expOpts.To, expOpts.From
		}

		return chat.Export(cmd.Context(), expOpts)
	},
}

func init() {
	cmdExport.Flags().StringVarP(&expOpts.Chat, "chat", "c", "", "chat id or domain")
	cmdExport.Flags().IntVar(&expOpts.From, "from", 0, "starting message")
	cmdExport.Flags().IntVar(&expOpts.To, "to", 0, "ending message, default value is NOW/LATEST")
	cmdExport.Flags().StringVarP(&expOpts.Output, "output", "o", "tdl-export.json", "output JSON file path")
	cmdExport.Flags().BoolVar(&expOpts.Time, "time", false, "the format for from&to is timestamp")
	cmdExport.Flags().BoolVar(&expOpts.Msg, "msg", false, "the format for from&to is message id")
}
