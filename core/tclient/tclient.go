package tclient

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"github.com/gotd/contrib/clock"
	"github.com/gotd/contrib/middleware/floodwait"
	tdclock "github.com/gotd/td/clock"
	"github.com/gotd/td/exchange"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"golang.org/x/net/proxy"

	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/middlewares/recovery"
	"github.com/iyear/tdl/core/middlewares/retry"
	"github.com/iyear/tdl/core/util/netutil"
	"github.com/iyear/tdl/core/util/tutil"
)

// dc values can be overridden globally
var (
	DCList     dcs.List
	DC         int
	PublicKeys []exchange.PublicKey
)

type Options struct {
	AppID            int
	AppHash          string
	Session          telegram.SessionStorage
	Middlewares      []telegram.Middleware
	Proxy            string
	NTP              string
	ReconnectTimeout time.Duration
	UpdateHandler    telegram.UpdateHandler
}

// New creates new telegram client with given options.
// Default middlewares(retry, recovery, flood wait) always added.
func New(ctx context.Context, o Options) (*telegram.Client, error) {
	// process clock
	tclock := tdclock.System
	if ntp := o.NTP; ntp != "" {
		var err error
		tclock, err = clock.NewNTP(ntp)
		if err != nil {
			return nil, errors.Wrap(err, "create network clock")
		}
	}

	// process proxy
	var dialer dcs.DialFunc = proxy.Direct.DialContext
	if p := o.Proxy; p != "" {
		d, err := netutil.NewProxy(p)
		if err != nil {
			return nil, errors.Wrap(err, "get dialer")
		}
		dialer = d.DialContext
	}

	opts := telegram.Options{
		Resolver: dcs.Plain(dcs.PlainOptions{
			Dial: dialer,
		}),
		ReconnectionBackoff: func() backoff.BackOff {
			return newBackoff(o.ReconnectTimeout)
		},
		DC:             DC,
		DCList:         DCList,
		PublicKeys:     PublicKeys,
		UpdateHandler:  o.UpdateHandler,
		Device:         tutil.Device,
		SessionStorage: o.Session,
		RetryInterval:  5 * time.Second,
		MaxRetries:     5,
		DialTimeout:    10 * time.Second,
		Middlewares:    append(NewDefaultMiddlewares(ctx, o.ReconnectTimeout), o.Middlewares...),
		Clock:          tclock,
		Logger:         logctx.From(ctx).Named("td"),
	}

	return telegram.NewClient(o.AppID, o.AppHash, opts), nil
}

func NewDefaultMiddlewares(ctx context.Context, timeout time.Duration) []telegram.Middleware {
	return []telegram.Middleware{
		recovery.New(ctx, newBackoff(timeout)),
		retry.New(5),
		floodwait.NewSimpleWaiter(),
	}
}

func newBackoff(timeout time.Duration) backoff.BackOff {
	b := backoff.NewExponentialBackOff()

	b.Multiplier = 1.1
	b.MaxElapsedTime = timeout
	b.MaxInterval = 10 * time.Second
	return b
}

func RunWithAuth(ctx context.Context, client *telegram.Client, f func(ctx context.Context) error) error {
	return client.Run(ctx, func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			return fmt.Errorf("not authorized. please login first")
		}

		return f(ctx)
	})
}
