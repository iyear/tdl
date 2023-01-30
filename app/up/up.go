package up

import (
	"context"
	"github.com/fatih/color"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/uploader"
	"github.com/spf13/viper"
)

type Options struct {
	Chat     string
	Paths    []string
	Excludes []string
}

func Run(ctx context.Context, opts *Options) error {
	files, err := walk(opts.Paths, opts.Excludes)
	if err != nil {
		return err
	}

	color.Blue("Files count: %d", len(files))

	c, kvd, err := tgc.NoLogin(ctx)
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) error {
		options := &uploader.Options{
			Client:   c.API(),
			KV:       kvd,
			PartSize: viper.GetInt(consts.FlagPartSize),
			Threads:  viper.GetInt(consts.FlagThreads),
			Iter:     newIter(files),
		}
		return uploader.New(options).Upload(ctx, opts.Chat, viper.GetInt(consts.FlagLimit))
	})
}
