package logctx

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

func From(ctx context.Context) *zap.Logger {
	return ctx.Value(ctxKey{}).(*zap.Logger)
}

func With(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}

func Named(ctx context.Context, name string) context.Context {
	return With(ctx, From(ctx).Named(name))
}
