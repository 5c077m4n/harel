//go:build !telegram && !http

package main

import "github.com/5c077m4n/harel/agent"

func register(a *agent.Agent) {}
