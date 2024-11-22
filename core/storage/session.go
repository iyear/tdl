package storage

import (
	"context"
	"errors"

	"github.com/gotd/td/telegram"

	"github.com/iyear/tdl/core/storage/keygen"
)

type Session struct {
	kv    Storage
	login bool
}

func NewSession(kv Storage, login bool) telegram.SessionStorage {
	return &Session{kv: kv, login: login}
}

func (s *Session) LoadSession(ctx context.Context) ([]byte, error) {
	if s.login {
		return nil, nil
	}

	b, err := s.kv.Get(ctx, s.key())
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func (s *Session) StoreSession(ctx context.Context, data []byte) error {
	return s.kv.Set(ctx, s.key(), data)
}

func (s *Session) key() string {
	return keygen.New("session")
}
