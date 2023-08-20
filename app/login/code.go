package login

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/fatih/color"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/spf13/viper"

	"github.com/iyear/tdl/app/internal/tgc"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/key"
)

func Code(ctx context.Context) error {
	c, kv, err := tgc.Login(ctx)
	if err != nil {
		return err
	}

	return c.Run(ctx, func(ctx context.Context) error {
		if err = c.Ping(ctx); err != nil {
			return err
		}

		if viper.GetString(consts.FlagTest) != "" {
			authClient := auth.NewClient(c.API(), rand.Reader, telegram.TestAppID, telegram.TestAppHash)

			return backoff.Retry(func() error {
				if err = auth.NewFlow(
					auth.Test(rand.Reader, 2),
					auth.SendCodeOptions{},
				).Run(ctx, authClient); err != nil {
					return err
				}
				return nil
			}, backoff.NewConstantBackOff(time.Second))
		}

		color.Yellow("WARN: Using the built-in APP_ID & APP_HASH may increase the probability of blocking")
		color.Blue("Login...")

		flow := auth.NewFlow(termAuth{}, auth.SendCodeOptions{})
		if err = c.Auth().IfNecessary(ctx, flow); err != nil {
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
