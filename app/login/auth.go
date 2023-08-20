package login

import (
	"context"
	"errors"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
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
	prompt := &survey.Input{
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
