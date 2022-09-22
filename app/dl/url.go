package dl

import (
	"context"
	"fmt"
	"github.com/gotd/contrib/middleware/floodwait"
	"github.com/gotd/td/telegram/peers"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/downloader"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/spf13/viper"
)

func Run(ctx context.Context, urls []string) error {
	kvd, err := kv.New(kv.Options{
		Path: consts.KVPath,
		NS:   viper.GetString(consts.FlagNamespace),
	})
	if err != nil {
		return err
	}

	c, err := tgc.New(viper.GetString(consts.FlagProxy), kvd, false, floodwait.NewSimpleWaiter())
	if err != nil {
		return err
	}

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

		return downloader.New(c.API(), viper.GetInt(consts.FlagPartSize), viper.GetInt(consts.FlagThreads), newIter(c.API(), msgs)).
			Download(ctx, viper.GetInt(consts.FlagLimit))
	})
}
