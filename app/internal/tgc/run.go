package tgc

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func RunWithAuth(ctx context.Context, client *telegram.Client, f func(ctx context.Context) error) error {
	if viper.GetString(consts.FlagTest) != "" {
		return runWithTest(ctx, client, f)
	}

	return client.Run(ctx, func(ctx context.Context) error {
		status, err := client.Auth().Status(ctx)
		if err != nil {
			return err
		}
		if !status.Authorized {
			return fmt.Errorf("not authorized. please login first")
		}

		logger.From(ctx).Info("Authorized",
			zap.Int64("id", status.User.ID),
			zap.String("username", status.User.Username))

		return f(ctx)
	})
}

func runWithTest(ctx context.Context, client *telegram.Client, f func(ctx context.Context) error) error {
	return client.Run(ctx, func(ctx context.Context) error {
		authClient := auth.NewClient(client.API(), rand.Reader, telegram.TestAppID, telegram.TestAppHash)

		reader := bytes.NewBufferString(viper.GetString(consts.FlagTest)) // stable account

		if err := auth.NewFlow(
			auth.Test(reader, 2),
			auth.SendCodeOptions{},
		).Run(ctx, authClient); err != nil {
			return err
		}

		return f(ctx)
	})
}
