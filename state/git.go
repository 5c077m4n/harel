package state

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func commitChanges(msg string, paths ...string) error {
	d, err := Dir()
	if err != nil {
		return err
	}
	git := func(args ...string) ([]byte, error) {
		cmd := exec.Command("git", append([]string{"-C", d}, args...)...)
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=harel", "GIT_AUTHOR_EMAIL=harel@localhost",
			"GIT_COMMITTER_NAME=harel", "GIT_COMMITTER_EMAIL=harel@localhost",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			return out, err
		}
		return out, nil
	}
	if _, err := git("rev-parse", "--git-dir"); err != nil {
		if _, err := git("init", "-q"); err != nil {
			return err
		}
	}
	for _, p := range paths {
		rel, err := filepath.Rel(d, p)
		if err != nil {
			return err
		}
		if _, err := os.Stat(p); os.IsNotExist(err) {
			if _, err := git("rm", "-q", "--ignore-unmatch", "--", rel); err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else if _, err := git("add", "--", rel); err != nil {
			return err
		}
	}
	out, err := git("status", "--porcelain")
	if err != nil {
		return err
	}
	if len(out) == 0 {
		return nil
	}
	_, err = git("commit", "-qm", msg, "--date", time.Now().Format(time.RFC3339))
	return err
}
