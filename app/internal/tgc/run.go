package tgc

import (
	"context"
	"fmt"

	"github.com/gotd/td/telegram"
	"go.uber.org/zap"

	"github.com/iyear/tdl/pkg/logger"
)

func RunWithAuth(ctx context.Context, client *telegram.Client, f func(ctx context.Context) error) error {
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
