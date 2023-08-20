package dcpool

import (
	"context"
	"sync"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/takeout"
)

type Pool interface {
	Client(ctx context.Context, dc int) *tg.Client
	Takeout(ctx context.Context, dc int) *tg.Client
	Default() int
	Close() error
}

type pool struct {
	api         *telegram.Client
	size        int64
	mu          *sync.Mutex
	middlewares []telegram.Middleware

	invokers map[int]tg.Invoker
	closes   map[int]func() error
	takeout  int64
}

func NewPool(c *telegram.Client, size int64, middlewares ...telegram.Middleware) Pool {
	return &pool{
		api:         c,
		size:        size,
		mu:          &sync.Mutex{},
		middlewares: middlewares,
		invokers:    make(map[int]tg.Invoker),
		closes:      make(map[int]func() error),
		takeout:     0,
	}
}

func (p *pool) current() int {
	return p.api.Config().ThisDC
}

func (p *pool) Client(ctx context.Context, dc int) *tg.Client {
	return tg.NewClient(p.invoker(ctx, dc))
}

func (p *pool) invoker(ctx context.Context, dc int) tg.Invoker {
	p.mu.Lock()
	defer p.mu.Unlock()

	if i, ok := p.invokers[dc]; ok {
		return i
	}

	// lazy init
	var (
		invoker telegram.CloseInvoker
		err     error
	)
	if dc == p.current() { // can't transfer dc to current dc
		invoker, err = p.api.Pool(p.size)
	} else {
		invoker, err = p.api.DC(ctx, dc, p.size)
	}

	if err != nil {
		logger.From(ctx).Error("create invoker", zap.Error(err))
		return p.api // degraded
	}

	p.closes[dc] = invoker.Close
	p.invokers[dc] = chainMiddlewares(invoker, p.middlewares...)

	return p.invokers[dc]
}

func (p *pool) Default() int {
	return p.api.Config().ThisDC
}

func (p *pool) Close() (err error) {
	if p.takeout != 0 {
		err = takeout.UnTakeout(context.TODO(), p.Takeout(context.TODO(), p.current()).Invoker())
	}

	for _, c := range p.closes {
		err = multierr.Append(err, c())
	}

	return err
}

func (p *pool) Takeout(ctx context.Context, dc int) *tg.Client {
	p.mu.Lock()
	defer p.mu.Unlock()

	// lazy init
	if p.takeout == 0 {
		sid, err := takeout.Takeout(ctx, p.api)
		if err != nil {
			logger.From(ctx).Warn("takeout error", zap.Error(err))
			// ignore init delay error and return non-takeout client
			return p.Client(ctx, dc)
		}
		p.takeout = sid
		logger.From(ctx).Info("get takeout id", zap.Int64("id", sid))
	}

	return tg.NewClient(chainMiddlewares(p.invoker(ctx, dc), takeout.Middleware(p.takeout)))
}
