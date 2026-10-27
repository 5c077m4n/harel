//go:build http

package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"sync"

	"github.com/5c077m4n/harel/plugin"
	"github.com/5c077m4n/harel/state"
)

type Plugin struct {
	addr string
}

func New() plugin.Plugin {
	return &Plugin{addr: os.Getenv("HAREL_HTTP_ADDR")}
}

func (p *Plugin) Name() string {
	return "server"
}

func (p *Plugin) Start(ctx context.Context, in chan<- plugin.Message, out <-chan plugin.Message) error {
	if p.addr == "" {
		return nil
	}
	var mu sync.Mutex
	mux := http.NewServeMux()
	mux.HandleFunc("POST /message", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			User string `json:"user"`
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if req.User == "" {
			req.User = state.DefaultUser()
		}
		mu.Lock()
		defer mu.Unlock()
		select {
		case in <- plugin.Message{User: req.User, Text: req.Text}:
		case <-r.Context().Done():
			return
		}
		select {
		case reply := <-out:
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]string{"reply": reply.Text}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case <-r.Context().Done():
		}
	})
	srv := &http.Server{Addr: p.addr, Handler: mux}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	err := srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
