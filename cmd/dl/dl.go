package dl

import (
	"fmt"
	"github.com/iyear/tdl/app/dl/dlurl"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
)

var (
	partSize int
	threads  int
	limit    int

	// url mode
	urls []string

	mode string
)

var Cmd = &cobra.Command{
	Use:     "dl",
	Aliases: []string{"download"},
	Short:   "Download what you want",
	Example: "tdl dl",
	RunE: func(cmd *cobra.Command, args []string) error {
		proxy, err := cmd.Flags().GetString("proxy")
		if err != nil {
			return err
		}

		ns, err := cmd.Flags().GetString("ns")
		if err != nil {
			return err
		}

		switch mode {
		case consts.DownloadModeURL:
			if err := dlurl.Run(cmd.Context(), ns, proxy, partSize, threads, limit, urls); err != nil {
				return fmt.Errorf("download failed: %v", err)
			}
			return nil
		}

		return fmt.Errorf("mode %s is not supported", mode)

	},
}

func init() {
	Cmd.Flags().IntVarP(&partSize, "part-size", "s", 512*1024, "part size for download, max is 512*1024")
	Cmd.Flags().IntVarP(&threads, "threads", "t", 8, "threads for downloading one item")
	Cmd.Flags().IntVarP(&limit, "limit", "l", 2, "max number of concurrent tasks")

	// url mode
	Cmd.Flags().StringSliceVarP(&urls, "url", "u", make([]string, 0), "array of message links to be downloaded")

	Cmd.Flags().StringVarP(&mode, "mode", "m", "", "mode for download")
}
