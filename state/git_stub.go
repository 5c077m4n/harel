//go:build wasip1 && !tinygo

package state

func commitChanges(msg string, paths ...string) error {
	return nil
}
