package login

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/skip2/go-qrcode"
	"github.com/spf13/viper"

	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/tclient"
)

func QR(ctx context.Context) error {
	kvd, err := kv.From(ctx).Open(viper.GetString(consts.FlagNamespace))
	if err != nil {
		return errors.Wrap(err, "open kv")
	}

	if err = kvd.Set(key.App(), []byte(tclient.AppDesktop)); err != nil {
		return errors.Wrap(err, "set app")
	}

	d := tg.NewUpdateDispatcher()

	c, err := tclient.New(ctx, tclient.Options{
		KV:               kvd,
		Proxy:            viper.GetString(consts.FlagProxy),
		NTP:              viper.GetString(consts.FlagNTP),
		ReconnectTimeout: viper.GetDuration(consts.FlagReconnectTimeout),
		Test:             viper.GetString(consts.FlagTest) != "",
		UpdateHandler:    d,
	}, true)
	if err != nil {
		return errors.Wrap(err, "create client")
	}

	return c.Run(ctx, func(ctx context.Context) error {
		color.Blue("Scan QR code with your Telegram app...")

		var lines int
		_, err = c.QR().Auth(ctx, qrlogin.OnLoginToken(d), func(ctx context.Context, token qrlogin.Token) error {
			qr, err := qrcode.New(token.URL(), qrcode.Medium)
			if err != nil {
				return errors.Wrap(err, "create qr")
			}
			code := qr.ToSmallString(false)
			lines = strings.Count(code, "\n")

			fmt.Print(code)
			fmt.Print(strings.Repeat(text.CursorUp.Sprint(), lines))
			return nil
		})

		// clear qrcode
		out := &strings.Builder{}
		for i := 0; i < lines; i++ {
			out.WriteString(text.EraseLine.Sprint())
			out.WriteString(text.CursorDown.Sprint())
		}
		out.WriteString(text.CursorUp.Sprintn(lines))
		fmt.Print(out.String())

		if err != nil {
			// https://core.telegram.org/api/auth#2fa
			if !tgerr.Is(err, "SESSION_PASSWORD_NEEDED") {
				return errors.Wrap(err, "qr auth")
			}

			pwd := ""
			prompt := &survey.Password{
				Message: "Enter 2FA Password:",
			}

			if err = survey.AskOne(prompt, &pwd, survey.WithValidator(survey.Required)); err != nil {
				return errors.Wrap(err, "2fa password")
			}

			if _, err = c.Auth().Password(ctx, pwd); err != nil {
				return errors.Wrap(err, "2fa auth")
			}
		}

		user, err := c.Self(ctx)
		if err != nil {
			return errors.Wrap(err, "get self")
		}

		fmt.Print(text.EraseLine.Sprint())
		color.Green("Login successfully! ID: %d, Username: %s", user.ID, user.Username)
		return nil
	})
}
