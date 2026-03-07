package forward

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	pw "github.com/jedib0t/go-pretty/v6/progress"
	"github.com/spf13/viper"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/app/internal/tctx"
	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/forwarder"
	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/texpr"
	"github.com/iyear/tdl/pkg/tmessage"
)

type Options struct {
	From   []string
	To     string
	Edit   string
	Mode   forwarder.Mode
	Silent bool
	DryRun bool
	Single bool
	Desc   bool
}

func Run(ctx context.Context, c *telegram.Client, kvd storage.Storage, opts Options) (rerr error) {
	if opts.To == "-" || opts.Edit == "-" {
		fg := texpr.NewFieldsGetter(nil)

		fields, err := fg.Walk(exprEnv(nil, nil))
		if err != nil {
			return fmt.Errorf("failed to walk fields: %w", err)
		}

		fmt.Print(fg.Sprint(fields, true))
		return nil
	}

	ctx = tctx.WithKV(ctx, kvd)

	pool := dcpool.NewPool(c,
		int64(viper.GetInt(consts.FlagPoolSize)),
		tclient.NewDefaultMiddlewares(ctx, viper.GetDuration(consts.FlagReconnectTimeout))...)
	defer multierr.AppendInvoke(&rerr, multierr.Close(pool))

	ctx = tctx.WithPool(ctx, pool)

	dialogs, err := collectDialogs(ctx, opts.From, opts.Desc)
	if err != nil {
		return errors.Wrap(err, "collect dialogs")
	}

	manager := peers.Options{Storage: storage.NewPeers(kvd)}.Build(pool.Default(ctx))

	to, err := resolveDest(ctx, manager, opts.To)
	if err != nil {
		return errors.Wrap(err, "resolve dest peer")
	}

	edit, err := resolveEdit(opts.Edit)
	if err != nil {
		return errors.Wrap(err, "resolve edit")
	}

	fwProgress := prog.New(pw.FormatNumber)
	fwProgress.SetNumTrackersExpected(totalMessages(dialogs))
	stopPS := func() {}
	if viper.GetBool(consts.FlagProgressPS) {
		stopPS = prog.EnablePS(ctx, fwProgress)
	} else {
		fwProgress.Style().Visibility.Pinned = false
	}

	fw := forwarder.New(forwarder.Options{
		Pool: pool,
		Iter: newIter(iterOptions{
			manager: manager,
			pool:    pool,
			to:      to,
			edit:    edit,
			dialogs: dialogs,
			mode:    opts.Mode,
			silent:  opts.Silent,
			dryRun:  opts.DryRun,
			grouped: !opts.Single,
			delay:   viper.GetDuration(consts.FlagDelay),
		}),
		Progress: newProgress(fwProgress),
		Threads:  viper.GetInt(consts.FlagThreads),
	})

	go fwProgress.Render()
	defer func() {
		stopPS()
		prog.Wait(ctx, fwProgress)
	}()

	return fw.Forward(ctx)
}

func collectDialogs(ctx context.Context, input []string, desc bool) ([]*tmessage.Dialog, error) {
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
			d, err = tmessage.Parse(tmessage.FromFile(ctx, tctx.Pool(ctx), tctx.KV(ctx), []string{p}, false))
			if err != nil {
				return nil, errors.Wrap(err, "parse from file")
			}
		}

		if desc {
			for _, dd := range d {
				for i, j := 0, len(dd.Messages)-1; i < j; i, j = i+1, j-1 {
					dd.Messages[i], dd.Messages[j] = dd.Messages[j], dd.Messages[i]
				}
			}
		}

		dialogs = append(dialogs, d...)
	}

	return dialogs, nil
}

// resolveDest parses the input string and returns a vm.Program. It can be a CHAT, a text or a file based on expression engine.
func resolveDest(ctx context.Context, manager *peers.Manager, input string) (*vm.Program, error) {
	compile := func(i string) (*vm.Program, error) {
		// we pass empty peer and message to enable type checking
		return expr.Compile(i, expr.Env(exprEnv(nil, nil)))
	}

	// default
	if input == "" {
		return compile(`""`)
	}

	// file
	if exp, err := os.ReadFile(input); err == nil {
		return compile(string(exp))
	}

	// chat
	if _, err := tutil.GetInputPeer(ctx, manager, input); err == nil {
		// convert to const string
		return compile(fmt.Sprintf(`"%s"`, input))
	}

	// text
	return compile(input)
}

// resolveEdit returns nil if input is empty, otherwise it returns a vm.Program. It can be a text or a file based on expression engine.
func resolveEdit(input string) (*vm.Program, error) {
	compile := func(i string) (*vm.Program, error) {
		// we pass empty peer and message to enable type checking
		return expr.Compile(i, expr.Env(exprEnv(nil, nil)), expr.AsKind(reflect.String))
	}

	// no edit, nil program
	if input == "" {
		return nil, nil
	}

	// file
	if exp, err := os.ReadFile(input); err == nil {
		return compile(string(exp))
	}

	// text
	return compile(input)
}

func totalMessages(dialogs []*tmessage.Dialog) int {
	var total int
	for _, d := range dialogs {
		total += len(d.Messages)
	}
	return total
}
