package up

import (
	"context"

	"github.com/fatih/color"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/tg"
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
		pool := dcpool.NewPool(c, int64(viper.GetInt(consts.FlagPoolSize)), floodwait.NewSimpleWaiter())
		defer multierr.AppendInvoke(&rerr, multierr.Close(pool))

		options := uploader.Options{
			Client:   tg.NewClient(pool.Client(ctx, pool.Default()).Invoker()),
			KV:       kvd,
			PartSize: viper.GetInt(consts.FlagPartSize),
			Threads:  viper.GetInt(consts.FlagThreads),
			Iter:     newIter(files, opts.Remove),
			Photo:    opts.Photo,
		}
		return uploader.New(options).Upload(ctx, opts.Chat, viper.GetInt(consts.FlagLimit))
	})
}
