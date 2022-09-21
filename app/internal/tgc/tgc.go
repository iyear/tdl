package tgc

import (
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

func New(proxy string, kvd *kv.KV, login bool, middlewares ...telegram.Middleware) (*telegram.Client, error) {
	var (
		_clock tdclock.Clock
		err    error
	)

	_clock = tdclock.System
	if ntp := viper.GetString(consts.FlagNTP); ntp != "" {
		_clock, err = clock.New()
		if err != nil {
			return nil, err
		}
	}

	return telegram.NewClient(consts.AppID, consts.AppHash, telegram.Options{
		Resolver: dcs.Plain(dcs.PlainOptions{
			Dial: utils.Proxy.GetDial(proxy).DialContext,
		}),
		Device:         consts.Device,
		SessionStorage: storage.NewSession(kvd, login),
		RetryInterval:  time.Second,
		MaxRetries:     10,
		DialTimeout:    10 * time.Second,
		Middlewares:    middlewares,
		Clock:          _clock,
		Logger:         logger.Logger.Named("client"),
	}), nil
}
