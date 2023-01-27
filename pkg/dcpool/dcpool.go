package dcpool

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/takeout"
	"go.uber.org/multierr"
	"golang.org/x/sync/errgroup"
	"sync"
)

type Pool interface {
	Client(dc int) *tg.Client
	Takeout(dc int) *tg.Client
	Default() int
	Close() error
}

type pool struct {
	invokers map[int]tg.Invoker
	closes   map[int]func() error
	_default int
	takeout  int64
}

func NewPool(ctx context.Context, c *telegram.Client, size int64, middlewares ...telegram.Middleware) (Pool, error) {
	m := make(map[int]tg.Invoker)
	closes := make(map[int]func() error)
	mu := &sync.Mutex{}
	curDC := c.Config().ThisDC

	dcs := collectDCs(c.Config().DCOptions)

	wg, errctx := errgroup.WithContext(ctx)

	for _, dc := range dcs {
		dc := dc
		wg.Go(func() error {
			var (
				invoker telegram.CloseInvoker
				err     error
			)

			if dc == curDC { // can't transfer dc to current dc
				invoker, err = c.Pool(size)
			} else {
				invoker, err = c.DC(errctx, dc, size)
			}

			if err != nil {
				return err
			}

			mu.Lock()
			closes[dc] = invoker.Close
			m[dc] = chainMiddlewares(invoker, middlewares...)
			mu.Unlock()

			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	if _, ok := m[curDC]; !ok {
		return nil, fmt.Errorf("default DC %d not in dcs", curDC)
	}

	sid, err := takeout.Takeout(ctx, m[curDC])
	if err != nil {
		return nil, err
	}

	return &pool{
		invokers: m,
		closes:   closes,
		_default: curDC,
		takeout:  sid,
	}, nil
}

func collectDCs(dcOpts []tg.DCOption) (dcs []int) {
	m := make(map[int]struct{})
	for _, opt := range dcOpts {
		m[opt.ID] = struct{}{}
	}

	for dc := range m {
		dcs = append(dcs, dc)
	}
	return dcs
}

func (p *pool) Client(dc int) *tg.Client {
	return tg.NewClient(p.invoker(dc))
}

func (p *pool) invoker(dc int) tg.Invoker {
	i, ok := p.invokers[dc]
	if !ok {
		return p.invokers[p._default]
	}
	return i
}

func (p *pool) Default() int {
	return p._default
}

func (p *pool) Close() error {
	var err error
	for _, c := range p.closes {
		err = multierr.Append(err, c())
	}

	err = multierr.Append(err, takeout.UnTakeout(context.TODO(), p.invoker(p._default)))

	return err
}
