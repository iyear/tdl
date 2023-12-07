package storage

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/gotd/td/telegram/peers"

	"github.com/iyear/tdl/pkg/key"
	"github.com/iyear/tdl/pkg/kv"
)

type Peers struct {
	kv kv.KV
}

func NewPeers(kv kv.KV) peers.Storage {
	return &Peers{kv: kv}
}

func (p *Peers) Save(_ context.Context, _key peers.Key, value peers.Value) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return p.kv.Set(key.PeersKey(_key), bytes)
}

func (p *Peers) Find(_ context.Context, _key peers.Key) (peers.Value, bool, error) {
	data, err := p.kv.Get(key.PeersKey(_key))
	if err != nil {
		if errors.Is(err, kv.ErrNotFound) {
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

func (p *Peers) SavePhone(_ context.Context, phone string, _key peers.Key) error {
	bytes, err := json.Marshal(_key)
	if err != nil {
		return err
	}

	return p.kv.Set(key.PeersPhone(phone), bytes)
}

func (p *Peers) FindPhone(ctx context.Context, phone string) (peers.Key, peers.Value, bool, error) {
	data, err := p.kv.Get(key.PeersPhone(phone))
	if err != nil {
		if errors.Is(err, kv.ErrNotFound) {
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

func (p *Peers) GetContactsHash(_ context.Context) (int64, error) {
	data, err := p.kv.Get(key.PeersContactsHash())
	if err != nil {
		if errors.Is(err, kv.ErrNotFound) {
			return 0, nil
		}
		return 0, err
	}

	return strconv.ParseInt(string(data), 10, 64)
}

func (p *Peers) SaveContactsHash(_ context.Context, hash int64) error {
	return p.kv.Set(key.PeersContactsHash(), []byte(strconv.FormatInt(hash, 10)))
}
