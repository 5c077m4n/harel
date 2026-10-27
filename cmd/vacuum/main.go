package main

import (
	"os"

	"github.com/5c077m4n/harel/state"
)

func main() {
	if err := state.Vacuum(state.DefaultUser()); err != nil {
		panic(err)
	}
	os.Exit(0)
}
