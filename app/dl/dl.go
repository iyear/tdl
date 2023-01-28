package dl

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/kv"
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

		// resume download and ask user to continue
		if err = resume(ctx, kvd, it); err != nil {
			return err
		}
		defer func() {
			if rerr != nil { // download is interrupted
				multierr.AppendInto(&rerr, saveProgress(kvd, it))
			} else { // if finished, we should clear resume key
				multierr.AppendInto(&rerr, kvd.Delete(key.Resume(it.fingerprint)))
			}
		}()

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

func resume(ctx context.Context, kvd kv.KV, it *iter) error {
	b, err := kvd.Get(key.Resume(it.fingerprint))
	if err != nil && !errors.Is(err, kv.ErrNotFound) {
		return err
	}
	if len(b) == 0 { // no progress
		return nil
	}

	finished := make(map[int]struct{})
	if err = json.Unmarshal(b, &finished); err != nil {
		return err
	}

	// finished is empty, no need to resume
	if len(finished) == 0 {
		return nil
	}

	confirm := false
	if err = survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Found unfinished download, continue from '%d/%d'?", len(finished), it.Total(ctx)),
	}, &confirm); err != nil {
		return err
	}

	if !confirm {
		// clear resume key
		return kvd.Delete(key.Resume(it.fingerprint))
	}

	it.setFinished(finished)
	return nil
}

func saveProgress(kvd kv.KV, it *iter) error {
	b, err := json.Marshal(it.finished)
	if err != nil {
		return err
	}
	return kvd.Set(key.Resume(it.fingerprint), b)
}
