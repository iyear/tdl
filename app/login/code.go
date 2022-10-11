package login

import (
	"context"
	"github.com/fatih/color"
	"github.com/gotd/td/telegram/auth"
	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/key"
)

func Code(ctx context.Context) error {
	c, kv, err := tgc.Login()
	if err != nil {
		return err
	}

	return c.Run(ctx, func(ctx context.Context) error {
		if err := c.Ping(ctx); err != nil {
			return err
		}

		color.Yellow("WARN: Using the built-in APP_ID & APP_HASH may increase the probability of blocking")
		color.Blue("Login...")

		flow := auth.NewFlow(termAuth{}, auth.SendCodeOptions{})
		if err := c.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		user, err := c.Self(ctx)
		if err != nil {
			return err
		}

		if err = kv.Set(key.App(), []byte(consts.AppBuiltin)); err != nil {
			return err
		}

		color.Blue("Login successfully! ID: %d, Username: %s", user.ID, user.Username)

		return nil
	})
}
