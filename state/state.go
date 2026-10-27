package state

import (
	"errors"
	"fmt"
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

// DefaultUser is the user scope used when a caller doesn't specify one;
// overridden by HAREL_USER.
func DefaultUser() string {
	if u := os.Getenv("HAREL_USER"); u != "" {
		return u
	}
	return "default"
}

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

// Context keys (per user): "<user>\x1fcontext/soul", ".../memory", ".../pending".
// Skills: "<user>\x1fskill/<name>".
func contextKey(user, name string) string { return "context/" + name }
func skillKey(name string) string         { return "skill/" + name }

func Load(user string) (string, error) {
	mu.Lock()
	defer mu.Unlock()
	soul, err := Get(user, contextKey(user, "soul"))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", err
	}
	if len(soul) == 0 {
		soul = defaultSoul
	}
	mem, err := Get(user, contextKey(user, "memory"))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", err
	}
	pending, err := Get(user, contextKey(user, "pending"))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return "", err
	}
	return string(soul) + "\n" + string(mem) + string(pending), nil
}

func AppendMemory(user, line string) error {
	mu.Lock()
	defer mu.Unlock()
	pending, err := Get(user, contextKey(user, "pending"))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	return Set(user, contextKey(user, "pending"), append(pending, []byte(line+"\n")...))
}

func PendingCount(user string) (int, error) {
	pending, err := Get(user, contextKey(user, "pending"))
	if errors.Is(err, ErrNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if len(pending) == 0 {
		return 0, nil
	}
	return strings.Count(strings.TrimRight(string(pending), "\n"), "\n") + 1, nil
}

func Vacuum(user string) error {
	mu.Lock()
	defer mu.Unlock()
	mem, err := Get(user, contextKey(user, "memory"))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if err := Delete(user, contextKey(user, "pending")); err != nil {
		return err
	}
	seen := make(map[string]bool, len(splitLines(mem)))
	var keep []string
	for _, l := range splitLines(mem) {
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
	var sb strings.Builder
	for _, l := range keep {
		sb.WriteString(l)
		sb.WriteString("\n")
	}
	return Set(user, contextKey(user, "memory"), []byte(sb.String()))
}

func Consolidate(user string, gen func(prompt string) (string, error)) error {
	mu.Lock()
	defer mu.Unlock()
	pending, err := Get(user, contextKey(user, "pending"))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	if len(pending) == 0 {
		return nil
	}
	mem, err := Get(user, contextKey(user, "memory"))
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	prompt := fmt.Sprintf("Merge the pending facts into the memory. Reply with ONLY the full updated memory, one fact per line, no preamble.\n\nCURRENT MEMORY:\n%s\n\nPENDING FACTS:\n%s", string(mem), string(pending))
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
		if err := Apply(user, ops); err != nil {
			return err
		}
	}
	return Delete(user, contextKey(user, "pending"))
}

// Skills are stored per user under "<user>\x1fskill/<name>" as raw markdown.
func SetSkill(user, name string, md []byte) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("skill name required")
	}
	return Set(user, skillKey(name), md)
}

func GetSkill(user, name string) ([]byte, error) {
	return Get(user, skillKey(name))
}

func DeleteSkill(user, name string) error {
	return Delete(user, skillKey(name))
}

func ListSkills(user string) ([]string, error) {
	entries, err := List(user, "skill/")
	if err != nil {
		return nil, err
	}
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = strings.TrimPrefix(e.Key, "skill/")
	}
	return names, nil
}
