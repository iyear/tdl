package tclient

import (
	"context"
	"fmt"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"

	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/core/tclient"
	"github.com/iyear/tdl/pkg/key"
)

type Options struct {
	KV               storage.Storage
	Proxy            string
	NTP              string
	ReconnectTimeout time.Duration
	UpdateHandler    telegram.UpdateHandler
}

func GetApp(kv storage.Storage) (App, error) {
	mode, err := kv.Get(context.TODO(), key.App())
	if err != nil {
		mode = []byte(AppBuiltin)
	}
	app, ok := Apps[string(mode)]
	if !ok {
		return App{}, fmt.Errorf("can't find app: %s, please try re-login", mode)
	}

	return app, nil
}

func New(ctx context.Context, o Options, login bool, middlewares ...telegram.Middleware) (*telegram.Client, error) {
	app, err := GetApp(o.KV)
	if err != nil {
		return nil, errors.Wrap(err, "get app")
	}

	return tclient.New(ctx, tclient.Options{
		AppID:            app.AppID,
		AppHash:          app.AppHash,
		Session:          storage.NewSession(o.KV, login),
		Middlewares:      middlewares,
		Proxy:            o.Proxy,
		NTP:              o.NTP,
		ReconnectTimeout: o.ReconnectTimeout,
		UpdateHandler:    o.UpdateHandler,
	})
}
