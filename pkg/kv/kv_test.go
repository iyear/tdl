package kv

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func forEachStorage(t *testing.T, fn func(e Storage, t *testing.T)) {
	storages := map[Driver]map[string]any{
		DriverBolt:   {"path": t.TempDir()},
		DriverLegacy: {"path": filepath.Join(t.TempDir(), "test.db")},
		DriverFile:   {"path": filepath.Join(t.TempDir(), "test.json")},
	}

	for driver, opts := range storages {
		storage, err := New(driver, opts)
		require.NoError(t, err)

		t.Run(driver.String(), func(t *testing.T) {
			fn(storage, t)
		})
		assert.NoError(t, storage.Close())
	}
}

func TestNew(t *testing.T) {
	tests := map[Driver][]struct {
		name    string
		opts    map[string]any
		wantErr bool
	}{
		DriverBolt: {
			{name: "valid", opts: map[string]any{"path": t.TempDir()}, wantErr: false},
			{name: "invalid", opts: map[string]any{"path": ""}, wantErr: true},
		},
		DriverLegacy: {
			{name: "valid", opts: map[string]any{"path": filepath.Join(t.TempDir(), "test.db")}, wantErr: false},
			{name: "invalid", opts: map[string]any{"path": ""}, wantErr: true},
		},
		DriverFile: {
			{name: "valid", opts: map[string]any{"path": filepath.Join(t.TempDir(), "test.json")}, wantErr: false},
		},
		Driver("unknown"): {
			{name: "unknown", opts: map[string]any{"path": ""}, wantErr: true},
		},
	}

	for driver, tests := range tests {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("%v/%s", driver, tt.name), func(t *testing.T) {
				kv, err := New(driver, tt.opts)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Nil(t, kv)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, kv)
					assert.NoError(t, kv.Close())
				}
			})
		}
	}
}

func TestStorage_Open(t *testing.T) {
	forEachStorage(t, func(e Storage, t *testing.T) {
		for _, ns := range []string{"foo", "bar", "foo"} {
			kv, err := e.Open(ns)
			require.NoError(t, err)
			require.NotNil(t, kv)
		}
	})
}

func TestStorage_Namespaces(t *testing.T) {
	namespaces := []string{"foo", "bar", "baz"}

	forEachStorage(t, func(e Storage, t *testing.T) {
		for _, ns := range namespaces {
			kv, err := e.Open(ns)
			require.NoError(t, err)
			require.NotNil(t, kv)
		}

		ns, err := e.Namespaces()
		require.NoError(t, err)
		require.ElementsMatch(t, namespaces, ns)
	})
}

func TestStorage_MigrateTo(t *testing.T) {
	meta := Meta{
		"foo": {
			"1": []byte("2"),
			"3": []byte("4"),
			"5": []byte("6"),
		},
		"bar": {
			"7":  []byte("8"),
			"9":  []byte("10"),
			"11": []byte("12"),
		},
	}

	forEachStorage(t, func(e Storage, t *testing.T) {
		for ns, pairs := range meta {
			kv, err := e.Open(ns)
			require.NoError(t, err)
			require.NotNil(t, kv)

			for key, value := range pairs {
				require.NoError(t, kv.Set(context.TODO(), key, value))
			}
		}

		m, err := e.MigrateTo()
		assert.NoError(t, err)
		assert.Equal(t, meta, m)
	})
}

func TestStorage_MigrateFrom(t *testing.T) {
	meta := Meta{
		"foo": {
			"1": []byte("2"),
			"3": []byte("4"),
			"5": []byte("6"),
		},
		"bar": {
			"7":  []byte("8"),
			"9":  []byte("10"),
			"11": []byte("12"),
		},
	}

	forEachStorage(t, func(e Storage, t *testing.T) {
		require.NoError(t, e.MigrateFrom(meta))

		for ns, pairs := range meta {
			kv, err := e.Open(ns)
			require.NoError(t, err)
			require.NotNil(t, kv)

			for key, value := range pairs {
				v, err := kv.Get(context.TODO(), key)
				require.NoError(t, err)
				require.Equal(t, value, v)
			}
		}
	})
}
