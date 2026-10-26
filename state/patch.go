package state

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Op struct {
	Pos int
	Del int
	Add []string
}

func readLines(path string) ([]string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return splitLines(b), nil
}

func splitLines(b []byte) []string {
	s := strings.TrimRight(string(b), "\n")
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func writeLines(path string, lines []string) error {
	tmp := path + ".tmp"
	var sb strings.Builder
	for _, l := range lines {
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	if err := os.WriteFile(tmp, []byte(sb.String()), 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		return err
	}
	return commitChanges("update "+filepath.Base(path), path)
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

func Apply(path string, ops []Op) error {
	lines, err := readLines(path)
	if err != nil {
		return err
	}
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
	return writeLines(path, lines)
}
