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

type Options struct {
	Path string `validate:"required"`
	NS   string `validate:"required"`
}

type KV struct {
	ns []byte
	db *bbolt.DB
}

func New(opts Options) (*KV, error) {
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

	return &KV{db: db, ns: []byte(opts.NS)}, nil
}

func (b *KV) Get(key string) ([]byte, error) {
	var val []byte

	if err := b.db.View(func(tx *bbolt.Tx) error {
		val = tx.Bucket(b.ns).Get([]byte(key))
		return nil
	}); err != nil {
		return nil, err
	}

	if val == nil {
		return nil, ErrNotFound
	}
	return val, nil
}

func (b *KV) Set(key string, val []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(b.ns).Put([]byte(key), val)
	})
}
