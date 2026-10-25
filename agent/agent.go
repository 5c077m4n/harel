package agent

import (
	"context"
	"strings"
	"sync"

	"github.com/5c077m4n/harel/llm"
	"github.com/5c077m4n/harel/plugin"
	"github.com/5c077m4n/harel/state"
)

type Agent struct {
	llm     *llm.Client
	plugins []plugin.Plugin
}

func New() *Agent {
	return &Agent{
		llm: llm.New(),
	}
}

func (a *Agent) Register(p plugin.Plugin) {
	a.plugins = append(a.plugins, p)
}

func (a *Agent) Process(msg string) (string, error) {
	systemPrompt, err := state.Load()
	if err != nil {
		return "", err
	}
	reply, err := a.llm.Chat([]llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: msg},
	})
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(reply, "\n") {
		if strings.HasPrefix(line, "MEMORY: ") {
			if err := state.AppendMemory(strings.TrimPrefix(line, "MEMORY: ")); err != nil {
				return "", err
			}
		}
	}
	if n, err := state.PendingCount(); err != nil {
		return "", err
	} else if n >= state.PendingLimit {
		if err := state.Consolidate(func(prompt string) (string, error) {
			return a.llm.Chat([]llm.Message{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: prompt},
			})
		}); err != nil {
			return "", err
		}
	}
	return reply, nil
}

func (a *Agent) Run(ctx context.Context) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errs := make(chan error, len(a.plugins))
	var wg sync.WaitGroup
	for _, p := range a.plugins {
		in := make(chan string)
		out := make(chan string)
		wg.Add(1)
		go func(pl plugin.Plugin) {
			defer wg.Done()
			if err := pl.Start(runCtx, in, out); err != nil {
				errs <- err
			}
		}(p)

		go func(inCh <-chan string, outCh chan<- string) {
			for {
				select {
				case <-runCtx.Done():
					return
				case msg := <-inCh:
					reply, err := a.Process(msg)
					if err != nil {
						errs <- err
						return
					}
					select {
					case outCh <- reply:
					case <-runCtx.Done():
						return
					}
				}
			}
		}(in, out)
	}

	select {
	case err := <-errs:
		cancel()
		wg.Wait()
		return err
	case <-runCtx.Done():
		wg.Wait()
		return runCtx.Err()
	}
}
