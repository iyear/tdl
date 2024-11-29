package tctx

import (
	"context"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/core/storage"
)

type kvKey struct{}

func KV(ctx context.Context) storage.Storage {
	return ctx.Value(kvKey{}).(storage.Storage)
}

func WithKV(ctx context.Context, kv storage.Storage) context.Context {
	return context.WithValue(ctx, kvKey{}, kv)
}

type poolKey struct{}

func Pool(ctx context.Context) dcpool.Pool {
	return ctx.Value(poolKey{}).(dcpool.Pool)
}

func WithPool(ctx context.Context, pool dcpool.Pool) context.Context {
	return context.WithValue(ctx, poolKey{}, pool)
}
