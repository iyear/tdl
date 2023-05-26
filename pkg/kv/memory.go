package kv

import "sync"

type Memory struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewMemory() *Memory {
	return &Memory{
		data: make(map[string][]byte),
	}
}

func (m *Memory) Get(key string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return data, nil
}

func (m *Memory) Set(key string, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = value
	return nil
}

func (m *Memory) Delete(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.data, key)
	return nil
}
