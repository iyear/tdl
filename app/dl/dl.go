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

func Run(ctx context.Context, dir string, rewriteExt, skipSame bool, template string, urls, files, include, exclude []string, poolSize int64) (rerr error) {
	c, kvd, err := tgc.NoLogin()
	if err != nil {
		return err
	}

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) error {
		color.Green("Preparing DC pool... It may take a while. size: %d", poolSize)

		start := time.Now()
		pool, err := dcpool.NewPool(ctx, c, poolSize, floodwait.NewSimpleWaiter())
		if err != nil {
			return err
		}
		defer multierr.AppendInvoke(&rerr, multierr.Close(pool))

		// clear prepare message
		fmt.Printf("%s%s", text.CursorUp.Sprint(), text.EraseLine.Sprint())
		color.Green("DC pool prepared in %s", time.Since(start))

		umsgs, err := parseURLs(ctx, pool, kvd, urls)
		if err != nil {
			return err
		}

		fmsgs, err := parseFiles(ctx, pool, kvd, files)
		if err != nil {
			return err
		}

		it, err := newIter(pool, kvd, template, include, exclude, umsgs, fmsgs)
		if err != nil {
			return err
		}

		return downloader.New(pool, dir, rewriteExt, skipSame,
			viper.GetInt(consts.FlagPartSize), viper.GetInt(consts.FlagThreads), it).
			Download(ctx, viper.GetInt(consts.FlagLimit))
	})
}
