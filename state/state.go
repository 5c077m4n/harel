package state

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var mu sync.Mutex

var defaultSoul = []byte("You are Harel, a helpful assistant.")

const (
	PendingLimit   = 20
	VacuumInterval = 5
	MaxMemoryLines = 100
)

func Dir() (string, error) {
	base := os.Getenv("XDG_STATE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "state")
	}
	d := filepath.Join(base, "harel")
	if err := os.MkdirAll(d, 0o755); err != nil {
		return "", err
	}
	return d, nil
}

func Load() (string, error) {
	d, err := Dir()
	if err != nil {
		return "", err
	}
	soul, err := os.ReadFile(filepath.Join(d, "soul.md"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if len(soul) == 0 {
		soul = defaultSoul
	}
	mem, err := os.ReadFile(filepath.Join(d, "memory.md"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	pending, err := os.ReadFile(filepath.Join(d, "pending.md"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	return string(soul) + "\n" + string(mem) + string(pending), nil
}

func AppendMemory(line string) error {
	mu.Lock()
	defer mu.Unlock()
	d, err := Dir()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(d, "pending.md"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}

func PendingCount() (int, error) {
	d, err := Dir()
	if err != nil {
		return 0, err
	}
	b, err := os.ReadFile(filepath.Join(d, "pending.md"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return 0, err
	}
	if len(b) == 0 {
		return 0, nil
	}
	return strings.Count(strings.TrimRight(string(b), "\n"), "\n") + 1, nil
}

func Vacuum() error {
	mu.Lock()
	defer mu.Unlock()
	d, err := Dir()
	if err != nil {
		return err
	}
	if err := os.Remove(filepath.Join(d, "pending.md")); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	lines, err := readLines(filepath.Join(d, "memory.md"))
	if err != nil {
		return err
	}
	seen := make(map[string]bool, len(lines))
	var keep []string
	for _, l := range lines {
		t := strings.TrimSpace(l)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		keep = append(keep, t)
	}
	if len(keep) > MaxMemoryLines {
		keep = keep[len(keep)-MaxMemoryLines:]
	}
	return writeLines(filepath.Join(d, "memory.md"), keep)
}

func Consolidate(gen func(prompt string) (string, error)) error {
	mu.Lock()
	defer mu.Unlock()
	d, err := Dir()
	if err != nil {
		return err
	}
	pendingPath := filepath.Join(d, "pending.md")
	pending, err := os.ReadFile(pendingPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if len(pending) == 0 {
		return nil
	}
	memPath := filepath.Join(d, "memory.md")
	mem, err := os.ReadFile(memPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	prompt := "Merge the pending facts into the memory. Reply with ONLY the full updated memory, one fact per line, no preamble.\n\nCURRENT MEMORY:\n" + string(mem) + "\n\nPENDING FACTS:\n" + string(pending)
	out, err := gen(prompt)
	if err != nil {
		return err
	}
	var updated []string
	for _, l := range strings.Split(strings.TrimRight(out, "\n"), "\n") {
		if strings.TrimSpace(l) != "" {
			updated = append(updated, l)
		}
	}
	ops := Diff(splitLines(mem), updated)
	if len(ops) > 0 {
		if err := Apply(memPath, ops); err != nil {
			return err
		}
	}
	return os.Remove(pendingPath)
}
