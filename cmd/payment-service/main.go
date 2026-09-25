// Command payment-service authorizes orders against the upstream processor.
//
// Settings come from the environment; deploy/payment-service.yaml is what sets
// them in production.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/AutopticAI/demo-app-mcp-incidents/internal/config"
	"github.com/AutopticAI/demo-app-mcp-incidents/internal/gateway"
)

// version is stamped at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	var (
		timeout  = config.Duration("UPSTREAM_TIMEOUT", 5*time.Second)
		retries  = config.Int("UPSTREAM_RETRIES", 2)
		upstream = config.String("UPSTREAM_URL", "")
		latency  = config.Duration("UPSTREAM_LATENCY", 3*time.Second)
		orderID  = config.String("ORDER_ID", "ord_1001")
	)

	// With no upstream configured, stand one up locally so the service can be
	// run on its own.
	if upstream == "" {
		stub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(latency)
			w.WriteHeader(http.StatusOK)
		}))
		defer stub.Close()
		upstream = stub.URL
		slog.Info("no UPSTREAM_URL set, using a local stub", "url", upstream, "latency", latency)
	}

	client := gateway.New(upstream, timeout, retries)
	slog.Info("starting payment-service", "version", version, "upstream", upstream)

	started := time.Now()
	err := client.Authorize(context.Background(), orderID)
	elapsed := time.Since(started).Round(time.Millisecond)

	if err != nil {
		slog.Error("authorization failed", "order", orderID, "elapsed", elapsed, "error", err)
		return
	}
	slog.Info("authorized", "order", orderID, "elapsed", elapsed)
}
