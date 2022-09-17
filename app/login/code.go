package login

import (
	"context"
	"github.com/fatih/color"
	"github.com/gotd/td/telegram/auth"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/kv"
)

func Code(ctx context.Context, ns, proxy string) error {
	kvd, err := kv.New(kv.Options{
		Path: consts.KVPath,
		NS:   ns,
	})
	if err != nil {
		return err
	}

	c := tgc.New(proxy, kvd, true)

	return c.Run(ctx, func(ctx context.Context) error {
		if err := c.Ping(ctx); err != nil {
			return err
		}

		color.Blue("Login...")

		flow := auth.NewFlow(termAuth{}, auth.SendCodeOptions{})
		if err := c.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		user, err := c.Self(ctx)
		if err != nil {
			return err
		}

		color.Blue("Login successfully! ID: %d, Username: %s", user.ID, user.Username)

		return nil
	})
}
