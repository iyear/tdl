package login

import (
	"context"
	"errors"
	"github.com/fatih/color"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/tcnksm/go-input"
	"strings"
)

// noSignUp can be embedded to prevent signing up.
type noSignUp struct{}

func (c noSignUp) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("searchx don't support sign up Telegram account")
}

func (c noSignUp) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

// termAuth implements authentication via terminal.
type termAuth struct {
	noSignUp
}

func (a termAuth) Phone(_ context.Context) (string, error) {
	phone, err := input.DefaultUI().Ask(color.BlueString("Enter your phone number:"), &input.Options{
		Default:  color.CyanString("+86 12345678900"),
		Loop:     true,
		Required: true,
	})
	if err != nil {
		return "", err
	}

	color.Blue("Sending Code...")
	return strings.TrimSpace(phone), nil
}

func (a termAuth) Password(_ context.Context) (string, error) {
	pwd, err := input.DefaultUI().Ask(color.BlueString("Enter 2FA Password:"), &input.Options{
		Required: true,
		Loop:     true,
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(pwd), nil
}

func (a termAuth) Code(_ context.Context, _ *tg.AuthSentCode) (string, error) {
	code, err := input.DefaultUI().Ask(color.BlueString("Enter Code:"), &input.Options{
		Required: true,
		Loop:     true,
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}
