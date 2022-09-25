package login

import (
	"context"
	"github.com/fatih/color"
	"github.com/gotd/td/telegram/auth"
	"github.com/iyear/tdl/app/internal/tgc"
)

func Code(ctx context.Context) error {
	c, _, err := tgc.Login()
	if err != nil {
		return err
	}

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
