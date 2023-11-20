package up

import (
	"context"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/spf13/viper"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/uploader"
)

type Options struct {
	Chat     string
	Paths    []string
	Excludes []string
	Remove   bool
	Photo    bool
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

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) (rerr error) {
		middlewares, err := tgc.NewDefaultMiddlewares(ctx)
		if err != nil {
			return errors.Wrap(err, "create middlewares")
		}

		pool := dcpool.NewPool(c, int64(viper.GetInt(consts.FlagPoolSize)), middlewares...)
		defer multierr.AppendInvoke(&rerr, multierr.Close(pool))

		options := uploader.Options{
			Client:   pool.Default(ctx),
			KV:       kvd,
			PartSize: viper.GetInt(consts.FlagPartSize),
			Threads:  viper.GetInt(consts.FlagThreads),
			Iter:     newIter(files, opts.Remove),
			Photo:    opts.Photo,
		}

		up, err := uploader.New(options)
		if err != nil {
			return errors.Wrap(err, "create uploader")
		}
		return up.Upload(ctx, opts.Chat, viper.GetInt(consts.FlagLimit))
	})
}
