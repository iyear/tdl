package up

import (
	"context"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/spf13/viper"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/core/uploader"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/utils"
)

type Options struct {
	Chat     string
	Paths    []string
	Excludes []string
	Remove   bool
	Photo    bool
}

func Run(ctx context.Context, c *telegram.Client, kvd storage.Storage, opts Options) (rerr error) {
	files, err := walk(opts.Paths, opts.Excludes)
	if err != nil {
		return err
	}

	color.Blue("Files count: %d", len(files))

	pool := dcpool.NewPool(c,
		int64(viper.GetInt(consts.FlagPoolSize)),
		tclient.NewDefaultMiddlewares(ctx, viper.GetDuration(consts.FlagReconnectTimeout))...)
	defer multierr.AppendInvoke(&rerr, multierr.Close(pool))

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(pool.Default(ctx))

	to, err := resolveDestPeer(ctx, manager, opts.Chat)
	if err != nil {
		return errors.Wrap(err, "get target peer")
	}

	upProgress := prog.New(utils.Byte.FormatBinaryBytes)
	upProgress.SetNumTrackersExpected(len(files))
	prog.EnablePS(ctx, upProgress)

	options := uploader.Options{
		Client:   pool.Default(ctx),
		Threads:  viper.GetInt(consts.FlagThreads),
		Iter:     newIter(files, to, opts.Photo, opts.Remove, viper.GetDuration(consts.FlagDelay)),
		Progress: newProgress(upProgress),
	}

	up := uploader.New(options)

	go upProgress.Render()
	defer prog.Wait(ctx, upProgress)

	return up.Upload(ctx, viper.GetInt(consts.FlagLimit))
}

func resolveDestPeer(ctx context.Context, manager *peers.Manager, chat string) (peers.Peer, error) {
	if chat == "" {
		return manager.FromInputPeer(ctx, &tg.InputPeerSelf{})
	}

	return tutil.GetInputPeer(ctx, manager, chat)
}
