package dl

import (
	"github.com/iyear/tdl/app/dl/dlurl"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
)

var (
	urls []string
)

var cmdURL = &cobra.Command{
	Use:     "url",
	Short:   "Download in url mode",
	Example: "tdl dl url -n iyear --proxy socks5://localhost:1080 -u https://t.me/tdl/1 -u https://t.me/tdl/2 -s 262144 -t 16 -l 3",
	RunE: func(cmd *cobra.Command, args []string) error {
		return dlurl.Run(cmd.Context(), urls)
	},
}

func init() {
	cmdURL.Flags().StringSliceVarP(&urls, consts.FlagDlUrls, "u", []string{}, "telegram message links to be downloaded")
}
