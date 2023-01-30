package dl

import (
	"errors"
	"github.com/iyear/tdl/app/dl"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var opts = &dl.Options{}

var Cmd = &cobra.Command{
	Use:     "dl",
	Aliases: []string{"download"},
	Short:   "Download anything from Telegram (protected) chat",
	RunE: func(cmd *cobra.Command, args []string) error {
		// only one of include and exclude can be specified
		if len(opts.Include) > 0 && len(opts.Exclude) > 0 {
			return errors.New("only one of `include` and `exclude` can be specified")
		}

		// mkdir if not exists
		if err := os.MkdirAll(opts.Dir, os.ModePerm); err != nil {
			return err
		}

		opts.Template = viper.GetString(consts.FlagDlTemplate)
		return dl.Run(logger.Named(cmd.Context(), "dl"), opts)
	},
}

func init() {
	Cmd.Flags().StringSliceVarP(&opts.URLs, consts.FlagDlUrl, "u", []string{}, "telegram message links")
	Cmd.Flags().StringSliceVarP(&opts.Files, consts.FlagDlFile, "f", []string{}, "official client exported files")
	Cmd.Flags().String(consts.FlagDlTemplate, "{{ .DialogID }}_{{ .MessageID }}_{{ .FileName }}", "download file name template")

	Cmd.Flags().StringSliceVarP(&opts.Include, consts.FlagDlInclude, "i", []string{}, "include the specified file extensions, and only judge by file name, not file MIME. Example: -i mp4,mp3")
	Cmd.Flags().StringSliceVarP(&opts.Exclude, consts.FlagDlExclude, "e", []string{}, "exclude the specified file extensions, and only judge by file name, not file MIME. Example: -e png,jpg")

	Cmd.Flags().StringVarP(&opts.Dir, consts.FlagDlDir, "d", "downloads", "specify the download directory. If the directory does not exist, it will be created automatically")
	Cmd.Flags().BoolVar(&opts.RewriteExt, consts.FlagDlRewriteExt, false, "rewrite file extension according to file header MIME")
	// do not match extension, because some files' extension is corrected by --rewrite-ext flag
	Cmd.Flags().BoolVar(&opts.SkipSame, consts.FlagDlSkipSame, false, "skip files with the same name(without extension) and size")

	Cmd.Flags().Int64Var(&opts.PoolSize, consts.FlagDlPool, 3, "specify the size of the DC pool")

	_ = viper.BindPFlag(consts.FlagDlTemplate, Cmd.Flags().Lookup(consts.FlagDlTemplate))
}
