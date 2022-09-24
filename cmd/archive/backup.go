package archive

import (
	"fmt"
	"github.com/iyear/tdl/app/archive"
	"github.com/spf13/cobra"
	"time"
)

var dst string

var CmdBackup = &cobra.Command{
	Use:     "backup",
	Short:   "Backup your data",
	Example: "tdl backup -d tdl-backup-iyear.zip",
	RunE: func(cmd *cobra.Command, args []string) error {
		if dst == "" {
			dst = fmt.Sprintf("tdl-backup-%s.zip", time.Now().Format("2006-01-02-15_04_05"))
		}

		return archive.Backup(cmd.Context(), dst)
	},
}

func init() {
	CmdBackup.Flags().StringVarP(&dst, "dst", "d", "", "destination file path. Default: tdl-backup-<time>.zip")
}
