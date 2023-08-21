package kv

import (
	"encoding/json"
	"os"
	"sync"
)

type File struct {
	path string
	mu   sync.Mutex
}

func NewFile(path string) (*File, error) {
	_, err := os.Stat(path)
	if err == nil {
		return &File{path: path}, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	if err = os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		return nil, err
	}

	return &File{path: path}, nil
}

func (f *File) Get(key string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	m, err := f.read()
	if err != nil {
		return nil, err
	}

	if val, ok := m[key]; ok {
		return val, nil
	}
	return nil, ErrNotFound
}

func (f *File) Set(key string, value []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	m, err := f.read()
	if err != nil {
		return err
	}

	m[key] = value

	return f.write(m)
}

func (f *File) Delete(key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	m, err := f.read()
	if err != nil {
		return err
	}

	delete(m, key)

	return f.write(m)
}

func (f *File) read() (map[string][]byte, error) {
	bytes, err := os.ReadFile(f.path)
	if err != nil {
		return nil, err
	}

	m := make(map[string][]byte)
	if err = json.Unmarshal(bytes, &m); err != nil {
		return nil, err
	}

	return m, nil
}

func (f *File) write(m map[string][]byte) error {
	bytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	return os.WriteFile(f.path, bytes, 0o644)
}
