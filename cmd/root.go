package cmd

import (
	"github.com/fatih/color"
	"github.com/iyear/tdl/cmd/login"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"path/filepath"
)

var cmd = &cobra.Command{
	Use:               "tdl",
	Short:             "Telegram downloader, but not only a downloader",
	Example:           "tdl -h",
	DisableAutoGenTag: true,
	SilenceErrors:     true,
	SilenceUsage:      true,
}

func init() {
	cmd.AddCommand(login.Cmd)
	cmd.PersistentFlags().String("proxy", "", "")
	cmd.PersistentFlags().StringP("ns", "n", "", "namespace")

	if err := doc.GenMarkdownTree(cmd, filepath.Join(consts.DocsPath, "command")); err != nil {
		color.Red("generate cmd docs failed: %v", err)
		return
	}
}

func Execute() {
	if err := cmd.Execute(); err != nil {
		color.Red("%v", err)
	}
}
