package state

import (
	"errors"
	"os"
	"path/filepath"
)

var defaultSoul = []byte("You are Harel, a helpful assistant.")

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
	return string(soul) + "\n" + string(mem), nil
}

func AppendMemory(line string) error {
	d, err := Dir()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(d, "memory.md"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(line + "\n")
	return err
}
