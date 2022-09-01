package login

import (
	"context"
	"github.com/fatih/color"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/dcs"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
	"github.com/tcnksm/go-input"
	"time"
)

func Run(ctx context.Context, ns, proxy string) error {
	kvd, err := kv.New(kv.Options{
		Path: consts.KVPath,
		NS:   ns,
	})
	if err != nil {
		return err
	}

	c := telegram.NewClient(consts.AppID, consts.AppHash, telegram.Options{
		Resolver: dcs.Plain(dcs.PlainOptions{
			Dial: utils.GetDial(proxy).DialContext,
		}),
		Device:         consts.Device,
		SessionStorage: storage.NewSession(kvd, true),
		RetryInterval:  time.Second,
	})

	return c.Run(ctx, func(ctx context.Context) error {
		if err := c.Ping(ctx); err != nil {
			return err
		}

		color.Blue("Login...")
		color.Yellow("WARN: If data exists in the namespace, data will be overwritten")

		phone, err := input.DefaultUI().Ask(color.BlueString("Enter your phone number:"), &input.Options{
			Default:  color.CyanString("+86 12345678900"),
			Loop:     true,
			Required: true,
		})
		if err != nil {
			return err
		}
		color.Blue("Send code...")

		flow := auth.NewFlow(termAuth{phone: phone}, auth.SendCodeOptions{})
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
