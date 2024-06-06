package tctx

import (
	"context"

	"github.com/iyear/tdl/core/dcpool"
	"github.com/iyear/tdl/pkg/kv"
)

type kvKey struct{}

func KV(ctx context.Context) kv.KV {
	return ctx.Value(kvKey{}).(kv.KV)
}

func WithKV(ctx context.Context, kv kv.KV) context.Context {
	return context.WithValue(ctx, kvKey{}, kv)
}

type poolKey struct{}

func Pool(ctx context.Context) dcpool.Pool {
	return ctx.Value(poolKey{}).(dcpool.Pool)
}

func WithPool(ctx context.Context, pool dcpool.Pool) context.Context {
	return context.WithValue(ctx, poolKey{}, pool)
}
