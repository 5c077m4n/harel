package main

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/5c077m4n/harel/agent"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a := agent.New()
	register(a)

	if err := a.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
}
