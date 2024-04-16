package dl

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/tclient"
	"github.com/iyear/tdl/pkg/tmessage"
	"github.com/iyear/tdl/pkg/utils"
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
	Desc       bool
	Takeout    bool

	// resume opts
	Continue, Restart bool

	// serve
	Serve bool
	Port  int
}

type parser struct {
	Data   []string
	Parser tmessage.ParseSource
}

func Run(ctx context.Context, c *telegram.Client, kvd kv.KV, opts Options) (rerr error) {
	pool := dcpool.NewPool(c,
		int64(viper.GetInt(consts.FlagPoolSize)),
		tclient.NewDefaultMiddlewares(ctx, viper.GetDuration(consts.FlagReconnectTimeout))...)
	defer multierr.AppendInvoke(&rerr, multierr.Close(pool))

	parsers := []parser{
		{Data: opts.URLs, Parser: tmessage.FromURL(ctx, pool, kvd, opts.URLs)},
		{Data: opts.Files, Parser: tmessage.FromFile(ctx, pool, kvd, opts.Files, true)},
	}
	dialogs, err := collectDialogs(parsers)
	if err != nil {
		return err
	}
	logger.From(ctx).Debug("Collect dialogs",
		zap.Any("dialogs", dialogs))

	if opts.Serve {
		return serve(ctx, kvd, pool, dialogs, opts.Port, opts.Takeout)
	}

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(pool.Default(ctx))

	it, err := newIter(pool, manager, dialogs, opts)
	if err != nil {
		return err
	}

	if !opts.Restart {
		// resume download and ask user to continue
		if err = resume(ctx, kvd, it, !opts.Continue); err != nil {
			return err
		}
	} else {
		color.Yellow("Restart download by 'restart' flag")
	}

	defer func() { // save progress
		if rerr != nil { // download is interrupted
			multierr.AppendInto(&rerr, saveProgress(ctx, kvd, it))
		} else { // if finished, we should clear resume key
			multierr.AppendInto(&rerr, kvd.Delete(key.Resume(it.Fingerprint())))
		}
	}()

	dlProgress := prog.New(utils.Byte.FormatBinaryBytes)
	dlProgress.SetNumTrackersExpected(it.Total())
	prog.EnablePS(ctx, dlProgress)

	options := downloader.Options{
		Pool:     pool,
		PartSize: viper.GetInt(consts.FlagPartSize),
		Threads:  viper.GetInt(consts.FlagThreads),
		Iter:     it,
		Progress: newProgress(dlProgress, it, opts),
	}
	limit := viper.GetInt(consts.FlagLimit)

	logger.From(ctx).Info("Start download",
		zap.String("dir", opts.Dir),
		zap.Bool("rewrite_ext", opts.RewriteExt),
		zap.Bool("skip_same", opts.SkipSame),
		zap.Int("part_size", options.PartSize),
		zap.Int("threads", options.Threads),
		zap.Int("limit", limit))

	color.Green("All files will be downloaded to '%s' dir", opts.Dir)

	go dlProgress.Render()
	defer prog.Wait(ctx, dlProgress)

	return downloader.New(options).Download(ctx, limit)
}

func collectDialogs(parsers []parser) ([][]*tmessage.Dialog, error) {
	var dialogs [][]*tmessage.Dialog
	for _, p := range parsers {
		d, err := tmessage.Parse(p.Parser)
		if err != nil {
			return nil, err
		}
		dialogs = append(dialogs, d)
	}
	return dialogs, nil
}

func resume(ctx context.Context, kvd kv.KV, iter *iter, ask bool) error {
	logger.From(ctx).Debug("Check resume key",
		zap.String("fingerprint", iter.Fingerprint()))

	b, err := kvd.Get(key.Resume(iter.Fingerprint()))
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
	resumeStr := fmt.Sprintf("Found unfinished download, continue from '%d/%d'", len(finished), iter.Total())
	if ask {
		if err = survey.AskOne(&survey.Confirm{
			Message: color.YellowString(resumeStr + "?"),
		}, &confirm); err != nil {
			return err
		}
	} else {
		color.Yellow(resumeStr)
		confirm = true
	}

	logger.From(ctx).Debug("Resume download",
		zap.Int("finished", len(finished)))

	if !confirm {
		// clear resume key
		return kvd.Delete(key.Resume(iter.Fingerprint()))
	}

	iter.SetFinished(finished)
	return nil
}

func saveProgress(ctx context.Context, kvd kv.KV, it *iter) error {
	finished := it.Finished()
	logger.From(ctx).Debug("Save progress",
		zap.Int("finished", len(finished)))

	b, err := json.Marshal(finished)
	if err != nil {
		return err
	}
	return kvd.Set(key.Resume(it.Fingerprint()), b)
}
