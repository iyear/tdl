package login

import (
	"context"
	"crypto/rand"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/cenkalti/backoff/v4"
	"github.com/fatih/color"
	"github.com/go-faster/errors"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
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

	if err = kvd.Set(key.App(), []byte(tclient.AppDesktop)); err != nil {
		return errors.Wrap(err, "set app")
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

		flow := auth.NewFlow(termAuth{}, auth.SendCodeOptions{})
		if err = c.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		user, err := c.Self(ctx)
		if err != nil {
			return err
		}

		color.Green("Login successfully! ID: %d, Username: %s", user.ID, user.Username)

		return nil
	})
}

// noSignUp can be embedded to prevent signing up.
type noSignUp struct{}

func (c noSignUp) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("don't support sign up Telegram account")
}

func (c noSignUp) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

// termAuth implements authentication via terminal.
type termAuth struct {
	noSignUp
}

func (a termAuth) Phone(_ context.Context) (string, error) {
	phone := ""
	prompt := &survey.Input{
		Message: "Enter your phone number:",
		Default: "+86 12345678900",
	}

	if err := survey.AskOne(prompt, &phone, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	color.Blue("Sending Code...")
	return strings.TrimSpace(phone), nil
}

func (a termAuth) Password(_ context.Context) (string, error) {
	pwd := ""
	prompt := &survey.Password{
		Message: "Enter 2FA Password:",
	}

	if err := survey.AskOne(prompt, &pwd, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	return strings.TrimSpace(pwd), nil
}

func (a termAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	code := ""
	prompt := &survey.Input{
		Message: "Enter Code:",
	}

	if err := survey.AskOne(prompt, &code, survey.WithValidator(survey.Required)); err != nil {
		return "", err
	}

	return strings.TrimSpace(code), nil
}
