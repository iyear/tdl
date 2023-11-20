package tgc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/go-faster/errors"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/contrib/middleware/ratelimit"
	tdclock "github.com/gotd/td/clock"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/time/rate"

	"github.com/iyear/tdl/pkg/clock"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/recovery"
	"github.com/iyear/tdl/pkg/retry"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
)

func NewDefaultMiddlewares(ctx context.Context) ([]telegram.Middleware, error) {
	_clock, err := Clock()
	if err != nil {
		return nil, errors.Wrap(err, "create clock")
	}

	return []telegram.Middleware{
		recovery.New(ctx, Backoff(_clock)),
		retry.New(5),
		floodwait.NewSimpleWaiter(),
	}, nil
}

func New(ctx context.Context, login bool, middlewares ...telegram.Middleware) (*telegram.Client, kv.KV, error) {
	var (
		kvd kv.KV
		err error
	)

	if test := viper.GetString(consts.FlagTest); test != "" {
		kvd, err = kv.NewFile(filepath.Join(os.TempDir(), test)) // persistent storage
	} else {
		kvd, err = kv.New(kv.Options{
			Path: consts.KVPath,
			NS:   viper.GetString(consts.FlagNamespace),
		})
	}
	if err != nil {
		return nil, nil, err
	}

	_clock, err := Clock()
	if err != nil {
		return nil, nil, errors.Wrap(err, "create clock")
	}

	mode, err := kvd.Get(key.App())
	if err != nil {
		mode = []byte(consts.AppBuiltin)
	}
	app, ok := consts.Apps[string(mode)]
	if !ok {
		return nil, nil, fmt.Errorf("can't find app: %s, please try re-login", mode)
	}
	appId, appHash := app.AppID, app.AppHash

	opts := telegram.Options{
		Resolver: dcs.Plain(dcs.PlainOptions{
			Dial: utils.Proxy.GetDial(viper.GetString(consts.FlagProxy)).DialContext,
		}),
		ReconnectionBackoff: func() backoff.BackOff {
			return Backoff(_clock)
		},
		Device:         consts.Device,
		SessionStorage: storage.NewSession(kvd, login),
		RetryInterval:  time.Second,
		MaxRetries:     10,
		DialTimeout:    10 * time.Second,
		Middlewares:    middlewares,
		Clock:          _clock,
		Logger:         logger.From(ctx).Named("td"),
	}

	// test mode, hook options
	if viper.GetString(consts.FlagTest) != "" {
		appId, appHash = telegram.TestAppID, telegram.TestAppHash
		opts.DC = 2
		opts.DCList = dcs.Test()
		// add rate limit to avoid frequent flood wait
		opts.Middlewares = append(opts.Middlewares, ratelimit.New(rate.Every(100*time.Millisecond), 5))
	}

	logger.From(ctx).Info("New telegram client",
		zap.Int("app", app.AppID),
		zap.String("mode", string(mode)),
		zap.Bool("is_login", login))

	return telegram.NewClient(appId, appHash, opts), kvd, nil
}

func NoLogin(ctx context.Context, middlewares ...telegram.Middleware) (*telegram.Client, kv.KV, error) {
	mid, err := NewDefaultMiddlewares(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "create default middlewares")
	}

	return New(ctx, false, append(middlewares, mid...)...)
}

func Login(ctx context.Context, middlewares ...telegram.Middleware) (*telegram.Client, kv.KV, error) {
	mid, err := NewDefaultMiddlewares(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "create default middlewares")
	}
	return New(ctx, true, append(middlewares, mid...)...)
}

func Clock() (tdclock.Clock, error) {
	_clock := tdclock.System
	if ntp := viper.GetString(consts.FlagNTP); ntp != "" {
		var err error
		_clock, err = clock.New()
		if err != nil {
			return nil, err
		}
	}

	return _clock, nil
}

func Backoff(_clock tdclock.Clock) backoff.BackOff {
	b := backoff.NewExponentialBackOff()

	b.Multiplier = 1.1
	b.MaxElapsedTime = viper.GetDuration(consts.FlagReconnectTimeout)
	b.Clock = _clock
	return b
}
