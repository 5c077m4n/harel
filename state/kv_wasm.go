//go:build wasip1

package state

import (
	"bytes"
	"strings"
	"sync"
)

// In-process KV fallback for WASI builds; pebble is native-only.
var (
	kvMu  sync.Mutex
	kvMap = map[string][]byte{}
)

func DB() (interface{}, error) { return nil, nil }

func CloseDB() error { return nil }

func Get(user, key string) ([]byte, error) {
	kvMu.Lock()
	defer kvMu.Unlock()
	v, ok := kvMap[userKey(user, key)]
	if !ok {
		return nil, ErrNotFound
	}
	return append([]byte{}, v...), nil
}

func Set(user, key string, val []byte) error {
	kvMu.Lock()
	defer kvMu.Unlock()
	kvMap[userKey(user, key)] = append([]byte{}, val...)
	return nil
}

func Delete(user, key string) error {
	kvMu.Lock()
	defer kvMu.Unlock()
	delete(kvMap, userKey(user, key))
	return nil
}

func List(user, prefix string) ([]KV, error) {
	kvMu.Lock()
	defer kvMu.Unlock()
	p := userKey(user, prefix)
	var out []KV
	for k, v := range kvMap {
		if strings.HasPrefix(k, p) {
			out = append(out, KV{Key: strings.TrimPrefix(k, p), Val: append([]byte{}, v...)})
		}
	}
	return out, nil
}

func Users() ([]string, error) {
	kvMu.Lock()
	defer kvMu.Unlock()
	seen := map[string]bool{}
	var out []string
	for k := range kvMap {
		u, _, ok := strings.Cut(k, userSep)
		if ok && !seen[u] {
			seen[u] = true
			out = append(out, u)
		}
	}
	return out, nil
}

var _ = bytes.Equal
