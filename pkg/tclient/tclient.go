package tclient

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	tdclock "github.com/gotd/td/clock"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"golang.org/x/net/proxy"
	"golang.org/x/time/rate"

	"github.com/iyear/tdl/pkg/clock"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/recovery"
	"github.com/iyear/tdl/pkg/retry"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
)

type Options struct {
	KV               kv.KV
	Proxy            string
	NTP              string
	ReconnectTimeout time.Duration
	Test             bool
	UpdateHandler    telegram.UpdateHandler
}

func New(ctx context.Context, o Options, login bool, middlewares ...telegram.Middleware) (*telegram.Client, error) {
	_clock := tdclock.System
	if ntp := o.NTP; ntp != "" {
		var err error
		_clock, err = clock.New()
		if err != nil {
			return nil, errors.Wrap(err, "create network clock")
		}
	}

	mode, err := o.KV.Get(key.App())
	if err != nil {
		mode = []byte(AppBuiltin)
	}
	app, ok := Apps[string(mode)]
	if !ok {
		return nil, fmt.Errorf("can't find app: %s, please try re-login", mode)
	}
	appId, appHash := app.AppID, app.AppHash

	// process proxy
	var dialer dcs.DialFunc = proxy.Direct.DialContext
	if p := o.Proxy; p != "" {
		d, err := utils.Proxy.GetDial(p)
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
		UpdateHandler:  o.UpdateHandler,
		Device:         Device,
		SessionStorage: storage.NewSession(o.KV, login),
		RetryInterval:  5 * time.Second,
		MaxRetries:     -1, // infinite retries
		DialTimeout:    10 * time.Second,
		Middlewares:    append(NewDefaultMiddlewares(ctx, o.ReconnectTimeout), middlewares...),
		Clock:          _clock,
		Logger:         logger.From(ctx).Named("td"),
	}

	// test mode, hook options
	if o.Test {
		appId, appHash = telegram.TestAppID, telegram.TestAppHash
		opts.DC = 2
		opts.DCList = dcs.Test()
		// add rate limit to avoid frequent flood wait
		opts.Middlewares = append(opts.Middlewares, ratelimit.New(rate.Every(100*time.Millisecond), 5))
	}

	return telegram.NewClient(appId, appHash, opts), nil
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

func Run(ctx context.Context, client *telegram.Client, f func(ctx context.Context) error) error {
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
