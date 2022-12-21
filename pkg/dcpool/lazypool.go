package dcpool

import (
	"context"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"go.uber.org/multierr"
	"sync"
)

type lazyPool struct {
	mu       *sync.Mutex
	client   *telegram.Client
	invokers map[int]telegram.CloseInvoker
	_default int
	size     int64
	ctx      context.Context
}

type gotdCloseInvoker struct {
	client *telegram.Client
}

func (g *gotdCloseInvoker) Invoke(ctx context.Context, input bin.Encoder, output bin.Decoder) error {
	return g.client.Invoke(ctx, input, output)
}

func (g *gotdCloseInvoker) Close() error {
	return nil
}

func NewLazyPool(ctx context.Context, c *telegram.Client, size int64) (Pool, error) {
	return &lazyPool{
		mu:       &sync.Mutex{},
		client:   c,
		invokers: make(map[int]telegram.CloseInvoker),
		_default: c.Config().ThisDC,
		size:     size,
		ctx:      ctx,
	}, nil
}

func (p *lazyPool) Client(dc int) *tg.Client {
	return tg.NewClient(p.Invoker(dc))
}

func (p *lazyPool) Invoker(dc int) telegram.CloseInvoker {
	i, ok := p.invokers[dc]
	if !ok {
		var (
			invoker telegram.CloseInvoker
			err     error
		)
		if dc == p._default {
			invoker, err = p.client.Pool(p.size)
		} else {
			invoker, err = p.client.DC(p.ctx, dc, p.size)
		}

		if err != nil {
			return &gotdCloseInvoker{client: p.client}
		}

		p.mu.Lock()
		p.invokers[dc] = invoker
		p.mu.Unlock()
		return invoker
	}

	return i
}

func (p *lazyPool) Default() int {
	return p._default
}

func (p *lazyPool) Close() error {
	var err error
	for _, invokers := range p.invokers {
		err = multierr.Append(err, invokers.Close())
	}

	return err
}
