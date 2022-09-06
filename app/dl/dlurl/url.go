package dlurl

import (
	"context"
	"fmt"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/peers"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
)

func Run(ctx context.Context, ns, proxy string, partSize, threads, limit int, urls []string) error {
	kvd, err := kv.New(kv.Options{
		Path: consts.KVPath,
		NS:   ns,
	})
	if err != nil {
		return err
	}

	c := telegram.NewClient(consts.AppID, consts.AppHash, telegram.Options{
		Resolver: dcs.Plain(dcs.PlainOptions{
			Dial: utils.Proxy.GetDial(proxy).DialContext,
		}),
		Device:         consts.Device,
		SessionStorage: storage.NewSession(kvd, false),
		// RetryInterval:  time.Second,
		MaxRetries: 10,
		Middlewares: []telegram.Middleware{
			floodwait.NewSimpleWaiter(),
			// ratelimit.New(rate.Every(300*time.Millisecond), 3),
		},
	})

	return c.Run(ctx, func(ctx context.Context) error {
		status, err := c.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			return fmt.Errorf("not authorized. please login first")
		}

		manager := peers.Options{}.Build(c.API())

		msgs := make([]*msg, 0, len(urls))

		for _, u := range urls {
			ch, msgid, err := utils.Telegram.ParseChannelMsgLink(ctx, manager, u)
			if err != nil {
				return err
			}

			msgs = append(msgs, &msg{ch: ch, msg: msgid})
		}

		return downloader.New(c.API(), partSize, threads, newIter(c.API(), msgs)).Download(ctx, limit)
	})
}
