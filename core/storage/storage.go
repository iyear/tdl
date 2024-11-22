package storage

import (
	"context"

	"github.com/go-faster/errors"
)

type Storage interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte) error
	Delete(ctx context.Context, key string) error
}

var ErrNotFound = errors.New("key not found")
