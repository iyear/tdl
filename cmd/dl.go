package cmd

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/iyear/tdl/app/dl"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/pkg/consts"
)

func NewDownload() *cobra.Command {
	var opts dl.Options

	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"dl"},
		Short:   "Download anything from Telegram (protected) chat",
		GroupID: groupTools.ID,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(opts.URLs) == 0 && len(opts.Files) == 0 {
				return fmt.Errorf("no urls or files provided")
			}

			opts.Template = viper.GetString(consts.FlagDlTemplate)

			return tRun(cmd.Context(), func(ctx context.Context, c *telegram.Client, kvd storage.Storage) error {
				return dl.Run(logctx.Named(ctx, "dl"), c, kvd, opts)
			})
		},
	}

	const (
		file      = "file"
		dir       = "dir"
		include   = "include"
		exclude   = "exclude"
		_continue = "continue"
		restart   = "restart"
	)

	cmd.Flags().StringSliceVarP(&opts.URLs, "url", "u", []string{}, "telegram message links")
	cmd.Flags().StringSliceVarP(&opts.Files, file, "f", []string{}, "official client exported files")

	cmd.Flags().String(consts.FlagDlTemplate, `{{ .DialogID }}_{{ .MessageID }}_{{ filenamify .FileName }}`, "download file name template")

	cmd.Flags().StringSliceVarP(&opts.Include, include, "i", []string{}, "include the specified file extensions, and only judge by file name, not file MIME. Example: -i mp4,mp3")
	cmd.Flags().StringSliceVarP(&opts.Exclude, exclude, "e", []string{}, "exclude the specified file extensions, and only judge by file name, not file MIME. Example: -e png,jpg")

	cmd.Flags().StringVarP(&opts.Dir, dir, "d", "downloads", "specify the download directory. If the directory does not exist, it will be created automatically")
	cmd.Flags().BoolVar(&opts.RewriteExt, "rewrite-ext", false, "rewrite file extension according to file header MIME")
	// do not match extension, because some files' extension is corrected by --rewrite-ext flag
	cmd.Flags().BoolVar(&opts.SkipSame, "skip-same", false, "skip files with the same name(without extension) and size")
	cmd.Flags().BoolVar(&opts.SkipName, "skip files that already exist by comparing their filenames with the exported JSON")

	cmd.Flags().BoolVar(&opts.Desc, "desc", false, "download files from the newest to the oldest ones (may affect resume download)")
	cmd.Flags().BoolVar(&opts.Takeout, "takeout", false, "takeout sessions let you export data from your account with lower flood wait limits.")
	cmd.Flags().BoolVar(&opts.Group, "group", false, "auto detect grouped message and download all of them")

	// resume flags, if both false then ask user
	cmd.Flags().BoolVar(&opts.Continue, _continue, false, "continue the last download directly")
	cmd.Flags().BoolVar(&opts.Restart, restart, false, "restart the last download directly")

	// serve flags
	cmd.Flags().BoolVar(&opts.Serve, "serve", false, "serve the media files as a http server instead of downloading them with built-in downloader")
	cmd.Flags().IntVar(&opts.Port, "port", 8080, "http server port")

	_ = viper.BindPFlag(consts.FlagDlTemplate, cmd.Flags().Lookup(consts.FlagDlTemplate))

	// completion and validation
	_ = cmd.RegisterFlagCompletionFunc(file, completeExtFiles("json"))
	_ = cmd.MarkFlagDirname(dir)
	cmd.MarkFlagsMutuallyExclusive(include, exclude)
	cmd.MarkFlagsMutuallyExclusive(_continue, restart)

	return cmd
}
