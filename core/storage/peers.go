package storage

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gotd/td/telegram/peers"

	"github.com/iyear/tdl/core/storage/keygen"
)

type Peers struct {
	kv Storage
}

func NewPeers(kv Storage) peers.Storage {
	return &Peers{kv: kv}
}

func (p *Peers) Save(ctx context.Context, key peers.Key, value peers.Value) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return p.kv.Set(ctx, p.key(key), bytes)
}

func (p *Peers) Find(ctx context.Context, key peers.Key) (peers.Value, bool, error) {
	data, err := p.kv.Get(ctx, p.key(key))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return peers.Value{}, false, nil
		}
		return peers.Value{}, false, err
	}

	var value peers.Value
	if err = json.Unmarshal(data, &value); err != nil {
		return peers.Value{}, false, err
	}

	return value, true, nil
}

func (p *Peers) SavePhone(ctx context.Context, phone string, _key peers.Key) error {
	bytes, err := json.Marshal(_key)
	if err != nil {
		return err
	}

	return p.kv.Set(ctx, p.phoneKey(phone), bytes)
}

func (p *Peers) FindPhone(ctx context.Context, phone string) (peers.Key, peers.Value, bool, error) {
	data, err := p.kv.Get(ctx, p.phoneKey(phone))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return peers.Key{}, peers.Value{}, false, nil
		}
		return peers.Key{}, peers.Value{}, false, err
	}

	var _key peers.Key
	if err = json.Unmarshal(data, &_key); err != nil {
		return peers.Key{}, peers.Value{}, false, err
	}

	value, found, err := p.Find(ctx, _key)
	if err != nil {
		return peers.Key{}, peers.Value{}, false, err
	}

	return _key, value, found, nil
}

func (p *Peers) GetContactsHash(ctx context.Context) (int64, error) {
	data, err := p.kv.Get(ctx, p.contactsKey())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return strconv.ParseInt(string(data), 10, 64)
}

func (p *Peers) SaveContactsHash(ctx context.Context, hash int64) error {
	return p.kv.Set(ctx, p.contactsKey(), []byte(strconv.FormatInt(hash, 10)))
}

func (p *Peers) key(key peers.Key) string {
	return keygen.New("peers", "key", key.Prefix, strconv.FormatInt(key.ID, 10))
}

func (p *Peers) phoneKey(phone string) string {
	return keygen.New("peers", "phone", phone)
}

func (p *Peers) contactsKey() string {
	return keygen.New("peers", "contacts", "hash")
}
