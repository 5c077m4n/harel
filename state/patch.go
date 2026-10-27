package state

import (
	"errors"
	"strings"
)

type Op struct {
	Pos int
	Del int
	Add []string
}

func splitLines(b []byte) []string {
	s := strings.TrimRight(string(b), "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func linesBytes(lines []string) []byte {
	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	return []byte(sb.String())
}

func Diff(a, b []string) []Op {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if a[i] == b[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	var ops []Op
	i, j := 0, 0
	for i < n || j < m {
		if i < n && j < m && a[i] == b[j] {
			i++
			j++
			continue
		}
		op := Op{Pos: i}
		for i < n && (j >= m || lcs[i+1][j] >= lcs[i][j+1]) {
			op.Del++
			i++
		}
		for j < m && (i >= n || lcs[i+1][j] < lcs[i][j+1]) {
			op.Add = append(op.Add, b[j])
			j++
		}
		ops = append(ops, op)
	}
	return ops
}

// Apply patches the user's memory with the minimal ops from Diff, then
// persists the result to pebble.
func Apply(user string, ops []Op) error {
	mem, err := Get(user, contextKey(user, "memory"))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	lines := splitLines(mem)
	delta := 0
	for _, op := range ops {
		pos := op.Pos + delta
		if pos < 0 || pos+op.Del > len(lines) {
			return errors.New("patch out of range")
		}
		out := append([]string{}, lines[:pos]...)
		out = append(out, op.Add...)
		out = append(out, lines[pos+op.Del:]...)
		lines = out
		delta += len(op.Add) - op.Del
	}
	return Set(user, contextKey(user, "memory"), linesBytes(lines))
}
