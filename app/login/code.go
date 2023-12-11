package login

import (
	"context"
	"crypto/rand"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/spf13/viper"

	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/tclient"
)

func Code(ctx context.Context) error {
	kvd, err := kv.From(ctx).Open(viper.GetString(consts.FlagNamespace))
	if err != nil {
		return errors.Wrap(err, "open kv")
	}
	c, err := tclient.New(ctx, tclient.Options{
		KV:               kvd,
		Proxy:            viper.GetString(consts.FlagProxy),
		NTP:              viper.GetString(consts.FlagNTP),
		ReconnectTimeout: viper.GetDuration(consts.FlagReconnectTimeout),
		Test:             viper.GetString(consts.FlagTest) != "",
		UpdateHandler:    nil,
	}, true)
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

		if err = kvd.Set(key.App(), []byte(tclient.AppBuiltin)); err != nil {
			return err
		}

		color.Blue("Login successfully! ID: %d, Username: %s", user.ID, user.Username)

		return nil
	})
}
