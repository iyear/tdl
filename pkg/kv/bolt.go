package kv

import (
	"go.etcd.io/bbolt"
)

type Bolt struct {
	ns []byte
	db *bbolt.DB
}

func (b *Bolt) Get(key string) ([]byte, error) {
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

func (b *Bolt) Set(key string, val []byte) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(b.ns).Put([]byte(key), val)
	})
}

// Delete removes a key from the bucket. If the key does not exist then nothing is done and a nil error is returned
func (b *Bolt) Delete(key string) error {
	return b.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(b.ns).Delete([]byte(key))
	})
}
