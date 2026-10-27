//go:build !wasip1

package state

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cockroachdb/pebble"
)

var (
	db     *pebble.DB
	dbOnce sync.Once
	dbErr  error
)

// ErrNotFound is defined in kv.go.

func DB() (*pebble.DB, error) {
	dbOnce.Do(func() {
		d, err := Dir()
		if err != nil {
			dbErr = err
			return
		}
		db, dbErr = pebble.Open(filepath.Join(d, "pebble"), &pebble.Options{})
	})
	return db, dbErr
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}

// ErrNotFound is defined in kv.go (shared with the WASI KV fallback).

func Get(user, key string) ([]byte, error) {
	kdb, err := DB()
	if err != nil {
		return nil, err
	}
	val, closer, err := kdb.Get([]byte(userKey(user, key)))
	if errors.Is(err, pebble.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	out := make([]byte, len(val))
	copy(out, val)
	if err := closer.Close(); err != nil {
		return nil, err
	}
	return out, nil
}

func Set(user, key string, val []byte) error {
	kdb, err := DB()
	if err != nil {
		return err
	}
	return kdb.Set([]byte(userKey(user, key)), val, pebble.Sync)
}

func Delete(user, key string) error {
	kdb, err := DB()
	if err != nil {
		return err
	}
	return kdb.Delete([]byte(userKey(user, key)), pebble.Sync)
}

// List returns every entry under "<user>\x1f<prefix>" (prefix without the
// user separator, e.g. "skill").
func List(user, prefix string) ([]KV, error) {
	kdb, err := DB()
	if err != nil {
		return nil, err
	}
	start := []byte(userKey(user, prefix))
	end := append(append([]byte{}, start...), 0xff)
	var out []KV
	iter, err := kdb.NewIter(&pebble.IterOptions{LowerBound: start, UpperBound: end})
	if err != nil {
		return nil, err
	}
	for iter.First(); iter.Valid(); iter.Next() {
		v := make([]byte, len(iter.Value()))
		copy(v, iter.Value())
		out = append(out, KV{Key: strings.TrimPrefix(string(iter.Key()), userKey(user, prefix)+""), Val: v})
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return out, nil
}

func Users() ([]string, error) {
	kdb, err := DB()
	if err != nil {
		return nil, err
	}
	var out []string
	iter, err := kdb.NewIter(nil)
	if err != nil {
		return nil, err
	}
	for iter.First(); iter.Valid(); iter.Next() {
		u, _, ok := strings.Cut(string(iter.Key()), userSep)
		if ok && (len(out) == 0 || out[len(out)-1] != u) {
			out = append(out, u)
		}
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return out, nil
}

var _ = os.Getpid
