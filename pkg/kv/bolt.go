package kv

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/go-faster/errors"
	"github.com/mitchellh/mapstructure"
	"go.etcd.io/bbolt"
	"go.uber.org/multierr"

	"github.com/iyear/tdl/core/storage"
	"github.com/iyear/tdl/pkg/validator"
)

func init() {
	register(DriverBolt, func(m map[string]any) (Storage, error) { return newBolt(m) })
}

type bolt struct {
	path string
	dbs  map[string]*bbolt.DB
	mu   *sync.Mutex
}

func newBolt(opts map[string]any) (*bolt, error) {
	type options struct {
		Path string `validate:"required" mapstructure:"path"`
	}

	var o options
	if err := mapstructure.WeakDecode(opts, &o); err != nil {
		return nil, errors.Wrap(err, "decode options")
	}

	if err := validator.Struct(&o); err != nil {
		return nil, errors.Wrap(err, "validate options")
	}

	if err := os.MkdirAll(o.Path, 0o755); err != nil {
		return nil, errors.Wrap(err, "create dir")
	}

	return &bolt{
		path: o.Path,
		dbs:  make(map[string]*bbolt.DB),
		mu:   &sync.Mutex{},
	}, nil
}

func (b *bolt) Name() string {
	return DriverBolt.String()
}

func (b *bolt) MigrateTo() (Meta, error) {
	meta := make(Meta)

	if err := b.walk(func(path string) (rerr error) {
		ns := filepath.Base(path)
		meta[ns] = make(map[string][]byte)

		db, err := b.open(ns)
		if err != nil {
			return errors.Wrap(err, "open")
		}

		return db.db.View(func(tx *bbolt.Tx) error {
			return tx.Bucket(db.ns).ForEach(func(k, v []byte) error {
				meta[ns][string(k)] = v
				return nil
			})
		})
	}); err != nil {
		return nil, errors.Wrap(err, "walk")
	}

	return meta, nil
}

func (b *bolt) MigrateFrom(meta Meta) error {
	for ns, pairs := range meta {
		db, err := b.open(ns)
		if err != nil {
			return errors.Wrap(err, "open")
		}

		if err = db.db.Update(func(tx *bbolt.Tx) error {
			bk, err := tx.CreateBucketIfNotExists(db.ns)
			if err != nil {
				return errors.Wrap(err, "create bucket")
			}
			for key, value := range pairs {
				if err = bk.Put([]byte(key), value); err != nil {
					return errors.Wrap(err, "put")
				}
			}
			return nil
		}); err != nil {
			return errors.Wrap(err, "update")
		}
	}

	return nil
}

func (b *bolt) Namespaces() ([]string, error) {
	namespaces := make([]string, 0)
	if err := b.walk(func(path string) error {
		namespaces = append(namespaces, filepath.Base(path))
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "walk")
	}

	return namespaces, nil
}

func (b *bolt) walk(fn func(path string) error) error {
	return filepath.Walk(b.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "walk")
		}
		if info.IsDir() {
			return nil
		}

		return fn(path)
	})
}

func (b *bolt) Open(ns string) (storage.Storage, error) {
	return b.open(ns)
}

func (b *bolt) open(ns string) (*legacyKV, error) {
	if ns == "" {
		return nil, errors.New("namespace is required")
	}
	b.mu.Lock()
	defer b.mu.Unlock()

	if db, ok := b.dbs[ns]; ok {
		return &legacyKV{db: db, ns: []byte(ns)}, nil
	}

	db, err := bbolt.Open(filepath.Join(b.path, ns), os.ModePerm, boltOptions)
	if err != nil {
		return nil, errors.Wrap(err, "open db")
	}
	if err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(ns))
		return err
	}); err != nil {
		return nil, errors.Wrap(err, "create bucket")
	}

	b.dbs[ns] = db

	return &legacyKV{db: db, ns: []byte(ns)}, nil
}

func (b *bolt) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var err error
	for _, db := range b.dbs {
		err = multierr.Append(err, db.Close())
	}

	return err
}
