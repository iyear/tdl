package chat

import (
	"github.com/iyear/tdl/app/chat"
	"github.com/spf13/cobra"
	"time"
)

var (
	_chat      string
	from, to   int
	output     string
	_time, msg bool
)

var cmdExport = &cobra.Command{
	Use:   "export",
	Short: "export messages from (protected) chat for download",
	RunE: func(cmd *cobra.Command, args []string) error {
		// only support unique true value
		if _time == msg {
			_time, msg = true, false
		}

		if to == 0 {
			to = int(time.Now().Unix()) // it's also the latest message id(very big message id)
		}

		if from > to {
			from, to = to, from
		}

		return chat.Export(cmd.Context(), _chat, from, to, output, _time, msg)
	},
}

func init() {
	cmdExport.Flags().StringVarP(&_chat, "chat", "c", "", "chat id or domain")
	cmdExport.Flags().IntVar(&from, "from", 0, "starting message")
	cmdExport.Flags().IntVar(&to, "to", 0, "ending message, default value is NOW/LATEST")
	cmdExport.Flags().StringVarP(&output, "output", "o", "tdl-export.json", "output JSON file path")
	cmdExport.Flags().BoolVar(&_time, "time", false, "the format for from&to is timestamp")
	cmdExport.Flags().BoolVar(&msg, "msg", false, "the format for from&to is message id")
}
