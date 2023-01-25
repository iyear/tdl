package dl

import (
	"context"
	"fmt"
	"github.com/fatih/color"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
	"time"
)

type Options struct {
	Dir        string
	RewriteExt bool
	SkipSame   bool
	Template   string
	URLs       []string
	Files      []string
	Include    []string
	Exclude    []string
	PoolSize   int64
}

func Run(ctx context.Context, opts *Options) error {
	c, kvd, err := tgc.NoLogin()
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) (rerr error) {
		color.Green("Preparing DC pool... It may take a while. size: %d", opts.PoolSize)

		start := time.Now()
		pool, err := dcpool.NewPool(ctx, c, opts.PoolSize, floodwait.NewSimpleWaiter())
		if err != nil {
			return err
		}
		defer multierr.AppendInvoke(&rerr, multierr.Close(pool))

		// clear prepare message
		fmt.Printf("%s%s", text.CursorUp.Sprint(), text.EraseLine.Sprint())
		color.Green("DC pool prepared in %s", time.Since(start))

		umsgs, err := parseURLs(ctx, pool, kvd, opts.URLs)
		if err != nil {
			return err
		}

		fmsgs, err := parseFiles(ctx, pool, kvd, opts.Files)
		if err != nil {
			return err
		}

		it, err := newIter(pool, kvd, opts.Template, opts.Include, opts.Exclude, umsgs, fmsgs)
		if err != nil {
			return err
		}

		options := &downloader.Options{
			Pool:       pool,
			Dir:        opts.Dir,
			RewriteExt: opts.RewriteExt,
			SkipSame:   opts.SkipSame,
			PartSize:   viper.GetInt(consts.FlagPartSize),
			Threads:    viper.GetInt(consts.FlagThreads),
			Iter:       it,
		}
		return downloader.New(options).Download(ctx, viper.GetInt(consts.FlagLimit))
	})
}
