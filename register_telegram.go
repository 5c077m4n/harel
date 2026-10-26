//go:build telegram

package main

import (
	"github.com/5c077m4n/harel/agent"
	"github.com/5c077m4n/harel/plugins/telegram"
)

func register(a *agent.Agent) {
	a.Register(telegram.New())
}
