package dl

import (
	"context"
	"fmt"
	"github.com/iyear/tdl/app/dl/dlurl"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
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
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
		defer cancel()

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
			if err := dlurl.Run(ctx, ns, proxy, partSize, threads, limit, urls); err != nil {
				return fmt.Errorf("download failed: %v", err)
			}
			return nil
		}

		return fmt.Errorf("mode %s is not supported", mode)

	},
}

func init() {
	Cmd.Flags().IntVarP(&partSize, "part-size", "s", 512*1024, "")
	Cmd.Flags().IntVarP(&threads, "threads", "t", 10, "")
	Cmd.Flags().IntVarP(&limit, "limit", "l", 2, "")

	// url mode
	Cmd.Flags().StringSliceVarP(&urls, "url", "u", make([]string, 0), "")

	Cmd.Flags().StringVarP(&mode, "mode", "m", "url", "")
}
