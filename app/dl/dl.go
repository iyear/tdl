package dl

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/spf13/viper"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/downloader"
	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/prog"
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
	Group      bool // auto detect grouped message

	// optimization control
	ForceWebCheck bool // force network-based skip-same check instead of using metadata optimization

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

func Run(ctx context.Context, c *telegram.Client, kvd storage.Storage, opts Options) (rerr error) {
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
	logctx.From(ctx).Debug("Collect dialogs",
		zap.Any("dialogs", dialogs))

	if opts.Serve {
		return serve(ctx, kvd, pool, dialogs, opts.Port, opts.Takeout)
	}

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(pool.Default(ctx))

	dlProgress := prog.New(utils.Byte.FormatBinaryBytes)

	it, err := newIter(pool, manager, dialogs, opts, viper.GetDuration(consts.FlagDelay), dlProgress)
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
			multierr.AppendInto(&rerr, kvd.Delete(ctx, key.Resume(it.Fingerprint())))
		}
	}()

	dlProgress.SetNumTrackersExpected(it.Total())
	prog.EnablePS(ctx, dlProgress)

	options := downloader.Options{
		Pool:     pool,
		Threads:  viper.GetInt(consts.FlagThreads),
		Iter:     it,
		Progress: newProgress(dlProgress, it, opts),
	}
	limit := viper.GetInt(consts.FlagLimit)

	logctx.From(ctx).Info("Start download",
		zap.String("dir", opts.Dir),
		zap.Bool("rewrite_ext", opts.RewriteExt),
		zap.Bool("skip_same", opts.SkipSame),
		zap.Int("threads", options.Threads),
		zap.Int("limit", limit))

	color.Green("All files will be downloaded to '%s' dir", opts.Dir)

	// Warn users if using skip-same with a template that doesn't include MessageID
	if opts.SkipSame && !strings.Contains(opts.Template, "MessageID") {
		color.Yellow("WARNING: Your template does not include MessageID - files may be skipped due to name collisions")
		color.Yellow("         Consider adding {{ .MessageID }} to your template for unique filenames")
	}

	go dlProgress.Render()
	defer prog.Wait(ctx, dlProgress)

	rerr = downloader.New(options).Download(ctx, limit)

	// Log optimization statistics if skip-same was used
	if opts.SkipSame && !opts.ForceWebCheck {
		hits, networkChecks := it.GetOptimizationStats()
		totalChecks := hits + networkChecks
		if totalChecks > 0 {
			savedCalls := hits * 2 // Each optimized skip saves 2 network calls (FromInputPeer + GetSingleMessage)
			percentOptimized := float64(hits) / float64(totalChecks) * 100

			logctx.From(ctx).Info("Skip-same optimization summary",
				zap.Int64("files_skipped_without_network", hits),
				zap.Int64("files_checked_via_network", networkChecks),
				zap.Int64("network_calls_saved", savedCalls),
				zap.Float64("optimization_efficiency_percent", percentOptimized))
		}
	}

	return rerr
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

func resume(ctx context.Context, kvd storage.Storage, iter *iter, ask bool) error {
	logctx.From(ctx).Debug("Check resume key",
		zap.String("fingerprint", iter.Fingerprint()))

	b, err := kvd.Get(ctx, key.Resume(iter.Fingerprint()))
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
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

	logctx.From(ctx).Debug("Resume download",
		zap.Int("finished", len(finished)))

	if !confirm {
		// clear resume key
		return kvd.Delete(ctx, key.Resume(iter.Fingerprint()))
	}

	iter.SetFinished(finished)
	return nil
}

func saveProgress(ctx context.Context, kvd storage.Storage, it *iter) error {
	finished := it.Finished()
	logctx.From(ctx).Debug("Save progress",
		zap.Int("finished", len(finished)))

	b, err := json.Marshal(finished)
	if err != nil {
		return err
	}
	return kvd.Set(ctx, key.Resume(it.Fingerprint()), b)
}
