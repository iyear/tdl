package tgc

import (
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
	"time"
)

func New(proxy string, kvd *kv.KV, login bool, middlewares ...telegram.Middleware) *telegram.Client {
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
	})
}
