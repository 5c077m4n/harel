//go:build http

package main

import (
	"github.com/5c077m4n/harel/agent"
	"github.com/5c077m4n/harel/plugins/server"
)

func register(a *agent.Agent) {
	a.Register(server.New())
}
