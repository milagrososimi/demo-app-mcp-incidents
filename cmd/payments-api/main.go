// Command payments-api serves payment lookups against a bounded database pool.
//
// Settings come from the environment; deploy/payments-api.yaml is what sets
// them in production.
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/milagrososimi/demo-app-mcp-incidents/internal/config"
	"github.com/milagrososimi/demo-app-mcp-incidents/internal/payments"
)

func main() {
	var (
		addr      = config.String("LISTEN_ADDR", ":8080")
		poolSize  = config.Int("DB_POOL_SIZE", 50)
		queryTime = config.Duration("DB_QUERY_DURATION", 20*time.Millisecond)
		waitTime  = config.Duration("DB_ACQUIRE_TIMEOUT", 2*time.Second)
	)

	pool := payments.NewPool(poolSize)
	slog.Info("starting payments-api", "addr", addr, "pool_size", pool.Size())

	mux := http.NewServeMux()
	mux.Handle("GET /payments/{id}", payments.NewHandler(pool, queryTime, waitTime))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		slog.Error("payments-api stopped", "error", err)
		os.Exit(1)
	}
}
