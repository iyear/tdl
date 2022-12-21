package dl

import (
	"errors"
	"github.com/iyear/tdl/app/dl"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	urls, files          []string
	include, exclude     []string
	dir                  string
	rewriteExt, skipSame bool
	poolSize             int64
)

var Cmd = &cobra.Command{
	Use:     "dl",
	Aliases: []string{"download"},
	Short:   "Download anything from Telegram (protected) chat",
	RunE: func(cmd *cobra.Command, args []string) error {
		// only one of include and exclude can be specified
		if len(include) > 0 && len(exclude) > 0 {
			return errors.New("only one of `include` and `exclude` can be specified")
		}

		// mkdir if not exists
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}

		return dl.Run(cmd.Context(), dir, rewriteExt, skipSame,
			viper.GetString(consts.FlagDlTemplate),
			urls, files, include, exclude, poolSize)
	},
}

func init() {
	Cmd.Flags().StringSliceVarP(&urls, consts.FlagDlUrl, "u", []string{}, "telegram message links")
	Cmd.Flags().StringSliceVarP(&files, consts.FlagDlFile, "f", []string{}, "official client exported files")
	Cmd.Flags().String(consts.FlagDlTemplate, "{{ .DialogID }}_{{ .MessageID }}_{{ .FileName }}", "download file name template")

	Cmd.Flags().StringSliceVarP(&include, consts.FlagDlInclude, "i", []string{}, "include the specified file extensions, and only judge by file name, not file MIME. Example: -i mp4,mp3")
	Cmd.Flags().StringSliceVarP(&exclude, consts.FlagDlExclude, "e", []string{}, "exclude the specified file extensions, and only judge by file name, not file MIME. Example: -e png,jpg")

	Cmd.Flags().StringVarP(&dir, consts.FlagDlDir, "d", "downloads", "specify the download directory. If the directory does not exist, it will be created automatically")
	Cmd.Flags().BoolVar(&rewriteExt, consts.FlagDlRewriteExt, false, "rewrite file extension according to file header MIME")
	// do not match extension, because some files' extension is corrected by --rewrite-ext flag
	Cmd.Flags().BoolVar(&skipSame, consts.FlagDlSkipSame, false, "skip files with the same name(without extension) and size")

	Cmd.Flags().Int64Var(&poolSize, consts.FlagDlPool, 3, "specify the size of the DC pool")

	_ = viper.BindPFlag(consts.FlagDlTemplate, Cmd.Flags().Lookup(consts.FlagDlTemplate))
}
