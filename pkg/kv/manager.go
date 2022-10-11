package kv

import (
	"context"
	"encoding/json"
	"github.com/gotd/td/telegram/peers"
	"github.com/iyear/tdl/pkg/key"
	"strconv"
)

func (b *KV) Save(_ context.Context, _key peers.Key, value peers.Value) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return b.Set(key.PeersKey(_key), bytes)
}

func (b *KV) Find(_ context.Context, _key peers.Key) (peers.Value, bool, error) {
	data, err := b.Get(key.PeersKey(_key))
	if err != nil {
		if err == ErrNotFound {
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

func (b *KV) SavePhone(_ context.Context, phone string, _key peers.Key) error {
	bytes, err := json.Marshal(_key)
	if err != nil {
		return err
	}

	return b.Set(key.PeersPhone(phone), bytes)
}

func (b *KV) FindPhone(ctx context.Context, phone string) (peers.Key, peers.Value, bool, error) {
	data, err := b.Get(key.PeersPhone(phone))
	if err != nil {
		if err == ErrNotFound {
			return peers.Key{}, peers.Value{}, false, nil
		}
		return peers.Key{}, peers.Value{}, false, err
	}

	var _key peers.Key
	if err = json.Unmarshal(data, &_key); err != nil {
		return peers.Key{}, peers.Value{}, false, err
	}

	value, found, err := b.Find(ctx, _key)
	if err != nil {
		return peers.Key{}, peers.Value{}, false, err
	}

	return _key, value, found, nil
}

func (b *KV) GetContactsHash(_ context.Context) (int64, error) {
	data, err := b.Get(key.PeersContactsHash())
	if err != nil {
		if err == ErrNotFound {
			return 0, nil
		}
		return 0, err
	}

	return strconv.ParseInt(string(data), 10, 64)
}

func (b *KV) SaveContactsHash(_ context.Context, hash int64) error {
	return b.Set(key.PeersContactsHash(), []byte(strconv.FormatInt(hash, 10)))
}
