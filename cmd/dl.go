package cmd

import (
	"errors"
	"fmt"
	"github.com/iyear/tdl/app/dl"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

func NewDownload() *cobra.Command {
	var opts dl.Options

	cmd := &cobra.Command{
		Use:     "download",
		Aliases: []string{"dl"},
		Short:   "Download anything from Telegram (protected) chat",
		RunE: func(cmd *cobra.Command, args []string) error {
			// only one of include and exclude can be specified
			if len(opts.Include) > 0 && len(opts.Exclude) > 0 {
				return errors.New("only one of `include` and `exclude` can be specified")
			}

			// only one of continue and restart can be specified
			if opts.Continue && opts.Restart {
				return errors.New("only one of `continue` and `restart` can be specified, or none of them")
			}

			opts.Template = viper.GetString(consts.FlagDlTemplate)
			return dl.Run(logger.Named(cmd.Context(), "dl"), &opts)
		},
	}

	cmd.Flags().StringSliceVarP(&opts.URLs, consts.FlagDlUrl, "u", []string{}, "telegram message links")
	cmd.Flags().StringSliceVarP(&opts.Files, consts.FlagDlFile, "f", []string{}, "official client exported files")

	// generate default replacer
	builder := strings.Builder{}
	chars := []string{`/`, `\`, `:`, `*`, `?`, `<`, `>`, `|`, ` `}
	for _, c := range chars {
		builder.WriteString(fmt.Sprintf("`%s` `_` ", c))
	}
	t := fmt.Sprintf(`{{ .DialogID }}_{{ .MessageID }}_{{ replace .FileName %s }}`, builder.String())
	cmd.Flags().String(consts.FlagDlTemplate, t, "download file name template")

	cmd.Flags().StringSliceVarP(&opts.Include, consts.FlagDlInclude, "i", []string{}, "include the specified file extensions, and only judge by file name, not file MIME. Example: -i mp4,mp3")
	cmd.Flags().StringSliceVarP(&opts.Exclude, consts.FlagDlExclude, "e", []string{}, "exclude the specified file extensions, and only judge by file name, not file MIME. Example: -e png,jpg")

	cmd.Flags().StringVarP(&opts.Dir, consts.FlagDlDir, "d", "downloads", "specify the download directory. If the directory does not exist, it will be created automatically")
	cmd.Flags().BoolVar(&opts.RewriteExt, consts.FlagDlRewriteExt, false, "rewrite file extension according to file header MIME")
	// do not match extension, because some files' extension is corrected by --rewrite-ext flag
	cmd.Flags().BoolVar(&opts.SkipSame, consts.FlagDlSkipSame, false, "skip files with the same name(without extension) and size")

	cmd.Flags().Int64Var(&opts.PoolSize, consts.FlagDlPool, 3, "specify the size of the DC pool")
	cmd.Flags().BoolVar(&opts.Desc, consts.FlagDlDesc, false, "download files from the newest to the oldest ones (may affect resume download)")

	// resume flags, if both false then ask user
	cmd.Flags().BoolVar(&opts.Continue, consts.FlagDlContinue, false, "continue the last download directly")
	cmd.Flags().BoolVar(&opts.Restart, consts.FlagDlRestart, false, "restart the last download directly")

	_ = viper.BindPFlag(consts.FlagDlTemplate, cmd.Flags().Lookup(consts.FlagDlTemplate))

	return cmd
}
