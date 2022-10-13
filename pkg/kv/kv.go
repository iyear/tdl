package kv

import (
	"errors"
	"github.com/iyear/tdl/pkg/validator"
	"go.etcd.io/bbolt"
	"os"
	"time"
)

var (
	ErrNotFound = errors.New("key not found")
)

type KV interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
}

type Options struct {
	NS   string `validate:"required"`
	Path string `validate:"required"`
}

func New(opts Options) (KV, error) {
	if err := validator.Struct(&opts); err != nil {
		return nil, err
	}

	db, err := bbolt.Open(opts.Path, os.ModePerm, &bbolt.Options{
		Timeout:      time.Second,
		NoGrowSync:   false,
		FreelistType: bbolt.FreelistArrayType,
	})
	if err != nil {
		return nil, err
	}

	if err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(opts.NS))
		return err
	}); err != nil {
		return nil, err
	}

	return &Bolt{db: db, ns: []byte(opts.NS)}, nil
}
