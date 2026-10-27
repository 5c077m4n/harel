package state

import (
	"errors"
	"strings"
)

var ErrNotFound = errors.New("not found")

// KV is a single stored entry returned by List.
type KV struct {
	Key string
	Val []byte
}

const userSep = "\x1f"

// userKey builds a per-user key: "<user>\x1f<path>" so users are top-level
// prefixes and each user's entries iterate as one contiguous range.
func userKey(user, key string) string {
	return strings.TrimSpace(user) + userSep + strings.TrimPrefix(key, "/")
}
