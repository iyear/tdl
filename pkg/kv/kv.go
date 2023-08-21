package kv

import (
	"errors"
	"os"
	"time"

	"go.etcd.io/bbolt"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/pkg/validator"
)

var ErrNotFound = errors.New("key not found")

type KV interface {
	Get(key string) ([]byte, error)
	Set(key string, value []byte) error
	Delete(key string) error
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

// Namespaces returns all namespaces in the database
func Namespaces(path string) (_ []string, rerr error) {
	db, err := bbolt.Open(path, os.ModePerm, &bbolt.Options{
		Timeout:  time.Second,
		ReadOnly: true,
	})
	if err != nil {
		return nil, err
	}
	defer multierr.AppendInvoke(&rerr, multierr.Close(db))

	namespaces := make([]string, 0)
	err = db.View(func(tx *bbolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bbolt.Bucket) error {
			namespaces = append(namespaces, string(name))
			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return namespaces, nil
}
