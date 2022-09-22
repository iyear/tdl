package dl

import (
	"github.com/iyear/tdl/app/dl"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
)

var (
	urls []string
)

var Cmd = &cobra.Command{
	Use:     "dl",
	Aliases: []string{"download"},
	Short:   "Download anything from Telegram (protected) chat",
	Example: "tdl dl -n iyear --proxy socks5://localhost:1080 -u https://t.me/tdl/1 -u https://t.me/tdl/2 -s 262144 -t 16 -l 3",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dl.Run(cmd.Context(), urls)
	},
}

func init() {
	Cmd.Flags().StringSliceVarP(&urls, consts.FlagDlUrl, "u", []string{}, "telegram message links to be downloaded")
}
