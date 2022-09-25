package tgc

import (
	"github.com/gotd/contrib/middleware/floodwait"
	tdclock "github.com/gotd/td/clock"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/iyear/tdl/pkg/clock"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/spf13/viper"
	"time"
)

func New(login bool, middlewares ...telegram.Middleware) (*telegram.Client, *kv.KV, error) {
	var (
		_clock tdclock.Clock
		err    error
	)

	kvd, err := kv.New(kv.Options{
		Path: consts.KVPath,
		NS:   viper.GetString(consts.FlagNamespace),
	})
	if err != nil {
		return nil, nil, err
	}

	_clock = tdclock.System
	if ntp := viper.GetString(consts.FlagNTP); ntp != "" {
		_clock, err = clock.New()
		if err != nil {
			return nil, nil, err
		}
	}

	return telegram.NewClient(consts.AppID, consts.AppHash, telegram.Options{
		Resolver: dcs.Plain(dcs.PlainOptions{
			Dial: utils.Proxy.GetDial(viper.GetString(consts.FlagProxy)).DialContext,
		}),
		Device:         consts.Device,
		SessionStorage: storage.NewSession(kvd, login),
		RetryInterval:  time.Second,
		MaxRetries:     10,
		DialTimeout:    10 * time.Second,
		Middlewares:    middlewares,
		Clock:          _clock,
		Logger:         logger.Logger.Named("client"),
	}), kvd, nil
}

func NoLogin(middlewares ...telegram.Middleware) (*telegram.Client, *kv.KV, error) {
	return New(false, append(middlewares, floodwait.NewSimpleWaiter())...)
}

func Login(middlewares ...telegram.Middleware) (*telegram.Client, *kv.KV, error) {
	return New(true, append(middlewares, floodwait.NewSimpleWaiter())...)
}
