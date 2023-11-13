package forward

import (
	"context"
	"strings"

	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/telegram/peers"
	pw "github.com/jedib0t/go-pretty/v6/progress"
	"github.com/spf13/viper"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/app/internal/tctx"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/forwarder"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/tmessage"
	"github.com/iyear/tdl/pkg/utils"
)

type Options struct {
	From   []string
	To     string
	Mode   forwarder.Mode
	Silent bool
}

func Run(ctx context.Context, opts Options) error {
	c, kvd, err := tgc.NoLogin(ctx)
	if err != nil {
		return err
	}
	ctx = tctx.WithKV(ctx, kvd)

	return tgc.RunWithAuth(ctx, c, func(ctx context.Context) (rerr error) {
		pool := dcpool.NewPool(c, int64(viper.GetInt(consts.FlagPoolSize)), floodwait.NewSimpleWaiter())
		defer multierr.AppendInvoke(&rerr, multierr.Close(pool))

		ctx = tctx.WithPool(ctx, pool)

		dialogs, err := collectDialogs(ctx, opts.From)
		if err != nil {
			return errors.Wrap(err, "collect dialogs")
		}

		manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(pool.Default(ctx))

		peerTo, err := utils.Telegram.GetInputPeer(ctx, manager, opts.To)
		if err != nil {
			return errors.Wrap(err, "resolve dest peer")
		}

		color.Green("All messages will be forwarded to %s(%d)", peerTo.VisibleName(), peerTo.ID())

		fwProgress := prog.New(pw.FormatNumber)

		fw := forwarder.New(forwarder.Options{
			Pool:     pool,
			Iter:     newIter(manager, pool, peerTo, dialogs),
			Silent:   opts.Silent,
			Mode:     opts.Mode,
			Progress: newProgress(fwProgress),
		})

		go fwProgress.Render()
		defer prog.Wait(fwProgress)

		return fw.Forward(ctx)
	})
}

func collectDialogs(ctx context.Context, input []string) ([]*tmessage.Dialog, error) {
	var dialogs []*tmessage.Dialog

	for _, p := range input {
		var (
			d   []*tmessage.Dialog
			err error
		)

		switch {
		case strings.HasPrefix(p, "http"):
			d, err = tmessage.Parse(tmessage.FromURL(ctx, tctx.Pool(ctx), tctx.KV(ctx), []string{p}))
			if err != nil {
				return nil, errors.Wrap(err, "parse from url")
			}
		default:
			d, err = tmessage.Parse(tmessage.FromFile(ctx, tctx.Pool(ctx), tctx.KV(ctx), []string{p}))
			if err != nil {
				return nil, errors.Wrap(err, "parse from file")
			}
		}

		dialogs = append(dialogs, d...)
	}

	return dialogs, nil
}
