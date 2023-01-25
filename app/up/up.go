package up

import (
	"context"
	"github.com/fatih/color"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/uploader"
	"github.com/spf13/viper"
)

func Run(ctx context.Context, chat string, paths, excludes []string) error {
	files, err := walk(paths, excludes)
	if err != nil {
		return err
	}

	color.Blue("Files count: %d", len(files))

	c, kvd, err := tgc.NoLogin()
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) error {
		return uploader.New(c.API(), kvd, viper.GetInt(consts.FlagPartSize), viper.GetInt(consts.FlagThreads), newIter(files)).
			Upload(ctx, chat, viper.GetInt(consts.FlagLimit))
	})
}
