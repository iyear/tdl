package dl

import (
	"github.com/iyear/tdl/app/dl/dlurl"
	"github.com/spf13/cobra"
)

var (
	urls []string
)

var cmdURL = &cobra.Command{
	Use:     "url",
	Short:   "Download in url mode",
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

		return dlurl.Run(cmd.Context(), ns, proxy, partSize, threads, limit, urls)
	},
}

func init() {
	cmdURL.Flags().StringSliceVarP(&urls, "urls", "u", []string{}, "telegram message links to be downloaded")
}
