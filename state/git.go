//go:build !wasip1 || tinygo

package state

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

func commitChanges(msg string, paths ...string) error {
	d, err := Dir()
	if err != nil {
		return err
	}
	repo, err := git.PlainOpen(d)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		repo, err = git.PlainInit(d, false)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	for _, p := range paths {
		rel, err := filepath.Rel(d, p)
		if err != nil {
			return err
		}
		if _, statErr := os.Stat(p); errors.Is(statErr, os.ErrNotExist) {
			idx, err := repo.Storer.Index()
			if err != nil {
				return err
			}
			tracked := false
			for _, e := range idx.Entries {
				if e.Name == rel {
					tracked = true
					break
				}
			}
			if tracked {
				if _, err := w.Remove(rel); err != nil {
					return err
				}
			}
		} else if statErr != nil {
			return statErr
		} else if _, err := w.Add(rel); err != nil {
			return err
		}
	}
	st, err := w.Status()
	if err != nil {
		return err
	}
	if st.IsClean() {
		return nil
	}
	_, err = w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{Name: "harel", Email: "harel@localhost", When: time.Now()},
	})
	return err
}
