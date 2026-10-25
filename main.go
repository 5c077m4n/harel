package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/5c077m4n/harel/agent"
	"github.com/5c077m4n/harel/plugins/telegram"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	a := agent.New()
	a.Register(telegram.New())

	errs := make(chan error, 2)
	addr := os.Getenv("HAREL_HTTP_ADDR")
	if addr != "" {
		mux := http.NewServeMux()
		mux.HandleFunc("POST /message", func(w http.ResponseWriter, r *http.Request) {
			var req struct {
				Text string `json:"text"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			reply, err := a.Process(req.Text)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(map[string]string{"reply": reply}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		})
		srv := &http.Server{Addr: addr, Handler: mux}
		go func() {
			if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				errs <- err
			}
		}()
		go func() {
			<-ctx.Done()
			srv.Close()
		}()
	}

	errs <- a.Run(ctx)

	if err := <-errs; err != nil && !errors.Is(err, context.Canceled) {
		panic(err)
	}
}
